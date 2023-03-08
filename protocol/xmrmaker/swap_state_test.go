package xmrmaker

import (
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/cockroachdb/apd/v3"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/net/message"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
	"github.com/athanorlabs/atomic-swap/protocol/xmrmaker/offers"
	"github.com/athanorlabs/atomic-swap/tests"

	ethcommon "github.com/ethereum/go-ethereum/common"
	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/require"
)

var (
	_                         = logging.SetLogLevel("protocol", "debug")
	_                         = logging.SetLogLevel("xmrmaker", "debug")
	desiredAmount             = coins.EtherToWei(apd.New(33, -2)) // "0.33"
	defaultTimeoutDuration, _ = time.ParseDuration("86400s")      // 1 day = 60s * 60min * 24hr
)

func newTestSwapStateAndDB(t *testing.T) (*Instance, *swapState, *offers.MockDatabase) {
	xmrmaker, db := newTestInstanceAndDB(t)

	swapState, err := newSwapStateFromStart(
		xmrmaker.backend,
		types.NewOffer("", new(apd.Decimal), new(apd.Decimal), new(coins.ExchangeRate), types.EthAssetETH),
		&types.OfferExtra{},
		xmrmaker.offerManager,
		coins.MoneroToPiconero(coins.StrToDecimal("0.05")),
		desiredAmount,
	)
	require.NoError(t, err)
	return xmrmaker, swapState, db
}

func newTestSwapStateAndNet(t *testing.T) (*Instance, *swapState, *mockNet) {
	xmrmaker, net := newTestInstanceAndNet(t)

	swapState, err := newSwapStateFromStart(
		xmrmaker.backend,
		types.NewOffer(
			coins.ProvidesXMR,
			coins.StrToDecimal("0.1"),
			coins.StrToDecimal("1"),
			coins.StrToExchangeRate("0.1"),
			types.EthAssetETH,
		),
		&types.OfferExtra{},
		xmrmaker.offerManager,
		coins.MoneroToPiconero(coins.StrToDecimal("0.1")),
		desiredAmount,
	)
	require.NoError(t, err)
	return xmrmaker, swapState, net
}

func newTestSwapState(t *testing.T) (*Instance, *swapState) {
	xmrmaker, swapState, _ := newTestSwapStateAndDB(t)
	return xmrmaker, swapState
}

func newTestXMRTakerSendKeysMessage(t *testing.T) (*message.SendKeysMessage, *pcommon.KeysAndProof) {
	keysAndProof, err := pcommon.GenerateKeysAndProof()
	require.NoError(t, err)

	msg := &message.SendKeysMessage{
		PublicSpendKey:     keysAndProof.PublicKeyPair.SpendKey(),
		PrivateViewKey:     keysAndProof.PrivateKeyPair.ViewKey(),
		DLEqProof:          keysAndProof.DLEqProof.Proof(),
		Secp256k1PublicKey: keysAndProof.Secp256k1PublicKey,
	}

	return msg, keysAndProof
}

