// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package xmrtaker

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net/message"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	pswap "github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/tests"
)

var (
	_             = logging.SetLogLevel("protocol", "debug")
	_             = logging.SetLogLevel("xmrtaker", "debug")
	testPeerID, _ = peer.Decode("12D3KooWQQRJuKTZ35eiHGNPGDpQqjpJSdaxEMJRxi6NWFrrvQVi")
)

type mockNet struct {
	msgMu sync.Mutex     // lock needed, as SendSwapMessage is called async from timeout handlers
	msg   common.Message // last value passed to SendSwapMessage
}

func (n *mockNet) LastSentMessage() common.Message {
	n.msgMu.Lock()
	defer n.msgMu.Unlock()
	return n.msg
}

func (n *mockNet) SendSwapMessage(msg common.Message, _ types.Hash) error {
	n.msgMu.Lock()
	defer n.msgMu.Unlock()
	n.msg = msg
	return nil
}

func (n *mockNet) DiscoverRelayers() ([]peer.ID, error) {
	return nil, nil
}

func (n *mockNet) SubmitClaimToRelayer(_ peer.ID, _ *message.RelayClaimRequest) (*message.RelayClaimResponse, error) {
	return new(message.RelayClaimResponse), nil
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

func newBackendAndNet(t *testing.T) (backend.Backend, *mockNet) {
	pk := tests.GetTakerTestKey(t)
	ec := extethclient.CreateTestClient(t, pk)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
	})

	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, ec.ChainID())
	require.NoError(t, err)

	var forwarderAddr ethcommon.Address
	_, tx, _, err := contracts.DeploySwapCreator(txOpts, ec.Raw(), forwarderAddr)
	require.NoError(t, err)

	addr, err := bind.WaitDeployed(ctx, ec.Raw(), tx)
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rdb := backend.NewMockRecoveryDB(ctrl)
	rdb.EXPECT().PutContractSwapInfo(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	rdb.EXPECT().PutSwapPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	rdb.EXPECT().PutCounterpartySwapPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	rdb.EXPECT().PutCounterpartySwapKeys(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	rdb.EXPECT().DeleteSwap(gomock.Any()).Return(nil).AnyTimes()

	net := new(mockNet)
	bcfg := &backend.Config{
		Ctx:             ctx,
		MoneroClient:    monero.CreateWalletClient(t),
		EthereumClient:  ec,
		Environment:     common.Development,
		SwapManager:     newSwapManager(t),
		SwapCreatorAddr: addr,
		Net:             net,
		RecoveryDB:      rdb,
	}

	b, err := backend.NewBackend(bcfg)
	require.NoError(t, err)
	return b, net
}

func newBackend(t *testing.T) backend.Backend {
	b, _ := newBackendAndNet(t)
	return b
}

func newTestSwapStateAndNet(t *testing.T) (*swapState, *mockNet) {
	b, net := newBackendAndNet(t)
	providedAmt := coins.EtherToWei(coins.StrToDecimal("1"))
	expectedAmt := coins.MoneroToPiconero(coins.StrToDecimal("1"))
	exchangeRate := coins.ToExchangeRate(coins.StrToDecimal("1.0")) // 100%
	swapState, err := newSwapStateFromStart(b, testPeerID, types.Hash{}, true,
		providedAmt, expectedAmt, exchangeRate, types.EthAssetETH)
	require.NoError(t, err)
	return swapState, net
}

func newTestSwapState(t *testing.T) *swapState {
	s, _ := newTestSwapStateAndNet(t)
	return s
}

func newTestSwapStateWithERC20(t *testing.T, initialBalance *big.Int) (*swapState, *contracts.TestERC20) {
	b := newBackend(t)

	txOpts, err := b.ETHClient().TxOpts(b.Ctx())
	require.NoError(t, err)

	_, tx, contract, err := contracts.DeployTestERC20(
		txOpts,
		b.ETHClient().Raw(),
		"Mock",
		"MOCK",
		b.ETHClient().Address(),
		initialBalance,
	)
	require.NoError(t, err)
	addr, err := bind.WaitDeployed(b.Ctx(), b.ETHClient().Raw(), tx)
	require.NoError(t, err)

	exchangeRate := coins.ToExchangeRate(apd.New(1, 0)) // 100%
	zeroPiconeros := coins.NewPiconeroAmount(0)
	swapState, err := newSwapStateFromStart(b, testPeerID, types.Hash{}, false,
		coins.IntToWei(1), zeroPiconeros, exchangeRate, types.EthAsset(addr))
	require.NoError(t, err)
	return swapState, contract
}

func newTestXMRMakerSendKeysMessage(t *testing.T) (*message.SendKeysMessage, *pcommon.KeysAndProof) {
	keysAndProof, err := pcommon.GenerateKeysAndProof()
	require.NoError(t, err)

	msg := &message.SendKeysMessage{
		PublicSpendKey:     keysAndProof.PublicKeyPair.SpendKey(),
		PrivateViewKey:     keysAndProof.PrivateKeyPair.ViewKey(),
		DLEqProof:          keysAndProof.DLEqProof.Proof(),
		Secp256k1PublicKey: keysAndProof.Secp256k1PublicKey,
		EthAddress:         ethcommon.Address{0x1},
		ProvidedAmount:     apd.New(1, 0),
	}

	return msg, keysAndProof
}

func TestSwapState_HandleProtocolMessage_SendKeysMessage(t *testing.T) {
	s, net := newTestSwapStateAndNet(t)
	defer s.cancel()

	msg := &message.SendKeysMessage{}
	err := s.HandleProtocolMessage(msg)
	require.True(t, errors.Is(err, errMissingProvidedAmount))

	msg, xmrmakerKeysAndProof := newTestXMRMakerSendKeysMessage(t)

	err = s.HandleProtocolMessage(msg)
	require.NoError(t, err)

	resp := net.LastSentMessage()
	require.NotNil(t, resp)
	require.Equal(t, s.SwapTimeout(), s.t1.Sub(s.t0))
	require.Equal(t, xmrmakerKeysAndProof.PublicKeyPair.SpendKey().String(), s.xmrmakerPublicSpendKey.String())
	require.Equal(t, xmrmakerKeysAndProof.PrivateKeyPair.ViewKey().String(), s.xmrmakerPrivateViewKey.String())
}

// test the case where XMRTaker deploys and locks her eth, but XMRMaker never locks his monero.
// XMRTaker should call refund before the timeout t0.
func TestSwapState_HandleProtocolMessage_SendKeysMessage_Refund(t *testing.T) {
	s, net := newTestSwapStateAndNet(t)
	defer s.cancel()
	s.SetSwapTimeout(time.Second * 15)

	msg, xmrmakerKeysAndProof := newTestXMRMakerSendKeysMessage(t)

	err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)

	resp := net.LastSentMessage()
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyETHLockedType, resp.Type())
	require.Equal(t, s.SwapTimeout(), s.t1.Sub(s.t0))
	require.Equal(t, xmrmakerKeysAndProof.PublicKeyPair.SpendKey().String(), s.xmrmakerPublicSpendKey.String())
	require.Equal(t, xmrmakerKeysAndProof.PrivateKeyPair.ViewKey().String(), s.xmrmakerPrivateViewKey.String())

	// ensure we refund before t0
	for status := range s.statusCh {
		if status == types.CompletedRefund {
			// check this is before t0
			// TODO: remove the 10-second buffer, this is needed for now
			// because the exact refund time isn't stored, and the time
			// between the refund happening and this line being called
			// causes it to fail
			require.Greater(t, s.t0.Add(time.Second*10), time.Now())
			break
		} else if !status.IsOngoing() {
			t.Fatalf("got wrong exit status %s, expected CompletedRefund", status)
		}
	}

	// check swap is marked completed
	stage, err := s.SwapCreator().Swaps(nil, s.contractSwapID)
	require.NoError(t, err)
	require.Equal(t, contracts.StageCompleted, stage)
}

