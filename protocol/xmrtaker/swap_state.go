// Package xmrtaker manages the swap state of individual swaps where the local swapd
// instance is offering Ethereum assets and accepting Monero in return.
package xmrtaker

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/cockroachdb/apd/v3"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/crypto/secp256k1"
	"github.com/athanorlabs/atomic-swap/db"
	"github.com/athanorlabs/atomic-swap/dleq"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/watcher"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net/message"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	pswap "github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/txsender"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/fatih/color"
)

const revertSwapCompleted = "swap is already completed"

var claimedTopic = common.GetTopic(common.ClaimedEventSignature)

// swapState is an instance of a swap. it holds the info needed for the swap,
// and its current state.
type swapState struct {
	backend.Backend
	sender txsender.Sender

	ctx            context.Context
	cancel         context.CancelFunc
	noTransferBack bool

	info           *pswap.Info
	statusCh       chan types.Status
	providedAmount EthereumAssetAmount

	// our keys for this session
	dleqProof    *dleq.Proof
	secp256k1Pub *secp256k1.PublicKey
	privkeys     *mcrypto.PrivateKeyPair
	pubkeys      *mcrypto.PublicKeyPair

	// XMRMaker's keys for this session
	xmrmakerPublicSpendKey     *mcrypto.PublicKey
	xmrmakerPrivateViewKey     *mcrypto.PrivateViewKey
	xmrmakerSecp256k1PublicKey *secp256k1.PublicKey
	xmrmakerAddress            ethcommon.Address

	// block height at start of swap used for fast wallet creation
	walletScanHeight uint64

	// swap contract and timeouts in it; set once contract is deployed
	contractSwapID [32]byte
	contractSwap   *contracts.SwapFactorySwap
	t0, t1         time.Time

	// tracks the state of the swap
	nextExpectedEvent EventType
	// set to true once funds are locked
	fundsLocked bool

	// channels

	// channel for swap events
	// the event handler in event.go ensures only one event is being handled at a time
	eventCh chan Event
	// channel for `Claimed` logs seen on-chain
	logClaimedCh chan ethtypes.Log
	// signals the t0 expiration handler to return
	xmrLockedCh chan struct{}
	// signals the t1 expiration handler to return
	claimedCh chan struct{}
	// signals to the creator xmrmaker instance that it can delete this swap
	done chan struct{}
}

func newSwapStateFromStart(
	b backend.Backend,
	offerID types.Hash,
	noTransferBack bool,
	providedAmount EthereumAssetAmount,
	expectedAmount *coins.PiconeroAmount,
	exchangeRate *coins.ExchangeRate,
	ethAsset types.EthAsset,
) (*swapState, error) {
	stage := types.ExpectingKeys
	statusCh := make(chan types.Status, 16)

	moneroStartNumber, err := b.XMRClient().GetHeight()
	if err != nil {
		return nil, err
	}

	// reduce the scan height a little in case there is a block reorg
	if moneroStartNumber >= monero.MinSpendConfirmations {
		moneroStartNumber -= monero.MinSpendConfirmations
	}

	ethHeader, err := b.ETHClient().Raw().HeaderByNumber(b.Ctx(), nil)
	if err != nil {
		return nil, err
	}

	info := pswap.NewInfo(
		offerID,
		coins.ProvidesETH,
		providedAmount.AsStandard(),
		expectedAmount.AsMonero(),
		exchangeRate,
		ethAsset,
		stage,
		moneroStartNumber,
		statusCh,
	)
	if err = b.SwapManager().AddSwap(info); err != nil {
		return nil, err
	}

	s, err := newSwapState(
		b,
		noTransferBack,
		info,
		ethHeader.Number,
		moneroStartNumber,
	)
	if err != nil {
		return nil, err
	}

	statusCh <- stage
	return s, nil
}

