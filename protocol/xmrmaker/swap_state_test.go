package xmrmaker

import (
	"context"
	"encoding/hex"
	"math/big"
	"path"
	"testing"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/net/message"
	pcommon "github.com/noot/atomic-swap/protocol"
	"github.com/noot/atomic-swap/protocol/backend"
	pswap "github.com/noot/atomic-swap/protocol/swap"
	"github.com/noot/atomic-swap/swapfactory"
	"github.com/noot/atomic-swap/tests"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/require"
)

var (
	_             = logging.SetLogLevel("xmrmaker", "debug")
	testWallet    = "test-wallet"
	desiredAmount = common.EtherToWei(0.33)
)

type mockNet struct {
	msg net.Message
}

func (n *mockNet) SendSwapMessage(msg net.Message, _ types.Hash) error {
	n.msg = msg
	return nil
}

var (
	defaultTimeoutDuration, _ = time.ParseDuration("86400s") // 1 day = 60s * 60min * 24hr
)

func newTestXMRMaker(t *testing.T) *Instance {
	pk := tests.GetMakerTestKey(t)
	ec, chainID := tests.NewEthClient(t)

	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, chainID)
	require.NoError(t, err)

	_, tx, contract, err := swapfactory.DeploySwapFactory(txOpts, ec)
	require.NoError(t, err)

	addr, err := bind.WaitDeployed(context.Background(), ec, tx)
	require.NoError(t, err)

	bcfg := &backend.Config{
		Ctx:                  context.Background(),
		MoneroWalletEndpoint: tests.CreateWalletRPCService(t),
		MoneroDaemonEndpoint: common.DefaultMoneroDaemonEndpoint,
		EthereumClient:       ec,
		EthereumPrivateKey:   pk,
		Environment:          common.Development,
		ChainID:              chainID,
		SwapContract:         contract,
		SwapContractAddress:  addr,
		SwapManager:          pswap.NewManager(),
		Net:                  new(mockNet),
	}

	b, err := backend.NewBackend(bcfg)
	require.NoError(t, err)

	cfg := &Config{
		Backend:        b,
		Basepath:       path.Join(t.TempDir(), "xmrmaker"),
		WalletFile:     testWallet,
		WalletPassword: "",
	}

	// NewInstance(..) below expects a pre-existing wallet, so create it
	err = monero.NewClient(bcfg.MoneroWalletEndpoint).CreateWallet(cfg.WalletFile, "")
	require.NoError(t, err)

	xmrmaker, err := NewInstance(cfg)
	require.NoError(t, err)

	xmrmakerAddr, err := b.GetAddress(0)
	require.NoError(t, err)

	_ = b.GenerateBlocks(xmrmakerAddr.Address, 512)
	err = b.Refresh()
	require.NoError(t, err)
	return xmrmaker
}

func newTestInstance(t *testing.T) (*Instance, *swapState) {
	xmrmaker := newTestXMRMaker(t)
	infoFile := path.Join(t.TempDir(), "test.keys")
	swapState, err := newSwapState(xmrmaker.backend, &types.Offer{}, xmrmaker.offerManager, nil, infoFile,
		common.MoneroAmount(33), desiredAmount)
	require.NoError(t, err)
	swapState.SetContract(xmrmaker.backend.Contract())
	swapState.SetContractAddress(xmrmaker.backend.ContractAddr())
	return xmrmaker, swapState
}

func newTestXMRTakerSendKeysMessage(t *testing.T) (*net.SendKeysMessage, *pcommon.KeysAndProof) {
	keysAndProof, err := pcommon.GenerateKeysAndProof()
	require.NoError(t, err)

	msg := &net.SendKeysMessage{
		PublicSpendKey:     keysAndProof.PublicKeyPair.SpendKey().Hex(),
		PublicViewKey:      keysAndProof.PublicKeyPair.ViewKey().Hex(),
		DLEqProof:          hex.EncodeToString(keysAndProof.DLEqProof.Proof()),
		Secp256k1PublicKey: keysAndProof.Secp256k1PublicKey.String(),
	}

	return msg, keysAndProof
}

