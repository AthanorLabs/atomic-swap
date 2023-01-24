// Package xmrmaker manages the swap state of individual swaps where the local swapd
// instance is offering Monero and accepting Ethereum assets in return.
package xmrmaker

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/fatih/color"

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
	"github.com/athanorlabs/atomic-swap/protocol/xmrmaker/offers"
)

var (
	readyTopic    = common.GetTopic(common.ReadyEventSignature)
	refundedTopic = common.GetTopic(common.RefundedEventSignature)
)

type swapState struct {
	backend.Backend
	sender txsender.Sender

	ctx    context.Context
	cancel context.CancelFunc

	info         *pswap.Info
	offer        *types.Offer
	offerExtra   *types.OfferExtra
	offerManager *offers.Manager

	// our keys for this session
	dleqProof    *dleq.Proof
	secp256k1Pub *secp256k1.PublicKey
	privkeys     *mcrypto.PrivateKeyPair
	pubkeys      *mcrypto.PublicKeyPair

	// swap contract and timeouts in it
	contract       *contracts.SwapFactory
	contractAddr   ethcommon.Address
	contractSwapID [32]byte
	contractSwap   contracts.SwapFactorySwap
	t0, t1         time.Time

	// XMRTaker's keys for this session
	xmrtakerPublicKeys         *mcrypto.PublicKeyPair
	xmrtakerSecp256K1PublicKey *secp256k1.PublicKey
	moneroStartHeight          uint64 // height of the monero blockchain when the swap is started

	// tracks the state of the swap
	nextExpectedEvent EventType

	// channels

	// channel for swap events
	// the event handler in event.go ensures only one event is being handled at a time
	eventCh chan Event
	// channel for `Ready` logs seen on-chain
	logReadyCh chan ethtypes.Log
	// channel for `Refunded` logs seen on-chain
	logRefundedCh chan ethtypes.Log
	// signals the t0 expiration handler to return
	readyCh chan struct{}
	// signals to the creator xmrmaker instance that it can delete this swap
	done chan struct{}
}

// newSwapStateFromStart returns a new *swapState for a fresh swap.
func newSwapStateFromStart(
	b backend.Backend,
	offer *types.Offer,
	offerExtra *types.OfferExtra,
	om *offers.Manager,
	providesAmount *coins.PiconeroAmount,
	desiredAmount EthereumAssetAmount,
) (*swapState, error) {
	// at this point, we've received the counterparty's keys,
	// and will send our own after this function returns.
	// see HandleInitiateMessage().
	stage := types.KeysExchanged
	if offerExtra.StatusCh == nil {
		offerExtra.StatusCh = make(chan types.Status, 7)
	}

	if offerExtra.RelayerEndpoint != "" {
		if err := b.RecoveryDB().PutSwapRelayerInfo(offer.ID, offerExtra); err != nil {
			return nil, err
		}
	}

	moneroStartHeight, err := b.XMRClient().GetChainHeight()
	if err != nil {
		return nil, err
	}
	// reduce the scan height a little in case there is a block reorg
	if moneroStartHeight >= monero.MinSpendConfirmations {
		moneroStartHeight -= monero.MinSpendConfirmations
	}

	ethHeader, err := b.ETHClient().Raw().HeaderByNumber(b.Ctx(), nil)
	if err != nil {
		return nil, err
	}

	info := pswap.NewInfo(
		offer.ID,
		coins.ProvidesXMR,
		providesAmount.AsMonero(),
		desiredAmount.AsStandard(),
		offer.ExchangeRate,
		offer.EthAsset,
		stage,
		moneroStartHeight,
		offerExtra.StatusCh,
	)

	if err = b.SwapManager().AddSwap(info); err != nil {
		return nil, err
	}

	s, err := newSwapState(
		b,
		offer,
		offerExtra,
		om,
		ethHeader.Number,
		moneroStartHeight,
		info,
	)
	if err != nil {
		return nil, err
	}

	err = s.generateAndSetKeys()
	if err != nil {
		return nil, err
	}

	offerExtra.StatusCh <- stage
	return s, nil
}