func newSwapStateFromOngoing(
	b backend.Backend,
	info *pswap.Info,
	noTransferBack bool,
	ethSwapInfo *db.EthereumSwapInfo,
	sk *mcrypto.PrivateKeyPair,
) (*swapState, error) {
	if info.Status != types.ETHLocked && info.Status != types.ContractReady {
		return nil, errInvalidStageForRecovery
	}

	makerSk, makerVk, err := b.RecoveryDB().GetCounterpartySwapKeys(info.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get xmrmaker swap keys from db: %w", err)
	}

	s, err := newSwapState(
		b,
		noTransferBack,
		info,
		ethSwapInfo.StartNumber,
		info.MoneroStartHeight,
	)
	if err != nil {
		return nil, err
	}

	if b.ContractAddr() != ethSwapInfo.ContractAddress {
		return nil, errContractAddrMismatch(ethSwapInfo.ContractAddress.String())
	}

	s.setTimeouts(ethSwapInfo.Swap.Timeout0, ethSwapInfo.Swap.Timeout1)
	s.privkeys = sk
	s.pubkeys = sk.PublicKeyPair()
	s.contractSwapID = ethSwapInfo.SwapID
	s.contractSwap = ethSwapInfo.Swap
	s.xmrmakerPublicSpendKey = makerSk
	s.xmrmakerPrivateViewKey = makerVk

	if info.Status == types.ETHLocked {
		go s.checkForXMRLock()
	}
	return s, nil
}

func newSwapState(
	b backend.Backend,
	noTransferBack bool,
	info *pswap.Info,
	ethStartNumber *big.Int,
	moneroStartNumber uint64,
) (*swapState, error) {
	// If the user specified `--external-signer=true` (no private eth key in the
	// client) and explicitly set `--no-transfer-back`, we override their
	// decision and set it back to `false`, because an external signer (UI) must
	// be used, which will prompt the user to set their XMR address for funds to
	// be transferred-back to.
	if !b.ETHClient().HasPrivateKey() {
		noTransferBack = false // front-end must set final deposit address
	}

	var sender txsender.Sender
	if info.EthAsset != types.EthAssetETH {
		erc20Contract, err := contracts.NewIERC20(info.EthAsset.Address(), b.ETHClient().Raw())
		if err != nil {
			return nil, err
		}

		sender, err = b.NewTxSender(info.EthAsset.Address(), erc20Contract)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		sender, err = b.NewTxSender(info.EthAsset.Address(), nil)
		if err != nil {
			return nil, err
		}
	}

	// set up ethereum event watchers
	const logChSize = 16
	logClaimedCh := make(chan ethtypes.Log, logChSize)

	ctx, cancel := context.WithCancel(b.Ctx())

	claimedWatcher := watcher.NewEventFilter(
		ctx,
		b.ETHClient().Raw(),
		b.ContractAddr(),
		ethStartNumber,
		claimedTopic,
		logClaimedCh,
	)

	err := claimedWatcher.Start()
	if err != nil {
		cancel()
		return nil, err
	}

	// note: if this is recovering an ongoing swap, this will only
	// be invoked if our status is ETHLocked or ContractReady; ie.
	// we've locked ETH, but not yet claimed or refunded.
	//
	// dleqProof and secp256k1Pub are never set, as they are only used
	// in the swap step before or where ETH is locked.
	//
	// similarly, xmrmaker secp256k1 public keys and ETH address are also
	// never set, as they're only used in the ETH lock step.
	s := &swapState{
		ctx:               ctx,
		cancel:            cancel,
		Backend:           b,
		sender:            sender,
		noTransferBack:    noTransferBack,
		walletScanHeight:  moneroStartNumber,
		nextExpectedEvent: nextExpectedEventFromStatus(info.Status),
		eventCh:           make(chan Event),
		logClaimedCh:      logClaimedCh,
		xmrLockedCh:       make(chan struct{}),
		claimedCh:         make(chan struct{}),
		done:              make(chan struct{}),
		info:              info,
		providedAmount:    coins.EtherToWei(info.ProvidedAmount),
		statusCh:          info.StatusCh(),
	}

	if err := s.generateAndSetKeys(); err != nil {
		return nil, err
	}

	go s.waitForSendKeysMessage()
	go s.runHandleEvents()
	go s.runContractEventWatcher()
	return s, nil
}

