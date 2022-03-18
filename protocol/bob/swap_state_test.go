package bob

import (
	"context"
	"encoding/hex"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/net/message"
	pcommon "github.com/noot/atomic-swap/protocol"
	pswap "github.com/noot/atomic-swap/protocol/swap"
	"github.com/noot/atomic-swap/swapfactory"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/require"
)

var infofile = os.TempDir() + "/test.keys"

var (
	_            = logging.SetLogLevel("bob", "debug")
	testWallet   = "test-wallet"
	desiredAmout = common.EtherToWei(0.33)
)

type mockNet struct {
	msg net.Message
}

func (n *mockNet) SendSwapMessage(msg net.Message) error {
	n.msg = msg
	return nil
}

var (
	defaultTimeoutDuration, _ = time.ParseDuration("86400s") // 1 day = 60s * 60min * 24hr
	defaultContractSwapID     = big.NewInt(0)
)

func newTestBob(t *testing.T) *Instance {
	pk, err := ethcrypto.HexToECDSA(common.DefaultPrivKeyBob)
	require.NoError(t, err)

	ec, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)

	cfg := &Config{
		Ctx:                  context.Background(),
		Basepath:             "/tmp/bob",
		MoneroWalletEndpoint: common.DefaultBobMoneroEndpoint,
		MoneroDaemonEndpoint: common.DefaultMoneroDaemonEndpoint,
		WalletFile:           testWallet,
		WalletPassword:       "",
		EthereumClient:       ec,
		EthereumPrivateKey:   pk,
		Environment:          common.Development,
		ChainID:              big.NewInt(common.MainnetConfig.EthereumChainID),
		SwapManager:          pswap.NewManager(),
	}

	bob, err := NewInstance(cfg)
	require.NoError(t, err)

	bobAddr, err := bob.client.GetAddress(0)
	require.NoError(t, err)

	_ = bob.daemonClient.GenerateBlocks(bobAddr.Address, 256)
	err = bob.client.Refresh()
	require.NoError(t, err)
	return bob
}

func newTestInstance(t *testing.T) (*Instance, *swapState) {
	bob := newTestBob(t)
	swapState, err := newSwapState(bob, &types.Offer{}, nil, infofile, common.MoneroAmount(33), desiredAmout)
	require.NoError(t, err)
	return bob, swapState
}