func lockXMRFunds(
	t *testing.T,
	ctx context.Context, //nolint:revive
	wc monero.WalletClient,
	destAddr *mcrypto.Address,
	amount *coins.PiconeroAmount,
) {
	monero.MineMinXMRBalance(t, wc, amount)
	_, err := wc.Transfer(ctx, destAddr, 0, amount, monero.MinSpendConfirmations)
	require.NoError(t, err)
}

func TestSwapState_NotifyXMRLock(t *testing.T) {
	s := newTestSwapState(t)
	defer s.cancel()
	s.nextExpectedEvent = EventXMRLockedType

	xmrmakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	err = s.setXMRMakerKeys(
		xmrmakerKeysAndProof.PublicKeyPair.SpendKey(),
		xmrmakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrmakerKeysAndProof.Secp256k1PublicKey,
	)
	require.NoError(t, err)

	_, err = s.lockAsset()
	require.NoError(t, err)

	kp := mcrypto.SumSpendAndViewKeys(xmrmakerKeysAndProof.PublicKeyPair, s.pubkeys)
	xmrAddr := kp.Address(common.Development)

	lockXMRFunds(t, s.ctx, s.XMRClient(), xmrAddr, s.expectedPiconeroAmount())
	event := newEventXMRLocked()
	s.eventCh <- event
	err = <-event.errCh
	require.NoError(t, err)
	require.Equal(t, EventETHClaimedType, s.nextExpectedEvent)
}