func (s *swapState) waitForSendKeysMessage() {
	waitDuration := time.Minute * 5
	timer := time.After(waitDuration)
	select {
	case <-s.ctx.Done():
		return
	case <-timer:
	}

	// check if we've received a response from the counterparty yet
	if reflect.TypeOf(s.nextExpectedEvent) != reflect.TypeOf(&EventKeysReceived{}) {
		return
	}

	// if not, just exit the swap
	if err := s.Exit(); err != nil {
		log.Warnf("Swap exit failure: %s", err)
	}
}

// SendKeysMessage ...
func (s *swapState) SendKeysMessage() common.Message {
	return &message.SendKeysMessage{
		PublicSpendKey:     s.pubkeys.SpendKey(),
		PrivateViewKey:     s.privkeys.ViewKey(),
		DLEqProof:          s.dleqProof.Proof(),
		Secp256k1PublicKey: s.secp256k1Pub,
	}
}

// ExpectedAmount returns the amount received, or expected to be received, at the end of the swap
func (s *swapState) ExpectedAmount() *apd.Decimal {
	return s.info.ExpectedAmount
}

func (s *swapState) expectedPiconeroAmount() *coins.PiconeroAmount {
	return coins.MoneroToPiconero(s.info.ExpectedAmount)
}

// ID returns the ID of the swap
func (s *swapState) ID() types.Hash {
	return s.info.ID
}

// Exit is called by the network when the protocol stream closes, or if the swap_refund RPC endpoint is called.
// It exists the swap by refunding if necessary. If no locking has been done, it simply aborts the swap.
// If the swap already completed successfully, this function does not do anything regarding the protocol.
func (s *swapState) Exit() error {
	event := newEventExit()
	s.eventCh <- event
	return <-event.errCh
}

// exit is the same as Exit, but assumes the calling code block already holds the swapState lock.
func (s *swapState) exit() error {
	defer func() {
		err := s.SwapManager().CompleteOngoingSwap(s.info)
		if err != nil {
			log.Warnf("failed to mark swap %s as completed: %s", s.info.ID, err)
			return
		}

		err = s.Backend.RecoveryDB().DeleteSwap(s.ID())
		if err != nil {
			log.Warnf("failed to delete temporary swap info %s from db: %s", s.ID(), err)
		}

		// Stop all per-swap goroutines
		s.cancel()
		close(s.done)

		var exitLog string
		switch s.info.Status {
		case types.CompletedSuccess:
			exitLog = color.New(color.Bold).Sprintf("**swap completed successfully: id=%s**", s.ID())
		case types.CompletedRefund:
			exitLog = color.New(color.Bold).Sprintf("**swap refunded successfully: id=%s**", s.ID())
		case types.CompletedAbort:
			exitLog = color.New(color.Bold).Sprintf("**swap aborted: id=%s**", s.ID())
		}

		log.Info(exitLog)
	}()

	log.Debugf("attempting to exit swap: nextExpectedEvent=%s", s.nextExpectedEvent)

	switch s.nextExpectedEvent {
	case EventKeysReceivedType:
		// we are fine, as we only just initiated the protocol.
		s.clearNextExpectedEvent(types.CompletedAbort)
		return nil
	case EventXMRLockedType, EventETHClaimedType:
		// for EventXMRLocked, we already deployed the contract,
		// so we should call Refund().
		//
		// for EventETHClaimed, the XMR has been locked, but the
		// ETH hasn't been claimed, but the contract has been set to ready.
		// we should also refund in this case, since we might be past t1.
		txHash, err := s.tryRefund()
		if err != nil {
			if strings.Contains(err.Error(), revertSwapCompleted) {
				// note: this should NOT ever error; it could if the ethclient
				// or monero clients crash during the course of the claim,
				// but that would be very bad.
				err = s.tryClaim()
				if err != nil {
					return fmt.Errorf("failed to claim even though swap was completed on-chain: %w", err)
				}
			}

			return fmt.Errorf("failed to refund: %w", err)
		}

		s.clearNextExpectedEvent(types.CompletedRefund)
		log.Infof("refunded ether: transaction hash=%s", txHash)
		return nil
	case EventNoneType:
		// the swap completed already, do nothing
		return nil
	default:
		log.Errorf("unexpected nextExpectedEvent: %s", s.nextExpectedEvent)
		s.clearNextExpectedEvent(types.CompletedAbort)
		return errUnexpectedEventType
	}
}

