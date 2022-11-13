package xmrmaker

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/net/message"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	pswap "github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/xmrmaker/offers"
	"github.com/athanorlabs/atomic-swap/tests"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/require"
)

var (
	_                         = logging.SetLogLevel("xmrmaker", "debug")
	testWallet                = "test-wallet"
	desiredAmount             = common.EtherToWei(0.33)
	defaultTimeoutDuration, _ = time.ParseDuration("86400s") // 1 day = 60s * 60min * 24hr
)

type mockNet struct {
	msgMu sync.Mutex  // lock needed, as SendSwapMessage is called async from timeout handlers
	msg   net.Message // last value passed to SendSwapMessage
}

func (n *mockNet) LastSentMessage() net.Message {
	n.msgMu.Lock()
	defer n.msgMu.Unlock()
	return n.msg
}

func (n *mockNet) SendSwapMessage(msg net.Message, _ types.Hash) error {
	n.msgMu.Lock()
	defer n.msgMu.Unlock()
	n.msg = msg
	return nil
}

func (n *mockNet) CloseProtocolStream(_ types.Hash) {}

func newSwapManager(t *testing.T) pswap.Manager {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := pswap.NewMockDatabase(ctrl)
	db.EXPECT().GetAllSwaps()
	db.EXPECT().PutSwap(gomock.Any()).AnyTimes()

	sm, err := pswap.NewManager(db)
	require.NoError(t, err)
	return sm
}

func newTestXMRMakerAndDB(t *testing.T) (*Instance, *offers.MockDatabase) {
	pk := tests.GetMakerTestKey(t)
	ec, chainID := tests.NewEthClient(t)

	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, chainID)
	require.NoError(t, err)

	var forwarderAddress ethcommon.Address
	_, tx, contract, err := contracts.DeploySwapFactory(txOpts, ec, forwarderAddress)
	require.NoError(t, err)

	addr, err := bind.WaitDeployed(context.Background(), ec, tx)
	require.NoError(t, err)

	bcfg := &backend.Config{
		Ctx:                 context.Background(),
		MoneroClient:        monero.CreateWalletClient(t),
		EthereumClient:      ec,
		EthereumPrivateKey:  pk,
		Environment:         common.Development,
		SwapContract:        contract,
		SwapContractAddress: addr,
		SwapManager:         newSwapManager(t),
		Net:                 new(mockNet),
	}

	b, err := backend.NewBackend(bcfg)
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := offers.NewMockDatabase(ctrl)
	db.EXPECT().GetAllOffers()

	net := NewMockHost(ctrl)

	cfg := &Config{
		Backend:        b,
		DataDir:        path.Join(t.TempDir(), "xmrmaker"),
		WalletFile:     testWallet,
		WalletPassword: "",
		Database:       db,
		Network:        net,
	}

	xmrmaker, err := NewInstance(cfg)
	require.NoError(t, err)

	monero.MineMinXMRBalance(t, b, 5.0)
	err = b.Refresh()
	require.NoError(t, err)
	return xmrmaker, db
}

func newTestInstanceAndDB(t *testing.T) (*Instance, *swapState, *offers.MockDatabase) {
	xmrmaker, db := newTestXMRMakerAndDB(t)
	// infoFile := path.Join(t.TempDir(), "test.keys")
	// oe := &types.OfferExtra{
	// 	InfoFile: infoFile,
	// }

	swapState, err := newSwapState(
		xmrmaker.backend,
		types.NewOffer("", 0, 0, 0, types.EthAssetETH),
		&types.OfferExtra{},
		xmrmaker.offerManager,
		common.MoneroAmount(33),
		desiredAmount,
	)
	require.NoError(t, err)
	return xmrmaker, swapState, db
}