func newSwap(t *testing.T, ss *swapState, claimKey, refundKey [32]byte, amount *big.Int,
	timeout time.Duration) ethcommon.Hash {
	tm := big.NewInt(int64(timeout.Seconds()))
	if claimKey == [32]byte{} {
		claimKey = ss.secp256k1Pub.Keccak256()
	}

	txOpts, err := ss.TxOpts()
	require.NoError(t, err)

	// TODO: this is sus, update this when signing interfaces are updated
	txOpts.Value = amount

	ethAddr := ss.EthAddress()
	nonce := big.NewInt(0)
	tx, err := ss.Contract().NewSwap(txOpts, claimKey, refundKey, ethAddr, tm, nonce)
	require.NoError(t, err)
	receipt := tests.MineTransaction(t, ss, tx)

	require.Equal(t, 1, len(receipt.Logs))
	ss.contractSwapID, err = swapfactory.GetIDFromLog(receipt.Logs[0])
	require.NoError(t, err)

	t0, t1, err := swapfactory.GetTimeoutsFromLog(receipt.Logs[0])
	require.NoError(t, err)

	ss.contractSwap = swapfactory.SwapFactorySwap{
		Owner:        ethAddr,
		Claimer:      ethAddr,
		PubKeyClaim:  claimKey,
		PubKeyRefund: refundKey,
		Timeout0:     t0,
		Timeout1:     t1,
		Value:        amount,
		Nonce:        nonce,
	}

	ss.setTimeouts(t0, t1)
	return tx.Hash()
}

func TestSwapState_GenerateAndSetKeys(t *testing.T) {
	_, swapState := newTestInstance(t)

	err := swapState.generateAndSetKeys()
	require.NoError(t, err)
	require.NotNil(t, swapState.privkeys)
	require.NotNil(t, swapState.pubkeys)
	require.NotNil(t, swapState.dleqProof)
}

func TestSwapState_ClaimFunds(t *testing.T) {
	_, swapState := newTestInstance(t)
	err := swapState.generateAndSetKeys()
	require.NoError(t, err)

	claimKey := swapState.secp256k1Pub.Keccak256()
	newSwap(t, swapState, claimKey,
		[32]byte{}, big.NewInt(33), defaultTimeoutDuration)

	txOpts, err := swapState.TxOpts()
	require.NoError(t, err)
	tx, err := swapState.Contract().SetReady(txOpts, swapState.contractSwap)
	require.NoError(t, err)
	tests.MineTransaction(t, swapState, tx)

	txHash, err := swapState.claimFunds()
	require.NoError(t, err)
	require.NotEqual(t, "", txHash)
	require.True(t, swapState.info.Status().IsOngoing())
}

func TestSwapState_handleSendKeysMessage(t *testing.T) {
	_, s := newTestInstance(t)

	msg := &net.SendKeysMessage{}
	err := s.handleSendKeysMessage(msg)
	require.Equal(t, errMissingKeys, err)

	msg, xmrtakerKeysAndProof := newTestXMRTakerSendKeysMessage(t)
	xmrtakerPubKeys := xmrtakerKeysAndProof.PublicKeyPair

	err = s.handleSendKeysMessage(msg)
	require.NoError(t, err)
	require.Equal(t, &message.NotifyETHLocked{}, s.nextExpectedMessage)
	require.Equal(t, xmrtakerPubKeys.SpendKey().Hex(), s.xmrtakerPublicKeys.SpendKey().Hex())
	require.Equal(t, xmrtakerPubKeys.ViewKey().Hex(), s.xmrtakerPublicKeys.ViewKey().Hex())
	require.True(t, s.info.Status().IsOngoing())
}

func TestSwapState_HandleProtocolMessage_NotifyETHLocked_ok(t *testing.T) {
	_, s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.NotifyETHLocked{}
	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrtakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	s.setXMRTakerPublicKeys(xmrtakerKeysAndProof.PublicKeyPair, xmrtakerKeysAndProof.Secp256k1PublicKey)

	msg := &message.NotifyETHLocked{}
	resp, done, err := s.HandleProtocolMessage(msg)
	require.Equal(t, errMissingAddress, err)
	require.Nil(t, resp)
	require.True(t, done)

	duration, err := time.ParseDuration("2s")
	require.NoError(t, err)
	hash := newSwap(t, s, s.secp256k1Pub.Keccak256(), s.xmrtakerSecp256K1PublicKey.Keccak256(),
		desiredAmount.BigInt(), duration)
	addr := s.ContractAddr()

	msg = &message.NotifyETHLocked{
		Address:        addr.String(),
		ContractSwapID: s.contractSwapID,
		TxHash:         hash.String(),
		ContractSwap:   pcommon.ConvertContractSwapToMsg(s.contractSwap),
	}

	resp, done, err = s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyXMRLockType, resp.Type())
	require.False(t, done)
	require.Equal(t, duration, s.t1.Sub(s.t0))
	require.Equal(t, &message.NotifyReady{}, s.nextExpectedMessage)
	require.True(t, s.info.Status().IsOngoing())
}