// newSwapStateFromOngoing returns a new *swapState given information about a swap
// that's ongoing, but not yet completed.
func newSwapStateFromOngoing(
	b backend.Backend,
	offer *types.Offer,
	offerExtra *types.OfferExtra,
	om *offers.Manager,
	ethSwapInfo *db.EthereumSwapInfo,
	info *pswap.Info,
	sk *mcrypto.PrivateKeyPair,
) (*swapState, error) {
	// TODO: do we want to support the case where the ETH has been locked,
	// but we haven't locked yet?
	if info.Status != types.XMRLocked {
		return nil, errInvalidStageForRecovery
	}

	s, err := newSwapState(
		b, offer, offerExtra, om, ethSwapInfo.StartNumber, info.MoneroStartHeight, info,
	)
	if err != nil {
		return nil, err
	}

	err = s.setContract(ethSwapInfo.ContractAddress)
	if err != nil {
		return nil, err
	}

	s.setTimeouts(ethSwapInfo.Swap.Timeout0, ethSwapInfo.Swap.Timeout1)
	s.privkeys = sk
	s.pubkeys = sk.PublicKeyPair()
	s.contractSwapID = ethSwapInfo.SwapID
	s.contractSwap = ethSwapInfo.Swap
	return s, nil
}

func newSwapState(
	b backend.Backend,
	offer *types.Offer,
	offerExtra *types.OfferExtra,
	om *offers.Manager,
	ethStartNumber *big.Int,
	moneroStartNumber uint64,
	info *pswap.Info,
) (*swapState, error) {
	var sender txsender.Sender
	if offer.EthAsset != types.EthAssetETH {
		erc20Contract, err := contracts.NewIERC20(offer.EthAsset.Address(), b.ETHClient().Raw())
		if err != nil {
			return nil, err
		}

		sender, err = b.NewTxSender(offer.EthAsset.Address(), erc20Contract)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		sender, err = b.NewTxSender(offer.EthAsset.Address(), nil)
		if err != nil {
			return nil, err
		}
	}

	// set up ethereum event watchers
	const logChSize = 16 // arbitrary, we just don't want the watcher to block on writing
	logReadyCh := make(chan ethtypes.Log, logChSize)
	logRefundedCh := make(chan ethtypes.Log, logChSize)

	// Create per swap context that is canceled when the swap completes
	ctx, cancel := context.WithCancel(b.Ctx())

	readyWatcher := watcher.NewEventFilter(
		ctx,
		b.ETHClient().Raw(),
		b.ContractAddr(),
		ethStartNumber,
		readyTopic,
		logReadyCh,
	)

	refundedWatcher := watcher.NewEventFilter(
		ctx,
		b.ETHClient().Raw(),
		b.ContractAddr(),
		ethStartNumber,
		refundedTopic,
		logRefundedCh,
	)

	err := readyWatcher.Start()
	if err != nil {
		cancel()
		return nil, err
	}

	err = refundedWatcher.Start()
	if err != nil {
		cancel()
		return nil, err
	}

	// note: if this is recovering an ongoing swap, this will only
	// be invoked if our status is XMRLocked; ie. we've locked XMR,
	// but not yet claimed or refunded.
	//
	// dleqProof and secp256k1Pub are never set, as they are only used
	// in the swap steps before XMR is locked.
	//
	// similarly, xmrtakerPublicKeys and xmrtakerSecp256K1PublicKey are
	// also never set, as they're only used to check the contract
	// before we lock XMR.
	s := &swapState{
		ctx:               ctx,
		cancel:            cancel,
		Backend:           b,
		sender:            sender,
		offer:             offer,
		offerExtra:        offerExtra,
		offerManager:      om,
		moneroStartHeight: moneroStartNumber,
		nextExpectedEvent: nextExpectedEventFromStatus(info.Status),
		logReadyCh:        logReadyCh,
		logRefundedCh:     logRefundedCh,
		eventCh:           make(chan Event, 1),
		readyCh:           make(chan struct{}),
		info:              info,
		done:              make(chan struct{}),
	}

	go s.runHandleEvents()
	go s.runContractEventWatcher()
	return s, nil
}