func newTestInstance(t *testing.T) (*Instance, *swapState) {
	xmrmaker, swapState, _ := newTestInstanceAndDB(t)
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

func newSwap(t *testing.T, ss *swapState, claimKey, refundKey types.Hash, amount *big.Int,
	timeout time.Duration) ethcommon.Hash {
	tm := big.NewInt(int64(timeout.Seconds()))
	if types.IsHashZero(claimKey) {
		claimKey = ss.secp256k1Pub.Keccak256()
	}

	txOpts, err := ss.TxOpts()
	require.NoError(t, err)

	// TODO: this is sus, update this when signing interfaces are updated
	txOpts.Value = amount

	ethAddr := ss.EthAddress()
	nonce := big.NewInt(0)
	asset := types.EthAssetETH
	tx, err := ss.Contract().NewSwap(txOpts, claimKey, refundKey, ethAddr, tm,
		ethcommon.Address(asset), amount, nonce)
	require.NoError(t, err)
	receipt := tests.MineTransaction(t, ss, tx)

	require.Equal(t, 1, len(receipt.Logs))
	ss.contractSwapID, err = contracts.GetIDFromLog(receipt.Logs[0])
	require.NoError(t, err)

	t0, t1, err := contracts.GetTimeoutsFromLog(receipt.Logs[0])
	require.NoError(t, err)

	ss.contractSwap = contracts.SwapFactorySwap{
		Owner:        ethAddr,
		Claimer:      ethAddr,
		PubKeyClaim:  claimKey,
		PubKeyRefund: refundKey,
		Timeout0:     t0,
		Timeout1:     t1,
		Asset:        ethcommon.Address(asset),
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
	require.True(t, swapState.info.Status.IsOngoing())
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
	require.Equal(t, EventETHLockedType, s.nextExpectedEvent)
	require.Equal(t, xmrtakerPubKeys.SpendKey().Hex(), s.xmrtakerPublicKeys.SpendKey().Hex())
	require.Equal(t, xmrtakerPubKeys.ViewKey().Hex(), s.xmrtakerPublicKeys.ViewKey().Hex())
	require.True(t, s.info.Status.IsOngoing())
}

func TestSwapState_HandleProtocolMessage_NotifyETHLocked_ok(t *testing.T) {
	_, s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedEvent = EventETHLockedType
	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrtakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	s.setXMRTakerPublicKeys(xmrtakerKeysAndProof.PublicKeyPair, xmrtakerKeysAndProof.Secp256k1PublicKey)

	msg := &message.NotifyETHLocked{}
	err = s.HandleProtocolMessage(msg)
	require.True(t, errors.Is(err, errMissingAddress))

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

	err = s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	resp := s.Net().(*mockNet).LastSentMessage()
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyXMRLockType, resp.Type())
	require.Equal(t, duration, s.t1.Sub(s.t0))
	require.Equal(t, EventContractReadyType, s.nextExpectedEvent)
	require.True(t, s.info.Status.IsOngoing())
}

func TestSwapState_HandleProtocolMessage_NotifyETHLocked_timeout(t *testing.T) {
	_, s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedEvent = EventETHLockedType
	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrtakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	s.setXMRTakerPublicKeys(xmrtakerKeysAndProof.PublicKeyPair, xmrtakerKeysAndProof.Secp256k1PublicKey)

	msg := &message.NotifyETHLocked{}
	err = s.HandleProtocolMessage(msg)
	require.True(t, errors.Is(err, errMissingAddress))

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

	err = s.HandleProtocolMessage(msg)
	require.NoError(t, err)

	resp := s.Net().(*mockNet).LastSentMessage()
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyXMRLockType, resp.Type())
	require.Equal(t, duration, s.t1.Sub(s.t0))
	require.Equal(t, EventContractReadyType, s.nextExpectedEvent)

	for status := range s.offerExtra.StatusCh {
		if status == types.CompletedSuccess {
			break
		} else if !status.IsOngoing() {
			t.Fatalf("got wrong exit status %s, expected CompletedSuccess", status)
		}
	}

	require.Equal(t, types.CompletedSuccess, s.info.Status)
}

