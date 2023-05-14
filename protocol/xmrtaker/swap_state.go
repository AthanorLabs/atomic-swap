// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package xmrtaker manages the swap state of individual swaps where the local swapd
// instance is offering Ethereum assets and accepting Monero in return.
package xmrtaker

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/libp2p/go-libp2p/core/peer"

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
	providedAmount coins.EthAssetAmount

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
	contractSwap   *contracts.SwapCreatorSwap
	t1, t2         time.Time

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
	// signals the t1 expiration handler to return
	xmrLockedCh chan struct{}
	// signals the t2 expiration handler to return
	claimedCh chan struct{}
	// signals to the creator xmrmaker instance that it can delete this swap
	done chan struct{}
}

func newSwapStateFromStart(
	b backend.Backend,
	makerPeerID peer.ID,
	offerID types.Hash,
	noTransferBack bool,
	providedAmount coins.EthAssetAmount,
	exchangeRate *coins.ExchangeRate,
	ethAsset types.EthAsset,
) (*swapState, error) {
	stage := types.ExpectingKeys

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

	expectedAmount, err := exchangeRate.ToXMR(providedAmount.AsStandard())
	if err != nil {
		return nil, err
	}

	info := pswap.NewInfo(
		makerPeerID,
		offerID,
		coins.ProvidesETH,
		providedAmount.AsStandard(),
		expectedAmount,
		exchangeRate,
		ethAsset,
		stage,
		moneroStartNumber,
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

	if err := s.generateAndSetKeys(); err != nil {
		return nil, err
	}

	s.SwapManager().PushNewStatus(offerID, stage)

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

	makerSk, makerVk, err := b.RecoveryDB().GetCounterpartySwapKeys(info.OfferID)
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

	if b.SwapCreatorAddr() != ethSwapInfo.SwapCreatorAddr {
		return nil, errContractAddrMismatch(ethSwapInfo.SwapCreatorAddr.String())
	}

	s.setTimeouts(ethSwapInfo.Swap.Timeout1, ethSwapInfo.Swap.Timeout2)
	s.privkeys = sk
	s.pubkeys = sk.PublicKeyPair()
	s.contractSwapID = ethSwapInfo.SwapID
	s.contractSwap = ethSwapInfo.Swap
	s.xmrmakerPublicSpendKey = makerSk
	s.xmrmakerPrivateViewKey = makerVk

	if info.Status == types.ETHLocked {
		go s.checkForXMRLock()
	}

	go s.runT1ExpirationHandler()
	go s.runT2ExpirationHandler()
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
	if info.EthAsset.IsToken() {
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
		b.SwapCreatorAddr(),
		ethStartNumber,
		claimedTopic,
		logClaimedCh,
	)

	err := claimedWatcher.Start()
	if err != nil {
		cancel()
		return nil, err
	}

	var providedAmt coins.EthAssetAmount
	if info.EthAsset.IsETH() {
		providedAmt = coins.EtherToWei(info.ProvidedAmount)
	} else {
		tokenInfo, err := b.ETHClient().ERC20Info(b.Ctx(), info.EthAsset.Address())
		if err != nil {
			cancel()
			return nil, err
		}
		providedAmt = coins.NewERC20TokenAmountFromDecimals(info.ProvidedAmount, tokenInfo)
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
		providedAmount:    providedAmt,
	}

	go s.runHandleEvents()
	go s.runContractEventWatcher()
	return s, nil
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

func (s *swapState) UpdateStatus(status types.Status) {
	s.info.SetStatus(status)
	s.SwapManager().PushNewStatus(s.OfferID(), status)
}

// ExpectedAmount returns the amount received, or expected to be received, at the end of the swap
func (s *swapState) ExpectedAmount() *apd.Decimal {
	return s.info.ExpectedAmount
}

func (s *swapState) expectedPiconeroAmount() *coins.PiconeroAmount {
	return coins.MoneroToPiconero(s.info.ExpectedAmount)
}

// OfferID returns the Offer ID of the swap
func (s *swapState) OfferID() types.Hash {
	return s.info.OfferID
}

// NotifyStreamClosed is called by the network when the swap stream closes.
func (s *swapState) NotifyStreamClosed() {
	switch s.nextExpectedEvent {
	case EventKeysReceivedType:
		// exit the swap, the remote peer closed the stream
		// before we received all expected messages
		err := s.Exit()
		if err != nil {
			log.Errorf("failed to exit swap: %s", err)
		}
	default:
		// do nothing, as we're not waiting for more network messages
	}
}

// Exit is called by the network when the protocol stream closes, or if the swap_refund RPC endpoint is called.
// It exists the swap by refunding if necessary. If no locking has been done, it simply aborts the swap.
// If the swap already completed successfully, this function does not do anything regarding the protocol.
func (s *swapState) Exit() error {
	event := newEventExit()
	s.eventCh <- event
	err := <-event.errCh
	if err != nil {
		log.Errorf("failed to exit swap: %s", err)
	}
	return err
}

// exit is the same as Exit, but assumes the calling code block already holds the swapState lock.
func (s *swapState) exit() error {
	defer func() {
		s.CloseProtocolStream(s.OfferID())

		err := s.SwapManager().CompleteOngoingSwap(s.info)
		if err != nil {
			log.Warnf("failed to mark swap %s as completed: %s", s.info.OfferID, err)
			return
		}

		// delete from network state
		s.Backend.DeleteOngoingSwap(s.OfferID())

		err = s.Backend.RecoveryDB().DeleteSwap(s.OfferID())
		if err != nil {
			log.Warnf("failed to delete temporary swap info %s from db: %s", s.OfferID(), err)
		}

		// Stop all per-swap goroutines
		s.cancel()
		close(s.done)

		var exitLog string
		switch s.info.Status {
		case types.CompletedSuccess:
			exitLog = color.New(color.Bold).Sprintf("**swap completed successfully: offerID=%s**", s.OfferID())
		case types.CompletedRefund:
			exitLog = color.New(color.Bold).Sprintf("**swap refunded successfully: offerID=%s**", s.OfferID())
		case types.CompletedAbort:
			exitLog = color.New(color.Bold).Sprintf("**swap aborted: id=%s**", s.OfferID())
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
		// for EventXMRLocked, we already locked our ETH-asset,
		// so we should call Refund().
		//
		// for EventETHClaimed, the XMR has been locked, but the
		// ETH hasn't been claimed, but the contract has been set to ready.
		// we should also refund in this case, since we might be past t2.
		receipt, err := s.tryRefund()
		if err != nil {
			if errors.Is(err, errRefundSwapCompleted) || strings.Contains(err.Error(), revertSwapCompleted) {
				log.Infof("swap was already completed")

				err = s.tryClaim()
				if err != nil {
					if errors.Is(err, errNoClaimLogsFound) {
						// in this case, assume we refunded
						s.clearNextExpectedEvent(types.CompletedRefund)
						return nil
					}

					// note: this should NOT occur; it could if the ethclient
					// or monero clients crash during the course of the claim,
					// but that would be very bad.
					return fmt.Errorf("failed to claim even though swap was completed on-chain: %w", err)
				}

				return nil
			}

			return fmt.Errorf("failed to refund: %w", err)
		}

		s.clearNextExpectedEvent(types.CompletedRefund)
		log.Infof("refunded ether: txID=%s", receipt.TxHash)
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

func (s *swapState) tryRefund() (*ethtypes.Receipt, error) {
	stage, err := s.SwapCreator().Swaps(s.ETHClient().CallOpts(s.ctx), s.contractSwapID)
	if err != nil {
		return nil, err
	}

	switch stage {
	case contracts.StageInvalid:
		return nil, fmt.Errorf("%w: contract swap ID: %s", errRefundInvalid, s.contractSwapID)
	case contracts.StageCompleted:
		return nil, errRefundSwapCompleted
	case contracts.StagePending, contracts.StageReady:
		// do nothing
	default:
		panic("Unhandled stage value")
	}

	isReady := stage == contracts.StageReady

	ts, err := s.ETHClient().LatestBlockTimestamp(s.ctx)
	if err != nil {
		return nil, err
	}

	log.Debugf("tryRefund isReady=%v untilT1=%vs untilT2=%vs",
		isReady, s.t1.Sub(ts).Seconds(), s.t2.Sub(ts).Seconds())

	if ts.Before(s.t1) && !isReady {
		receipt, err := s.refund() //nolint:govet
		// TODO: Have refund() return errors that we can use errors.Is to check against
		if err == nil {
			return receipt, nil
		}

		// There is a small, but non-zero chance that our transaction gets placed in a block that is after T1
		// even though the current block is before T1. In this case, the transaction will be reverted, the
		// gas fee is lost, but we can wait until T2 and try again.
		log.Warnf("first refund attempt failed: err=%s", err)
	}

	if ts.After(s.t2) {
		return s.refund()
	}

	// the contract is "ready", so we can't do anything until
	// the counterparty claims or until t2 passes.
	//
	// we let the runT2ExpirationHandler() routine continue to run and read
	// from s.eventCh for EventShouldRefund or EventETHClaimed.
	// (since this function is called from inside the event handler routine,
	// it won't handle those events while this function is executing.)
	log.Infof("waiting until time %s to refund", s.t2)

	waitCtx, waitCtxCancel := context.WithCancel(s.ctx)
	defer waitCtxCancel()

	waitCh := make(chan error)
	go func() {
		waitCh <- s.ETHClient().WaitForTimestamp(waitCtx, s.t2)
		close(waitCh)
	}()

	for {
		select {
		case event := <-s.eventCh:
			log.Debugf("got event %s while waiting for T2", event.Type())
			switch event.(type) {
			case *EventShouldRefund:
				return s.refund()
			case *EventETHClaimed:
				// we should claim; returning this error
				// causes the calling function to claim
				return nil, fmt.Errorf(revertSwapCompleted)
			case *EventExit:
				// do nothing, we're already exiting
			default:
				panic(fmt.Sprintf("got unexpected event while waiting for Claimed/T2: %s", event.Type()))
			}
		case err = <-waitCh:
			if err != nil {
				return nil, fmt.Errorf("failed to wait for T2: %w", err)
			}

			return s.refund()
		}
	}
}

func (s *swapState) setTimeouts(t1, t2 *big.Int) {
	s.t1 = time.Unix(t1.Int64(), 0)
	s.t2 = time.Unix(t2.Int64(), 0)
	s.info.Timeout1 = &s.t1
	s.info.Timeout2 = &s.t2
}

// generateAndSetKeys generates and sets the XMRTaker's monero spend and view keys (S_b, V_b), a secp256k1 public key,
// and a DLEq proof proving that the two keys correspond.
func (s *swapState) generateAndSetKeys() error {
	if s.privkeys != nil {
		panic("generateAndSetKeys should only be called once")
	}

	keysAndProof, err := pcommon.GenerateKeysAndProof()
	if err != nil {
		return err
	}

	s.dleqProof = keysAndProof.DLEqProof
	s.secp256k1Pub = keysAndProof.Secp256k1PublicKey
	s.privkeys = keysAndProof.PrivateKeyPair
	s.pubkeys = keysAndProof.PublicKeyPair

	return s.Backend.RecoveryDB().PutSwapPrivateKey(s.OfferID(), s.privkeys.SpendKey())
}

// getSecret secrets returns the current secret scalar used to unlock funds from the contract.
func (s *swapState) getSecret() [32]byte {
	if s.dleqProof == nil {
		// the EVM expects the bytes to be big endian, and the ed25519 lib uses little endian
		return [32]byte(common.Reverse(s.privkeys.SpendKey().Bytes()))
	}

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
	return s.Backend.RecoveryDB().PutCounterpartySwapKeys(s.info.OfferID, sk, vk)
}

// lockAsset calls the Swap contract function new_swap and locks `amount` ether in it.
func (s *swapState) lockAsset() (*ethtypes.Receipt, error) {
	if s.xmrmakerPublicSpendKey == nil || s.xmrmakerPrivateViewKey == nil {
		panic(errCounterpartyKeysNotSet)
	}

	cmtXMRTaker := s.secp256k1Pub.Keccak256()
	cmtXMRMaker := s.xmrmakerSecp256k1PublicKey.Keccak256()
	providedAmt := s.providedAmount

	log.Debugf("locking %s %s in contract", providedAmt.AsStandard(), providedAmt.StandardSymbol())

	nonce := contracts.GenerateNewSwapNonce()
	receipt, err := s.sender.NewSwap(
		cmtXMRMaker,
		cmtXMRTaker,
		s.xmrmakerAddress,
		big.NewInt(int64(s.SwapTimeout().Seconds())),
		nonce,
		providedAmt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate swap on-chain: %w", err)
	}

	log.Infof("instantiated swap on-chain: amount=%s asset=%s %s",
		s.providedAmount, s.info.EthAsset, common.ReceiptInfo(receipt))

	if len(receipt.Logs) == 0 {
		return nil, errSwapInstantiationNoLogs
	}

	for _, rLog := range receipt.Logs {
		s.contractSwapID, err = contracts.GetIDFromLog(rLog)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("swap ID not found in transaction receipt's logs: %w", err)
	}

	var t1 *big.Int
	var t2 *big.Int
	for _, log := range receipt.Logs {
		t1, t2, err = contracts.GetTimeoutsFromLog(log)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("timeouts not found in transaction receipt's logs: %w", err)
	}

	s.fundsLocked = true
	s.setTimeouts(t1, t2)

	s.contractSwap = &contracts.SwapCreatorSwap{
		Owner:        s.ETHClient().Address(),
		Claimer:      s.xmrmakerAddress,
		PubKeyClaim:  cmtXMRMaker,
		PubKeyRefund: cmtXMRTaker,
		Timeout1:     t1,
		Timeout2:     t2,
		Asset:        ethcommon.Address(s.info.EthAsset),
		Value:        s.providedAmount.BigInt(),
		Nonce:        nonce,
	}

	ethInfo := &db.EthereumSwapInfo{
		StartNumber:     receipt.BlockNumber,
		SwapID:          s.contractSwapID,
		Swap:            s.contractSwap,
		SwapCreatorAddr: s.Backend.SwapCreatorAddr(),
	}

	if err := s.Backend.RecoveryDB().PutContractSwapInfo(s.OfferID(), ethInfo); err != nil {
		return nil, err
	}

	log.Infof("locked %s in swap contract, waiting for XMR to be locked", providedAmt.StandardSymbol())
	return receipt, nil
}

// ready calls the Ready() method on the Swap contract, indicating to XMRMaker he has until time t_1 to
// call Claim(). Ready() should only be called once XMRTaker sees XMRMaker lock his XMR.
// If time t_0 has passed, there is no point of calling Ready().
func (s *swapState) ready() error {
	stage, err := s.SwapCreator().Swaps(s.ETHClient().CallOpts(s.ctx), s.contractSwapID)
	if err != nil {
		return err
	}

	if stage != contracts.StagePending {
		if stage == contracts.StageReady {
			log.Warnf("contract already set to ready, ignoring call to ready()")
			return nil
		}

		if stage == contracts.StageCompleted {
			log.Infof("contract aleady set to completed, ignoring call to ready() and sending EventExit")
			go func() {
				err = s.Exit()
				if err != nil {
					log.Errorf("failed to handle EventExit: %s", err)
				}
			}()
			return nil
		}

		return fmt.Errorf("cannot set contract to ready when swap stage is %s", contracts.StageToString(stage))
	}

	receipt, err := s.sender.SetReady(s.contractSwap)
	if err != nil {
		if strings.Contains(err.Error(), revertSwapCompleted) && !s.info.Status.IsOngoing() {
			return nil
		}
		return err
	}

	log.Infof("contract set to ready %s", common.ReceiptInfo(receipt))

	return nil
}

// refund calls the Refund() method in the Swap contract, revealing XMRTaker's secret
// and returns to her the ether in the contract.
// If time t_1 passes and Claim() has not been called, XMRTaker should call Refund().
func (s *swapState) refund() (*ethtypes.Receipt, error) {
	sc := s.getSecret()

	log.Infof("attempting to call Refund()...")
	receipt, err := s.sender.Refund(s.contractSwap, sc)
	if err != nil {
		return nil, err
	}
	log.Infof("refund succeeded %s", common.ReceiptInfo(receipt))

	s.clearNextExpectedEvent(types.CompletedRefund)
	return receipt, nil
}