// SendKeysMessage ...
func (s *swapState) SendKeysMessage() *message.SendKeysMessage {
	return &message.SendKeysMessage{
		ProvidedAmount:     s.info.ProvidedAmount,
		PublicSpendKey:     s.pubkeys.SpendKey().Hex(),
		PrivateViewKey:     s.privkeys.ViewKey().Hex(),
		DLEqProof:          hex.EncodeToString(s.dleqProof.Proof()),
		Secp256k1PublicKey: s.secp256k1Pub.String(),
		EthAddress:         s.ETHClient().Address().String(),
	}
}

// ExpectedAmount returns the amount received, or expected to be received, at the end of the swap
func (s *swapState) ExpectedAmount() *apd.Decimal {
	return s.info.ExpectedAmount
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
	log.Debugf("attempting to exit swap: nextExpectedEvent=%v", s.nextExpectedEvent)

	defer func() {
		err := s.SwapManager().CompleteOngoingSwap(s.info)
		if err != nil {
			log.Warnf("failed to mark swap %s as completed: %s", s.offer.ID, err)
			return
		}

		log.Infof("exit status %s", s.info.Status)

		if s.info.Status != types.CompletedSuccess && s.offer.IsSet() {
			// re-add offer, as it wasn't taken successfully
			_, err = s.offerManager.AddOffer(s.offer, s.offerExtra.RelayerEndpoint, s.offerExtra.RelayerCommission)
			if err != nil {
				log.Warnf("failed to re-add offer %s: %s", s.offer.ID, err)
			}

			log.Debugf("re-added offer %s", s.offer.ID)
		} else if s.info.Status == types.CompletedSuccess {
			err = s.offerManager.DeleteOfferFromDB(s.offer.ID)
			if err != nil {
				log.Warnf("failed to delete offer %s from db: %s", s.offer.ID, err)
			}
		}

		err = s.Backend.RecoveryDB().DeleteSwap(s.offer.ID)
		if err != nil {
			log.Warnf("failed to delete temporary swap info %s from db: %s", s.offer.ID, err)
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

	switch s.nextExpectedEvent {
	case EventETHLockedType:
		// we were waiting for the contract to be deployed, but haven't
		// locked out funds yet, so we're fine.
		s.clearNextExpectedEvent(types.CompletedAbort)
		return nil
	case EventContractReadyType:
		// this case takes control of the event channel.
		// the next event will either be EventContractReady or EventETHRefunded.

		var err error
		event := <-s.eventCh
		switch e := event.(type) {
		case *EventETHRefunded:
			err = s.handleEventETHRefunded(e)
		case *EventContractReady:
			err = s.handleEventContractReady()
		}
		if err != nil {
			return err
		}

		return nil
	case EventNoneType:
		// we already completed the swap, do nothing
		return nil
	default:
		s.clearNextExpectedEvent(types.CompletedAbort)
		log.Errorf("unexpected nextExpectedEvent in Exit: type=%s", s.nextExpectedEvent)
		return errUnexpectedMessageType
	}
}

// TODO: Unit test this
func sweepRefundBack(
	ctx context.Context,
	primaryCli monero.WalletClient,
	abPrivKeyPair *mcrypto.PrivateKeyPair,
	restoreHeight uint64,
) error {
	conf := primaryCli.CreateABWalletConf("xmrmaker-swap-wallet-refund")
	abWalletCli, err := monero.CreateSpendWalletFromKeys(conf, abPrivKeyPair, restoreHeight)
	if err != nil {
		return err
	}
	defer abWalletCli.CloseAndRemoveWallet()
	balance, err := abWalletCli.GetBalance(0)
	if err != nil {
		return err
	}
	if balance.BlocksToUnlock > 0 {
		if _, err = monero.WaitForBlocks(ctx, abWalletCli, int(balance.BlocksToUnlock)); err != nil {
			return err
		}
	}
	log.Infof("Sweeping refund of %s XMR back to primary address %s",
		coins.FmtPiconeroAmtAsXMR(balance.Balance), primaryCli.PrimaryAddress())
	sweepResp, err := abWalletCli.SweepAll(primaryCli.PrimaryAddress(), 0)
	if err != nil {
		return err
	}
	if len(sweepResp.TxHashList) < 1 {
		// this shouldn't be possible, but it is not our code that sent the response
		return errors.New("received invalid monero sweep response with no TX hashes")
	}
	transfer, err := abWalletCli.WaitForReceipt(&monero.WaitForReceiptRequest{
		Ctx:              context.Background(),
		TxID:             sweepResp.TxHashList[0],
		NumConfirmations: 2,
		AccountIdx:       0,
	})
	if err != nil {
		return err
	}
	log.Infof("XMRMaker swept refund of %s XMR back to primary wallet, but %s XMR was lost to fees",
		coins.FmtPiconeroAmtAsXMR(balance.Balance), coins.FmtPiconeroAmtAsXMR(transfer.Fee))
	return nil
}

func (s *swapState) reclaimMonero(skA *mcrypto.PrivateSpendKey) (mcrypto.Address, error) {
	vkA, err := skA.View()
	if err != nil {
		return "", err
	}

	skAB := mcrypto.SumPrivateSpendKeys(skA, s.privkeys.SpendKey())
	vkAB := mcrypto.SumPrivateViewKeys(vkA, s.privkeys.ViewKey())
	kpAB := mcrypto.NewPrivateKeyPair(skAB, vkAB)

	// write keys to file in case something goes wrong
	if err = s.Backend.RecoveryDB().PutSharedSwapPrivateKey(s.ID(), kpAB.SpendKey()); err != nil {
		return "", err
	}

	if err = sweepRefundBack(s.ctx, s.XMRClient(), kpAB, s.moneroStartHeight); err != nil {
		return "", fmt.Errorf("failed to sweep refund back to primary wallet: %w", err)
	}
	return kpAB.Address(s.Env()), nil
}

// generateKeys generates XMRMaker's spend and view keys (s_b, v_b)
// It returns XMRMaker's public spend key and his private view key, so that XMRTaker can see
// if the funds are locked.
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

func generateKeys() (*pcommon.KeysAndProof, error) {
	return pcommon.GenerateKeysAndProof()
}

// getSecret secrets returns the current secret scalar used to unlock funds from the contract.
func (s *swapState) getSecret() [32]byte {
	if s.dleqProof != nil {
		return s.dleqProof.Secret()
	}

	var secret [32]byte
	copy(secret[:], common.Reverse(s.privkeys.SpendKey().Bytes()))
	return secret
}

// setXMRTakerPublicKeys sets XMRTaker's public spend and view keys
func (s *swapState) setXMRTakerPublicKeys(sk *mcrypto.PublicKeyPair, secp256k1Pub *secp256k1.PublicKey) {
	s.xmrtakerPublicKeys = sk
	s.xmrtakerSecp256K1PublicKey = secp256k1Pub
}

// setContract sets the contract in which XMRTaker has locked her ETH.
func (s *swapState) setContract(address ethcommon.Address) error {
	s.contractAddr = address

	var err error
	s.contract, err = s.NewSwapFactory(address)
	if err != nil {
		return err
	}

	s.sender.SetContractAddress(address)
	s.sender.SetContract(s.contract)
	return nil
}

// lockFunds locks XMRMaker's funds in the monero account specified by public key
// (S_a + S_b), viewable with (V_a + V_b)
// It accepts the amount to lock as the input
func (s *swapState) lockFunds(amount *coins.PiconeroAmount) (*message.NotifyXMRLock, error) {
	swapDestAddr := mcrypto.SumSpendAndViewKeys(s.xmrtakerPublicKeys, s.pubkeys).Address(s.Env())
	log.Infof("going to lock XMR funds, amount=%s XMR", amount.AsMoneroString())

	balance, err := s.XMRClient().GetBalance(0)
	if err != nil {
		return nil, err
	}

	log.Debug("total XMR balance: ", coins.FmtPiconeroAmtAsXMR(balance.Balance))
	log.Info("unlocked XMR balance: ", coins.FmtPiconeroAmtAsXMR(balance.UnlockedBalance))

	log.Infof("Starting lock of %s XMR in address %s", amount.AsMoneroString(), swapDestAddr)
	transfer, err := s.XMRClient().Transfer(s.ctx, swapDestAddr, 0, amount, monero.MinSpendConfirmations)
	if err != nil {
		return nil, err
	}
	log.Infof("Successfully locked XMR funds: txID=%s address=%s block=%d",
		transfer.TxID, swapDestAddr, transfer.Height)
	return &message.NotifyXMRLock{
		Address: string(swapDestAddr),
		TxID:    transfer.TxID,
	}, nil
}