func TestSwapState_handleRefund(t *testing.T) {
	_, s, db := newTestInstanceAndDB(t)
	db.EXPECT().PutOffer(s.offer)

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
	_, err = s.lockFunds(common.MoneroToPiconero(s.info.ProvidedAmount))
	require.NoError(t, err)

	// call refund w/ XMRTaker's spend key
	secret := xmrtakerKeysAndProof.PrivateKeyPair.SpendKeyBytes()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret))

	txOpts, err := s.TxOpts()
	require.NoError(t, err)
	tx, err := s.Contract().Refund(txOpts, s.contractSwap, sc)
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(s.Backend.Ctx(), s.EthClient(), tx.Hash())
	require.NoError(t, err)
	require.Equal(t, 1, len(receipt.Logs))

	// runContractEventWatcher will trigger EventETHRefunded,
	// which will then set the next expected event to EventExit.
	for status := range s.info.StatusCh() {
		if !status.IsOngoing() {
			break
		}
	}

	require.Equal(t, types.CompletedRefund, s.info.Status)
}

// test that if the protocol exits early, and XMRTaker refunds, XMRMaker can reclaim his monero
func TestSwapState_Exit_Reclaim(t *testing.T) {
	_, s, db := newTestInstanceAndDB(t)
	db.EXPECT().PutOffer(s.offer)

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
	_, err = s.lockFunds(common.MoneroToPiconero(s.info.ProvidedAmount))
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

	s.nextExpectedEvent = EventContractReadyType
	err = s.Exit()
	require.NoError(t, err)

	balance, err := s.GetBalance(0)
	require.NoError(t, err)
	require.Equal(t, common.MoneroToPiconero(s.info.ProvidedAmount).Uint64(), balance.Balance)
	require.Equal(t, types.CompletedRefund, s.info.Status)
}

func TestSwapState_Exit_Aborted(t *testing.T) {
	_, s, db := newTestInstanceAndDB(t)
	db.EXPECT().PutOffer(s.offer)

	s.nextExpectedEvent = EventETHLockedType
	err := s.Exit()
	require.NoError(t, err)
	require.Equal(t, types.CompletedAbort, s.info.Status)
}

func TestSwapState_Exit_Aborted_1(t *testing.T) {
	_, s, db := newTestInstanceAndDB(t)
	db.EXPECT().PutOffer(s.offer)

	s.nextExpectedEvent = EventETHRefundedType
	err := s.Exit()
	require.True(t, errors.Is(err, errUnexpectedMessageType))
	require.Equal(t, types.CompletedAbort, s.info.Status)
}

func TestSwapState_Exit_Success(t *testing.T) {
	b, s := newTestInstance(t)
	s.offer = types.NewOffer(types.ProvidesXMR, 0.1, 0.2, 0.1, types.EthAssetETH)
	s.info.SetStatus(types.CompletedSuccess)
	err := s.Exit()
	require.NoError(t, err)

	// since the swap was successful, the offer should be removed.
	o, oe, _ := b.offerManager.GetOffer(s.offer.ID)
	require.Nil(t, o)
	require.Nil(t, oe)
}

func TestSwapState_Exit_Refunded(t *testing.T) {
	b, s, db := newTestInstanceAndDB(t)

	b.net.(*MockHost).EXPECT().Advertise()

	s.offer = types.NewOffer(types.ProvidesXMR, 0.1, 0.2, 0.1, types.EthAssetETH)
	db.EXPECT().PutOffer(s.offer)
	b.MakeOffer(s.offer, "", 0)

	s.info.SetStatus(types.CompletedRefund)
	err := s.Exit()
	require.NoError(t, err)

	// since the swap was not successful, the offer should be re-added to the offer manager.
	o, oe, err := b.offerManager.GetOffer(s.offer.ID)
	require.NoError(t, err)
	require.NotNil(t, o)
	require.NotNil(t, oe)
}