func newSwap(
	t *testing.T,
	ss *swapState,
	claimKey,
	refundKey types.Hash,
	amount *big.Int,
	timeout time.Duration,
) ethcommon.Hash {
	tm := big.NewInt(int64(timeout.Seconds()))
	if types.IsHashZero(claimKey) {
		claimKey = ss.secp256k1Pub.Keccak256()
	}

	txOpts, err := ss.ETHClient().TxOpts(ss.ctx)
	require.NoError(t, err)
	txOpts.Value = amount

	ethAddr := ss.ETHClient().Address()
	nonce := big.NewInt(0)
	asset := types.EthAssetETH
	tx, err := ss.Contract().NewSwap(txOpts, claimKey, refundKey, ethAddr, tm,
		ethcommon.Address(asset), amount, nonce)
	require.NoError(t, err)
	receipt := tests.MineTransaction(t, ss.ETHClient().Raw(), tx)

	require.Equal(t, 1, len(receipt.Logs))
	ss.contractSwapID, err = contracts.GetIDFromLog(receipt.Logs[0])
	require.NoError(t, err)

	t0, t1, err := contracts.GetTimeoutsFromLog(receipt.Logs[0])
	require.NoError(t, err)

	ss.contractSwap = &contracts.SwapFactorySwap{
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

func TestNewSwapState_generateAndSetKeys(t *testing.T) {
	_, swapState := newTestSwapState(t)
	require.NotNil(t, swapState.privkeys)
	require.NotNil(t, swapState.pubkeys)
	require.NotNil(t, swapState.dleqProof)
}

func TestSwapState_ClaimFunds(t *testing.T) {
	_, swapState := newTestSwapState(t)

	claimKey := swapState.secp256k1Pub.Keccak256()
	newSwap(t, swapState, claimKey,
		[32]byte{}, big.NewInt(33), defaultTimeoutDuration)

	txOpts, err := swapState.ETHClient().TxOpts(swapState.ctx)
	require.NoError(t, err)
	tx, err := swapState.Contract().SetReady(txOpts, *swapState.contractSwap)
	require.NoError(t, err)
	tests.MineTransaction(t, swapState.ETHClient().Raw(), tx)

	txHash, err := swapState.claimFunds()
	require.NoError(t, err)
	require.NotEqual(t, "", txHash)
	require.True(t, swapState.info.Status.IsOngoing())
}

func TestSwapState_handleSendKeysMessage(t *testing.T) {
	_, s := newTestSwapState(t)

	msg := &message.SendKeysMessage{}
	err := s.handleSendKeysMessage(msg)
	require.Equal(t, errMissingKeys, err)

	msg, xmrtakerKeysAndProof := newTestXMRTakerSendKeysMessage(t)

	err = s.handleSendKeysMessage(msg)
	require.NoError(t, err)
	require.Equal(t, EventETHLockedType, s.nextExpectedEvent)
	require.Equal(t, xmrtakerKeysAndProof.PublicKeyPair.SpendKey().String(), s.xmrtakerPublicSpendKey.String())
	require.Equal(t, xmrtakerKeysAndProof.PrivateKeyPair.ViewKey().String(), s.xmrtakerPrivateViewKey.String())
	require.True(t, s.info.Status.IsOngoing())
}

func TestSwapState_HandleProtocolMessage_NotifyETHLocked_ok(t *testing.T) {
	_, s, net := newTestSwapStateAndNet(t)
	defer s.cancel()
	s.nextExpectedEvent = EventETHLockedType

	xmrtakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	err = s.setXMRTakerKeys(
		xmrtakerKeysAndProof.PublicKeyPair.SpendKey(),
		xmrtakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrtakerKeysAndProof.Secp256k1PublicKey,
	)
	require.NoError(t, err)

	msg := &message.NotifyETHLocked{}
	err = s.HandleProtocolMessage(msg)
	require.True(t, errors.Is(err, errMissingAddress))

	duration := common.SwapTimeoutFromEnv(common.Development)
	hash := newSwap(t, s, s.secp256k1Pub.Keccak256(), s.xmrtakerSecp256K1PublicKey.Keccak256(),
		desiredAmount.BigInt(), duration)
	addr := s.ContractAddr()

	msg = &message.NotifyETHLocked{
		Address:        addr,
		ContractSwapID: s.contractSwapID,
		TxHash:         hash,
		ContractSwap:   s.contractSwap,
	}

	err = s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	resp := net.LastSentMessage()
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyXMRLockType, resp.Type())
	require.Equal(t, duration, s.t1.Sub(s.t0))
	require.Equal(t, EventContractReadyType, s.nextExpectedEvent)
	require.True(t, s.info.Status.IsOngoing())
}

func TestSwapState_HandleProtocolMessage_NotifyETHLocked_timeout(t *testing.T) {
	_, s := newTestSwapState(t)
	defer s.cancel()

	xmrtakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	err = s.setXMRTakerKeys(
		xmrtakerKeysAndProof.PublicKeyPair.SpendKey(),
		xmrtakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrtakerKeysAndProof.Secp256k1PublicKey,
	)
	require.NoError(t, err)

	msg := &message.NotifyETHLocked{}
	err = s.HandleProtocolMessage(msg)
	require.True(t, errors.Is(err, errMissingAddress))

	duration, err := time.ParseDuration("5s")
	require.NoError(t, err)
	_ = newSwap(t, s, s.secp256k1Pub.Keccak256(), s.xmrtakerSecp256K1PublicKey.Keccak256(),
		desiredAmount.BigInt(), duration)
	addr := s.ContractAddr()

	err = s.setContract(addr)
	require.NoError(t, err)
	err = s.setNextExpectedEvent(EventContractReadyType)
	require.NoError(t, err)
	require.Equal(t, duration, s.t1.Sub(s.t0))
	require.Equal(t, EventContractReadyType, s.nextExpectedEvent)

	go s.runT0ExpirationHandler()

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
	_, s, db := newTestSwapStateAndDB(t)
	db.EXPECT().PutOffer(s.offer)

	xmrtakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	err = s.setXMRTakerKeys(
		xmrtakerKeysAndProof.PublicKeyPair.SpendKey(),
		xmrtakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrtakerKeysAndProof.Secp256k1PublicKey,
	)
	require.NoError(t, err)

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)

	refundKey := xmrtakerKeysAndProof.Secp256k1PublicKey.Keccak256()
	newSwap(t, s, [32]byte{}, refundKey, desiredAmount.BigInt(), duration)

	// lock XMR
	_, err = s.lockFunds(coins.MoneroToPiconero(s.info.ProvidedAmount))
	require.NoError(t, err)

	// call refund w/ XMRTaker's spend key
	secret := xmrtakerKeysAndProof.PrivateKeyPair.SpendKeyBytes()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret))

	txOpts, err := s.ETHClient().TxOpts(s.ctx)
	require.NoError(t, err)
	tx, err := s.Contract().Refund(txOpts, *s.contractSwap, sc)
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(s.Backend.Ctx(), s.ETHClient().Raw(), tx.Hash())
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
	_, s, db := newTestSwapStateAndDB(t)
	db.EXPECT().PutOffer(s.offer)

	xmrtakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)
	err = s.setXMRTakerKeys(
		xmrtakerKeysAndProof.PublicKeyPair.SpendKey(),
		xmrtakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrtakerKeysAndProof.Secp256k1PublicKey,
	)
	require.NoError(t, err)

	duration, err := time.ParseDuration("10m")
	require.NoError(t, err)

	refundKey := xmrtakerKeysAndProof.Secp256k1PublicKey.Keccak256()
	newSwap(t, s, [32]byte{}, refundKey, desiredAmount.BigInt(), duration)

	// lock XMR
	_, err = s.lockFunds(coins.MoneroToPiconero(s.info.ProvidedAmount))
	require.NoError(t, err)

	balAfterLock, err := s.XMRClient().GetBalance(0)
	require.NoError(t, err)
	t.Logf("Balance after locking funds: %s XMR (%d blocks to unlock)",
		coins.FmtPiconeroAmtAsXMR(balAfterLock.Balance), balAfterLock.BlocksToUnlock)

	// call refund w/ XMRTaker's secret
	secret := xmrtakerKeysAndProof.DLEqProof.Secret()
	var sc [32]byte
	copy(sc[:], secret[:])

	s.nextExpectedEvent = EventContractReadyType

	txOpts, err := s.ETHClient().TxOpts(s.ctx)
	require.NoError(t, err)
	tx, err := s.Contract().Refund(txOpts, *s.contractSwap, sc)
	require.NoError(t, err)
	receipt := tests.MineTransaction(t, s.ETHClient().Raw(), tx)

	require.Equal(t, 1, len(receipt.Logs))
	require.Equal(t, 3, len(receipt.Logs[0].Topics))
	require.Equal(t, refundedTopic, receipt.Logs[0].Topics[0])

	// runContractEventWatcher will trigger EventETHRefunded,
	// which will then set the next expected event to EventExit.
	for status := range s.info.StatusCh() {
		if !status.IsOngoing() {
			break
		}
	}

	balance, err := s.XMRClient().GetBalance(0)
	require.NoError(t, err)
	t.Logf("End balance after refund: %s XMR (%d blocks to unlock)",
		coins.FmtPiconeroAmtAsXMR(balance.Balance), balance.BlocksToUnlock)
	require.Greater(t, balance.Balance, balAfterLock.Balance) // increased by refund (minus some fees)
	require.Equal(t, types.CompletedRefund, s.info.Status)
}