// doRefund is called by the RPC function swap_refund.
// If it's possible to refund the ongoing swap, it does that, then notifies the counterparty.
func (s *swapState) doRefund() (ethcommon.Hash, error) {
	switch s.nextExpectedEvent {
	case EventXMRLockedType, EventETHClaimedType:
		event := newEventShouldRefund()
		s.eventCh <- event
		err := <-event.errCh
		if err != nil {
			return ethcommon.Hash{}, err
		}

		txHash := <-event.txHashCh
		return txHash, nil
	default:
		return ethcommon.Hash{}, errCannotRefund
	}
}

func (s *swapState) tryRefund() (ethcommon.Hash, error) {
	stage, err := s.Contract().Swaps(s.ETHClient().CallOpts(s.ctx), s.contractSwapID)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	switch stage {
	case contracts.StageInvalid:
		return ethcommon.Hash{}, fmt.Errorf("%w: contract swap ID: %s", errRefundInvalid, s.contractSwapID)
	case contracts.StageCompleted:
		return ethcommon.Hash{}, errRefundSwapCompleted
	case contracts.StagePending, contracts.StageReady:
		// do nothing
	default:
		panic("Unhandled stage value")
	}

	isReady := stage == contracts.StageReady

	ts, err := s.ETHClient().LatestBlockTimestamp(s.ctx)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	log.Debugf("tryRefund isReady=%v untilT0=%vs untilT1=%vs",
		isReady, s.t0.Sub(ts).Seconds(), s.t1.Sub(ts).Seconds())

	if ts.Before(s.t0) && !isReady {
		txHash, err := s.refund() //nolint:govet
		// TODO: Have refund() return errors that we can use errors.Is to check against
		if err == nil {
			return txHash, nil
		}

		// There is a small, but non-zero chance that our transaction gets placed in a block that is after T0
		// even though the current block is before T0. In this case, the transaction will be reverted, the
		// gas fee is lost, but we can wait until T1 and try again.
		log.Warnf("first refund attempt failed: err=%s", err)
	}

	if ts.After(s.t1) {
		return s.refund()
	}

	// the contract is "ready", so we can't do anything until
	// the counterparty claims or until t1 passes.
	//
	// we let the runT1ExpirationHandler() routine continue to run and read
	// from s.eventCh for EventShouldRefund or EventETHClaimed.
	// (since this function is called from inside the event handler routine,
	// it won't handle those events while this function is executing.)
	log.Infof("waiting until time %s to refund", s.t1)

	waitCtx, waitCtxCancel := context.WithCancel(s.ctx)
	defer waitCtxCancel()

	waitCh := make(chan error)
	go func() {
		waitCh <- s.ETHClient().WaitForTimestamp(waitCtx, s.t1)
		close(waitCh)
	}()

	select {
	case event := <-s.eventCh:
		log.Debugf("got event %s while waiting for T1", event.Type())
		switch event.(type) {
		case *EventShouldRefund:
			return s.refund()
		case *EventETHClaimed:
			// we should claim; returning this error
			// causes the calling function to claim
			return ethcommon.Hash{}, fmt.Errorf(revertSwapCompleted)
		default:
			panic(fmt.Sprintf("got unexpected event while waiting for Claimed/T1: %s", event))
		}
	case err = <-waitCh:
		if err != nil {
			return ethcommon.Hash{}, fmt.Errorf("failed to wait for T1: %w", err)
		}

		return s.refund()
	}
}