func TestSwapState_HandleProtocolMessage_NotifyETHLocked_timeout(t *testing.T) {
	if testing.Short() {
		t.Skip() // TODO: times out on CI with error
		// "xmrmaker/swap_state.go:227	failed to claim funds: err=no contract code at given address"
	}

	_, s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.NotifyETHLocked{}
	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrtakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	s.setXMRTakerPublicKeys(xmrtakerKeysAndProof.PublicKeyPair, xmrtakerKeysAndProof.Secp256k1PublicKey)

	msg := &message.NotifyETHLocked{}
	resp, done, err := s.HandleProtocolMessage(msg)
	require.Equal(t, errMissingAddress, err)
	require.Nil(t, resp)
	require.True(t, done)

	duration, err := time.ParseDuration("15s")
	require.NoError(t, err)
	hash := newSwap(t, s, s.secp256k1Pub.Keccak256(), s.xmrtakerSecp256K1PublicKey.Keccak256(),
		desiredAmount.BigInt(), duration)
	addr := s.ContractAddr()

	msg = &message.NotifyETHLocked{
		Address:        addr.String(),
		ContractSwapID: s.contractSwapID,
		TxHash:         hash.String(),
		ContractSwap:   pcommon.ConvertContractSwapToMsg(s.contractSwap),
	}

	resp, done, err = s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyXMRLockType, resp.Type())
	require.False(t, done)
	require.Equal(t, duration, s.t1.Sub(s.t0))
	require.Equal(t, &message.NotifyReady{}, s.nextExpectedMessage)

	for status := range s.statusCh {
		if status == types.CompletedSuccess {
			break
		} else if !status.IsOngoing() {
			t.Fatalf("got wrong exit status %s, expected CompletedSuccess", status)
		}
	}

	require.NotNil(t, s.Net().(*mockNet).msg)
	require.Equal(t, types.CompletedSuccess, s.info.Status())
}

func TestSwapState_HandleProtocolMessage_NotifyReady(t *testing.T) {
	_, s := newTestInstance(t)

	s.nextExpectedMessage = &message.NotifyReady{}
	err := s.generateAndSetKeys()
	require.NoError(t, err)

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)
	newSwap(t, s, [32]byte{}, [32]byte{}, desiredAmount.BigInt(), duration)

	txOpts, err := s.TxOpts()
	require.NoError(t, err)
	tx, err := s.Contract().SetReady(txOpts, s.contractSwap)
	require.NoError(t, err)
	tests.MineTransaction(t, s, tx)

	msg := &message.NotifyReady{}

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.True(t, done)
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyClaimedType, resp.Type())
	require.Equal(t, types.CompletedSuccess, s.info.Status())
}

func TestSwapState_handleRefund(t *testing.T) {
	_, s := newTestInstance(t)

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrtakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	s.setXMRTakerPublicKeys(xmrtakerKeysAndProof.PublicKeyPair, xmrtakerKeysAndProof.Secp256k1PublicKey)

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)

	refundKey := xmrtakerKeysAndProof.Secp256k1PublicKey.Keccak256()
	newSwap(t, s, [32]byte{}, refundKey, desiredAmount.BigInt(), duration)

	// lock XMR
	addrAB, err := s.lockFunds(common.MoneroToPiconero(s.info.ProvidedAmount()))
	require.NoError(t, err)

	// call refund w/ XMRTaker's spend key
	secret := xmrtakerKeysAndProof.PrivateKeyPair.SpendKeyBytes()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret))

	txOpts, err := s.TxOpts()
	require.NoError(t, err)
	tx, err := s.Contract().Refund(txOpts, s.contractSwap, sc)
	require.NoError(t, err)
	tests.MineTransaction(t, s, tx)

	addr, err := s.handleRefund(tx.Hash().String())
	require.NoError(t, err)
	require.Equal(t, addrAB, addr)
}