func TestSwapState_Exit_Aborted(t *testing.T) {
	_, s, db := newTestSwapStateAndDB(t)
	db.EXPECT().PutOffer(s.offer)

	s.nextExpectedEvent = EventETHLockedType
	err := s.Exit()
	require.NoError(t, err)
	require.Equal(t, types.CompletedAbort, s.info.Status)
}

func TestSwapState_Exit_Aborted_1(t *testing.T) {
	_, s, db := newTestSwapStateAndDB(t)
	db.EXPECT().PutOffer(s.offer)

	s.nextExpectedEvent = EventETHRefundedType
	err := s.Exit()
	require.True(t, errors.Is(err, errUnexpectedMessageType))
	require.Equal(t, types.CompletedAbort, s.info.Status)
}

func TestSwapState_Exit_Success(t *testing.T) {
	b, s := newTestSwapState(t)
	s.nextExpectedEvent = EventNoneType
	min := coins.StrToDecimal("0.1")
	max := coins.StrToDecimal("0.2")
	rate := coins.ToExchangeRate(coins.StrToDecimal("0.1"))
	s.offer = types.NewOffer(coins.ProvidesXMR, min, max, rate, types.EthAssetETH)
	s.info.SetStatus(types.CompletedSuccess)
	err := s.Exit()
	require.NoError(t, err)

	// since the swap was successful, the offer should not have been re-added.
	o, oe, _ := b.offerManager.GetOffer(s.offer.ID)
	require.Nil(t, o)
	require.Nil(t, oe)
}

func TestSwapState_Exit_Refunded(t *testing.T) {
	b, s, db := newTestSwapStateAndDB(t)

	b.net.(*MockP2pHost).EXPECT().RefreshNamespaces()

	min := coins.StrToDecimal("0.1")
	max := coins.StrToDecimal("0.2")
	rate := coins.ToExchangeRate(coins.StrToDecimal("0.1"))
	s.offer = types.NewOffer(coins.ProvidesXMR, min, max, rate, types.EthAssetETH)
	db.EXPECT().PutOffer(s.offer)
	_, err := b.MakeOffer(s.offer, nil)
	require.NoError(t, err)

	s.info.SetStatus(types.CompletedRefund)
	err = s.Exit()
	require.NoError(t, err)

	// since the swap was not successful, the offer should be re-added to the offer manager.
	o, oe, err := b.offerManager.GetOffer(s.offer.ID)
	require.NoError(t, err)
	require.NotNil(t, o)
	require.NotNil(t, oe)
}