func (s *swapState) setTimeouts(t0, t1 *big.Int) {
	s.t0 = time.Unix(t0.Int64(), 0)
	s.t1 = time.Unix(t1.Int64(), 0)
	s.info.Timeout0 = &s.t0
	s.info.Timeout1 = &s.t1
}

func (s *swapState) generateAndSetKeys() error {
	if s.privkeys != nil {
		panic("generateAndSetKeys should only be called once")
	}

	keysAndProof, err := generateKeys()
	if err != nil {
		return err
	}

	s.dleqProof = keysAndProof.DLEqProof
	s.secp256k1Pub = keysAndProof.Secp256k1PublicKey
	s.privkeys = keysAndProof.PrivateKeyPair
	s.pubkeys = keysAndProof.PublicKeyPair

	return s.Backend.RecoveryDB().PutSwapPrivateKey(s.ID(), s.privkeys.SpendKey())
}

// getSecret secrets returns the current secret scalar used to unlock funds from the contract.
func (s *swapState) getSecret() [32]byte {
	secret := s.dleqProof.Secret()
	var sc [32]byte
	copy(sc[:], secret[:])
	return sc
}

// setXMRMakerKeys sets XMRMaker's public spend key (to be stored in the contract) and XMRMaker's
// private view key (used to check XMR balance before calling Ready())
func (s *swapState) setXMRMakerKeys(
	sk *mcrypto.PublicKey,
	vk *mcrypto.PrivateViewKey,
	secp256k1Pub *secp256k1.PublicKey,
) error {
	s.xmrmakerPublicSpendKey = sk
	s.xmrmakerPrivateViewKey = vk
	s.xmrmakerSecp256k1PublicKey = secp256k1Pub
	return s.Backend.RecoveryDB().PutCounterpartySwapKeys(s.info.ID, sk, vk)
}

func (s *swapState) approveToken() error {
	token, err := contracts.NewIERC20(s.info.EthAsset.Address(), s.ETHClient().Raw())
	if err != nil {
		return fmt.Errorf("failed to instantiate IERC20: %w", err)
	}

	balance, err := token.BalanceOf(s.ETHClient().CallOpts(s.ctx), s.ETHClient().Address())
	if err != nil {
		return fmt.Errorf("failed to get balance for token: %w", err)
	}

	log.Info("approving token for use by the swap contract...")
	_, _, err = s.sender.Approve(s.ContractAddr(), balance)
	if err != nil {
		return fmt.Errorf("failed to approve token: %w", err)
	}

	log.Info("approved token for use by the swap contract")
	return nil
}