func newTestAliceSendKeysMessage(t *testing.T) (*net.SendKeysMessage, *pcommon.KeysAndProof) {
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

func newSwap(t *testing.T, bob *Instance, swapState *swapState, claimKey, refundKey [32]byte, amount *big.Int,
	timeout time.Duration) (ethcommon.Address, *swapfactory.SwapFactory) {
	tm := big.NewInt(int64(timeout.Seconds()))
	if claimKey == [32]byte{} {
		claimKey = swapState.secp256k1Pub.Keccak256()
	}

	addr, _, contract, err := swapfactory.DeploySwapFactory(swapState.txOpts, bob.ethClient)
	require.NoError(t, err)

	swapState.txOpts.Value = amount
	defer func() {
		swapState.txOpts.Value = nil
	}()

	tx, err := contract.NewSwap(swapState.txOpts, claimKey, refundKey, bob.ethAddress, tm)
	require.NoError(t, err)

	receipt, err := bob.ethClient.TransactionReceipt(context.Background(), tx.Hash())
	require.NoError(t, err)
	require.Equal(t, 1, len(receipt.Logs))
	swapState.contractSwapID, err = swapfactory.GetIDFromLog(receipt.Logs[0])
	require.NoError(t, err)

	return addr, contract
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
	bob, swapState := newTestInstance(t)
	err := swapState.generateAndSetKeys()
	require.NoError(t, err)

	claimKey := swapState.secp256k1Pub.Keccak256()
	swapState.contractAddr, swapState.contract = newSwap(t, bob, swapState, claimKey,
		[32]byte{}, big.NewInt(33), defaultTimeoutDuration)

	_, err = swapState.contract.SetReady(swapState.txOpts, defaultContractSwapID)
	require.NoError(t, err)

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

	msg, aliceKeysAndProof := newTestAliceSendKeysMessage(t)
	alicePubKeys := aliceKeysAndProof.PublicKeyPair

	err = s.handleSendKeysMessage(msg)
	require.NoError(t, err)
	require.Equal(t, &message.NotifyETHLocked{}, s.nextExpectedMessage)
	require.Equal(t, alicePubKeys.SpendKey().Hex(), s.alicePublicKeys.SpendKey().Hex())
	require.Equal(t, alicePubKeys.ViewKey().Hex(), s.alicePublicKeys.ViewKey().Hex())
	require.True(t, s.info.Status().IsOngoing())
}

func TestSwapState_HandleProtocolMessage_NotifyETHLocked_ok(t *testing.T) {
	bob, s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.NotifyETHLocked{}
	err := s.generateAndSetKeys()
	require.NoError(t, err)

	aliceKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	s.setAlicePublicKeys(aliceKeysAndProof.PublicKeyPair, aliceKeysAndProof.Secp256k1PublicKey)

	msg := &message.NotifyETHLocked{}
	resp, done, err := s.HandleProtocolMessage(msg)
	require.Equal(t, errMissingAddress, err)
	require.Nil(t, resp)
	require.True(t, done)

	duration, err := time.ParseDuration("2s")
	require.NoError(t, err)
	addr, _ := newSwap(t, bob, s, s.secp256k1Pub.Keccak256(), s.aliceSecp256K1PublicKey.Keccak256(),
		desiredAmout.BigInt(), duration)

	msg = &message.NotifyETHLocked{
		Address:        addr.String(),
		ContractSwapID: defaultContractSwapID,
	}

	resp, done, err = s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyXMRLockType, resp.Type())
	require.False(t, done)
	require.NotNil(t, s.contract)
	require.Equal(t, addr, s.contractAddr)
	require.Equal(t, duration, s.t1.Sub(s.t0))
	require.Equal(t, &message.NotifyReady{}, s.nextExpectedMessage)
	require.True(t, s.info.Status().IsOngoing())
}

func TestSwapState_HandleProtocolMessage_NotifyETHLocked_timeout(t *testing.T) {
	bob, s := newTestInstance(t)
	defer s.cancel()
	s.bob.net = new(mockNet)
	s.nextExpectedMessage = &message.NotifyETHLocked{}
	err := s.generateAndSetKeys()
	require.NoError(t, err)

	aliceKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	s.setAlicePublicKeys(aliceKeysAndProof.PublicKeyPair, aliceKeysAndProof.Secp256k1PublicKey)

	msg := &message.NotifyETHLocked{}
	resp, done, err := s.HandleProtocolMessage(msg)
	require.Equal(t, errMissingAddress, err)
	require.Nil(t, resp)
	require.True(t, done)

	duration, err := time.ParseDuration("15s")
	require.NoError(t, err)
	addr, _ := newSwap(t, bob, s, s.secp256k1Pub.Keccak256(), s.aliceSecp256K1PublicKey.Keccak256(),
		desiredAmout.BigInt(), duration)

	msg = &message.NotifyETHLocked{
		Address:        addr.String(),
		ContractSwapID: defaultContractSwapID,
	}

	resp, done, err = s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyXMRLockType, resp.Type())
	require.False(t, done)
	require.NotNil(t, s.contract)
	require.Equal(t, addr, s.contractAddr)
	require.Equal(t, duration, s.t1.Sub(s.t0))
	require.Equal(t, &message.NotifyReady{}, s.nextExpectedMessage)

	for status := range s.statusCh {
		if status == types.CompletedSuccess {
			break
		} else if !status.IsOngoing() {
			t.Fatalf("got wrong exit status %s, expected CompletedSuccess", status)
		}
	}

	require.NotNil(t, s.bob.net.(*mockNet).msg)
	require.Equal(t, types.CompletedSuccess, s.info.Status())
}

func TestSwapState_HandleProtocolMessage_NotifyReady(t *testing.T) {
	bob, s := newTestInstance(t)

	s.nextExpectedMessage = &message.NotifyReady{}
	err := s.generateAndSetKeys()
	require.NoError(t, err)

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)
	_, s.contract = newSwap(t, bob, s, [32]byte{}, [32]byte{}, desiredAmout.BigInt(), duration)

	_, err = s.contract.SetReady(s.txOpts, defaultContractSwapID)
	require.NoError(t, err)

	msg := &message.NotifyReady{}

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.True(t, done)
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyClaimedType, resp.Type())
	require.Equal(t, types.CompletedSuccess, s.info.Status())
}