// test the case where the monero is locked, but XMRMaker never claims.
// XMRTaker should call refund after the timeout t1.
func TestSwapState_NotifyXMRLock_Refund(t *testing.T) {
	s := newTestSwapState(t)
	defer s.cancel()
	s.nextExpectedEvent = EventXMRLockedType
	s.SetSwapTimeout(time.Second * 3)

	xmrmakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	err = s.setXMRMakerKeys(
		xmrmakerKeysAndProof.PublicKeyPair.SpendKey(),
		xmrmakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrmakerKeysAndProof.Secp256k1PublicKey,
	)
	require.NoError(t, err)

	_, err = s.lockAsset()
	require.NoError(t, err)

	kp := mcrypto.SumSpendAndViewKeys(xmrmakerKeysAndProof.PublicKeyPair, s.pubkeys)
	xmrAddr := kp.Address(common.Development)

	lockXMRFunds(t, s.ctx, s.XMRClient(), xmrAddr, s.expectedPiconeroAmount())
	event := newEventXMRLocked()
	s.eventCh <- event
	err = <-event.errCh
	require.NoError(t, err)
	require.Equal(t, EventETHClaimedType, s.nextExpectedEvent)

	for status := range s.statusCh {
		if status == types.CompletedRefund {
			// check this is after t1
			require.Less(t, s.t1, time.Now())
			break
		} else if !status.IsOngoing() {
			t.Fatalf("got wrong exit status %s, expected CompletedRefund", status)
		}
	}

	// check balance of contract is 0
	balance, err := s.ETHClient().Raw().BalanceAt(context.Background(), s.SwapCreatorAddr(), nil)
	require.NoError(t, err)
	require.Equal(t, uint64(0), balance.Uint64())
}

func TestExit_afterSendKeysMessage(t *testing.T) {
	s := newTestSwapState(t)
	defer s.cancel()
	s.nextExpectedEvent = EventKeysReceivedType
	err := s.Exit()
	require.NoError(t, err)
	info, err := s.SwapManager().GetPastSwap(s.info.OfferID)
	require.NoError(t, err)
	require.Equal(t, types.CompletedAbort, info.Status)
}

func TestExit_afterNotifyXMRLock(t *testing.T) {
	s := newTestSwapState(t)
	defer s.cancel()
	s.nextExpectedEvent = EventXMRLockedType

	xmrmakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	err = s.setXMRMakerKeys(
		xmrmakerKeysAndProof.PublicKeyPair.SpendKey(),
		xmrmakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrmakerKeysAndProof.Secp256k1PublicKey,
	)
	require.NoError(t, err)

	_, err = s.lockAsset()
	require.NoError(t, err)

	err = s.Exit()
	require.NoError(t, err)

	info, err := s.SwapManager().GetPastSwap(s.info.OfferID)
	require.NoError(t, err)
	require.Equal(t, types.CompletedRefund, info.Status)
}

func TestExit_afterNotifyClaimed(t *testing.T) {
	s := newTestSwapState(t)
	defer s.cancel()
	s.nextExpectedEvent = EventETHClaimedType

	xmrmakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	err = s.setXMRMakerKeys(
		xmrmakerKeysAndProof.PublicKeyPair.SpendKey(),
		xmrmakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrmakerKeysAndProof.Secp256k1PublicKey,
	)
	require.NoError(t, err)

	_, err = s.lockAsset()
	require.NoError(t, err)

	err = s.Exit()
	require.NoError(t, err)

	info, err := s.SwapManager().GetPastSwap(s.info.OfferID)
	require.NoError(t, err)
	require.Equal(t, types.CompletedRefund, info.Status)
}

func TestExit_invalidNextMessageType(t *testing.T) {
	// this case shouldn't ever really happen
	s := newTestSwapState(t)
	defer s.cancel()
	s.nextExpectedEvent = EventExitType

	xmrmakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	err = s.setXMRMakerKeys(
		xmrmakerKeysAndProof.PublicKeyPair.SpendKey(),
		xmrmakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrmakerKeysAndProof.Secp256k1PublicKey,
	)
	require.NoError(t, err)

	_, err = s.lockAsset()
	require.NoError(t, err)

	err = s.Exit()
	require.True(t, errors.Is(err, errUnexpectedEventType))

	info, err := s.SwapManager().GetPastSwap(s.info.OfferID)
	require.NoError(t, err)
	require.Equal(t, types.CompletedAbort, info.Status)
}

func TestSwapState_ApproveToken(t *testing.T) {
	initialBalance := big.NewInt(999999)
	s, contract := newTestSwapStateWithERC20(t, initialBalance)
	err := s.approveToken()
	require.NoError(t, err)
	allowance, err := contract.Allowance(&bind.CallOpts{}, s.ETHClient().Address(), s.SwapCreatorAddr())
	require.NoError(t, err)
	require.Equal(t, initialBalance, allowance)
}