// lockAsset calls the Swap contract function new_swap and locks `amount` ether in it.
func (s *swapState) lockAsset() (ethcommon.Hash, error) {
	if s.xmrmakerPublicSpendKey == nil || s.xmrmakerPrivateViewKey == nil {
		panic(errCounterpartyKeysNotSet)
	}

	if s.info.EthAsset != types.EthAssetETH {
		err := s.approveToken()
		if err != nil {
			return ethcommon.Hash{}, err
		}
	}

	cmtXMRTaker := s.secp256k1Pub.Keccak256()
	cmtXMRMaker := s.xmrmakerSecp256k1PublicKey.Keccak256()

	log.Debugf("locking ETH in contract")

	nonce := generateNonce()
	txHash, receipt, err := s.sender.NewSwap(
		cmtXMRMaker,
		cmtXMRTaker,
		s.xmrmakerAddress,
		big.NewInt(int64(s.SwapTimeout().Seconds())),
		nonce,
		s.info.EthAsset,
		s.providedAmount.BigInt(),
	)
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to instantiate swap on-chain: %w", err)
	}

	log.Debugf("instantiated swap on-chain: amount=%s asset=%s txHash=%s", s.providedAmount, s.info.EthAsset, txHash)

	if len(receipt.Logs) == 0 {
		return ethcommon.Hash{}, errSwapInstantiationNoLogs
	}

	for _, rLog := range receipt.Logs {
		s.contractSwapID, err = contracts.GetIDFromLog(rLog)
		if err == nil {
			break
		}
	}
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("swap ID not found in transaction receipt's logs: %w", err)
	}

	var t0 *big.Int
	var t1 *big.Int
	for _, log := range receipt.Logs {
		t0, t1, err = contracts.GetTimeoutsFromLog(log)
		if err == nil {
			break
		}
	}
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("timeouts not found in transaction receipt's logs: %w", err)
	}

	s.fundsLocked = true
	s.setTimeouts(t0, t1)

	s.contractSwap = &contracts.SwapFactorySwap{
		Owner:        s.ETHClient().Address(),
		Claimer:      s.xmrmakerAddress,
		PubKeyClaim:  cmtXMRMaker,
		PubKeyRefund: cmtXMRTaker,
		Timeout0:     t0,
		Timeout1:     t1,
		Asset:        ethcommon.Address(s.info.EthAsset),
		Value:        s.providedAmount.BigInt(),
		Nonce:        nonce,
	}

	ethInfo := &db.EthereumSwapInfo{
		StartNumber:     receipt.BlockNumber,
		SwapID:          s.contractSwapID,
		Swap:            s.contractSwap,
		ContractAddress: s.Backend.ContractAddr(),
	}

	if err := s.Backend.RecoveryDB().PutContractSwapInfo(s.ID(), ethInfo); err != nil {
		return ethcommon.Hash{}, err
	}

	return txHash, nil
}

// ready calls the Ready() method on the Swap contract, indicating to XMRMaker he has until time t_1 to
// call Claim(). Ready() should only be called once XMRTaker sees XMRMaker lock his XMR.
// If time t_0 has passed, there is no point of calling Ready().
func (s *swapState) ready() error {
	stage, err := s.Contract().Swaps(s.ETHClient().CallOpts(s.ctx), s.contractSwapID)
	if err != nil {
		return err
	}

	if stage != contracts.StagePending {
		return fmt.Errorf("cannot set contract to ready when swap stage is %s", contracts.StageToString(stage))
	}

	txHash, receipt, err := s.sender.SetReady(s.contractSwap)
	if err != nil {
		if strings.Contains(err.Error(), revertSwapCompleted) && !s.info.Status.IsOngoing() {
			return nil
		}
		return err
	}

	log.Debugf("contract set to ready in block %d, tx %s", receipt.BlockNumber, txHash)
	return nil
}

// refund calls the Refund() method in the Swap contract, revealing XMRTaker's secret
// and returns to her the ether in the contract.
// If time t_1 passes and Claim() has not been called, XMRTaker should call Refund().
func (s *swapState) refund() (ethcommon.Hash, error) {
	sc := s.getSecret()

	log.Infof("attempting to call Refund()...")
	txHash, _, err := s.sender.Refund(s.contractSwap, sc)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	s.clearNextExpectedEvent(types.CompletedRefund)
	return txHash, nil
}

// generateKeys generates XMRTaker's monero spend and view keys (S_b, V_b), a secp256k1 public key,
// and a DLEq proof proving that the two keys correspond.
func generateKeys() (*pcommon.KeysAndProof, error) {
	return pcommon.GenerateKeysAndProof()
}

func generateNonce() *big.Int {
	u256PlusOne := new(big.Int).Lsh(big.NewInt(1), 256)
	maxU256 := new(big.Int).Sub(u256PlusOne, big.NewInt(1))
	n, _ := rand.Int(rand.Reader, maxU256)
	return n
}