func TestSwapState_handleRefund(t *testing.T) {
	bob, s := newTestInstance(t)

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	aliceKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	s.setAlicePublicKeys(aliceKeysAndProof.PublicKeyPair, aliceKeysAndProof.Secp256k1PublicKey)

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)

	refundKey := aliceKeysAndProof.Secp256k1PublicKey.Keccak256()
	_, s.contract = newSwap(t, bob, s, [32]byte{}, refundKey, desiredAmout.BigInt(), duration)

	// lock XMR
	addrAB, err := s.lockFunds(common.MoneroToPiconero(s.info.ProvidedAmount()))
	require.NoError(t, err)

	// call refund w/ Alice's spend key
	secret := aliceKeysAndProof.PrivateKeyPair.SpendKeyBytes()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret))

	tx, err := s.contract.Refund(s.txOpts, defaultContractSwapID, sc)
	require.NoError(t, err)

	addr, err := s.handleRefund(tx.Hash().String())
	require.NoError(t, err)
	require.Equal(t, addrAB, addr)
}

func TestSwapState_HandleProtocolMessage_NotifyRefund(t *testing.T) {
	bob, s := newTestInstance(t)

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	aliceKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	s.setAlicePublicKeys(aliceKeysAndProof.PublicKeyPair, aliceKeysAndProof.Secp256k1PublicKey)

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)

	refundKey := aliceKeysAndProof.Secp256k1PublicKey.Keccak256()
	_, s.contract = newSwap(t, bob, s, [32]byte{}, refundKey, desiredAmout.BigInt(), duration)

	// lock XMR
	_, err = s.lockFunds(common.MoneroToPiconero(s.info.ProvidedAmount()))
	require.NoError(t, err)

	// call refund w/ Alice's secret
	secret := aliceKeysAndProof.DLEqProof.Secret()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret[:]))

	tx, err := s.contract.Refund(s.txOpts, defaultContractSwapID, sc)
	require.NoError(t, err)

	msg := &message.NotifyRefund{
		TxHash: tx.Hash().String(),
	}

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.True(t, done)
	require.Nil(t, resp)
	require.Equal(t, types.CompletedRefund, s.info.Status())
}

// test that if the protocol exits early, and Alice refunds, Bob can reclaim his monero
func TestSwapState_Exit_Reclaim(t *testing.T) {
	bob, s := newTestInstance(t)

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	aliceKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	s.setAlicePublicKeys(aliceKeysAndProof.PublicKeyPair, aliceKeysAndProof.Secp256k1PublicKey)

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)

	refundKey := aliceKeysAndProof.Secp256k1PublicKey.Keccak256()
	s.contractAddr, s.contract = newSwap(t, bob, s, [32]byte{}, refundKey, desiredAmout.BigInt(), duration)

	// lock XMR
	_, err = s.lockFunds(common.MoneroToPiconero(s.info.ProvidedAmount()))
	require.NoError(t, err)

	// call refund w/ Alice's secret
	secret := aliceKeysAndProof.DLEqProof.Secret()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret[:]))

	tx, err := s.contract.Refund(s.txOpts, defaultContractSwapID, sc)
	require.NoError(t, err)

	receipt, err := bob.ethClient.TransactionReceipt(s.ctx, tx.Hash())
	require.NoError(t, err)
	require.Equal(t, 1, len(receipt.Logs))
	require.Equal(t, 1, len(receipt.Logs[0].Topics))
	require.Equal(t, refundedTopic, receipt.Logs[0].Topics[0])

	s.nextExpectedMessage = &message.NotifyReady{}
	err = s.Exit()
	require.NoError(t, err)

	balance, err := bob.client.GetBalance(0)
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

	s.nextExpectedMessage = &message.NotifyETHLocked{}
	err = s.Exit()
	require.NoError(t, err)
	require.Equal(t, types.CompletedAbort, s.info.Status())

	s.nextExpectedMessage = nil
	err = s.Exit()
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