func TestSwapState_HandleProtocolMessage_NotifyRefund(t *testing.T) {
	_, s := newTestInstance(t)

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrtakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	s.setXMRTakerPublicKeys(xmrtakerKeysAndProof.PublicKeyPair, xmrtakerKeysAndProof.Secp256k1PublicKey)

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)

	refundKey := xmrtakerKeysAndProof.Secp256k1PublicKey.Keccak256()
	newSwap(t, s, [32]byte{}, refundKey, desiredAmount.BigInt(), duration)

	// lock XMR
	_, err = s.lockFunds(common.MoneroToPiconero(s.info.ProvidedAmount()))
	require.NoError(t, err)

	// call refund w/ XMRTaker's secret
	secret := xmrtakerKeysAndProof.DLEqProof.Secret()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret[:]))

	txOpts, err := s.TxOpts()
	require.NoError(t, err)
	tx, err := s.Contract().Refund(txOpts, s.contractSwap, sc)
	require.NoError(t, err)
	tests.MineTransaction(t, s, tx)

	msg := &message.NotifyRefund{
		TxHash: tx.Hash().String(),
	}

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.True(t, done)
	require.Nil(t, resp)
	require.Equal(t, types.CompletedRefund, s.info.Status())
}

// test that if the protocol exits early, and XMRTaker refunds, XMRMaker can reclaim his monero
func TestSwapState_Exit_Reclaim(t *testing.T) {
	_, s := newTestInstance(t)

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrtakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	s.setXMRTakerPublicKeys(xmrtakerKeysAndProof.PublicKeyPair, xmrtakerKeysAndProof.Secp256k1PublicKey)

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)

	refundKey := xmrtakerKeysAndProof.Secp256k1PublicKey.Keccak256()
	newSwap(t, s, [32]byte{}, refundKey, desiredAmount.BigInt(), duration)

	// lock XMR
	_, err = s.lockFunds(common.MoneroToPiconero(s.info.ProvidedAmount()))
	require.NoError(t, err)

	// call refund w/ XMRTaker's secret
	secret := xmrtakerKeysAndProof.DLEqProof.Secret()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret[:]))

	txOpts, err := s.TxOpts()
	require.NoError(t, err)
	tx, err := s.Contract().Refund(txOpts, s.contractSwap, sc)
	require.NoError(t, err)
	receipt := tests.MineTransaction(t, s, tx)

	require.Equal(t, 1, len(receipt.Logs))
	require.Equal(t, 1, len(receipt.Logs[0].Topics))
	require.Equal(t, refundedTopic, receipt.Logs[0].Topics[0])

	s.nextExpectedMessage = &message.NotifyReady{}
	err = s.Exit()
	require.NoError(t, err)

	balance, err := s.GetBalance(0)
	require.NoError(t, err)
	require.Equal(t, common.MoneroToPiconero(s.info.ProvidedAmount()).Uint64(), uint64(balance.Balance))
	require.Equal(t, types.CompletedRefund, s.info.Status())
}

func TestSwapState_Exit_Aborted(t *testing.T) {
	_, s := newTestInstance(t)
	s.nextExpectedMessage = &message.SendKeysMessage{}
	err := s.Exit()
	require.NoError(t, err)
	require.Equal(t, types.CompletedAbort, s.info.Status())
}

func TestSwapState_Exit_Aborted_1(t *testing.T) {
	_, s := newTestInstance(t)
	s.nextExpectedMessage = &message.NotifyETHLocked{}
	err := s.Exit()
	require.NoError(t, err)
	require.Equal(t, types.CompletedAbort, s.info.Status())
}

func TestSwapState_Exit_Aborted_2(t *testing.T) {
	_, s := newTestInstance(t)
	s.nextExpectedMessage = nil
	err := s.Exit()
	require.Equal(t, errUnexpectedMessageType, err)
	require.Equal(t, types.CompletedAbort, s.info.Status())
}

func TestSwapState_Exit_Success(t *testing.T) {
	b, s := newTestInstance(t)
	s.offer = &types.Offer{
		Provides:      types.ProvidesXMR,
		MinimumAmount: 0.1,
		MaximumAmount: 0.2,
		ExchangeRate:  0.1,
	}

	s.info.SetStatus(types.CompletedSuccess)
	err := s.Exit()
	require.NoError(t, err)
	require.Nil(t, b.offerManager.offers[s.offer.GetID()])
}

func TestSwapState_Exit_Refunded(t *testing.T) {
	b, s := newTestInstance(t)
	s.offer = &types.Offer{
		Provides:      types.ProvidesXMR,
		MinimumAmount: 0.1,
		MaximumAmount: 0.2,
		ExchangeRate:  0.1,
	}
	b.MakeOffer(s.offer)

	s.info.SetStatus(types.CompletedRefund)
	err := s.Exit()
	require.NoError(t, err)
	require.NotNil(t, b.offerManager.offers[s.offer.GetID()])
}
