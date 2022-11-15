package xmrmaker

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	eth "github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/fatih/color" //nolint:misspell

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/crypto/secp256k1"
	"github.com/athanorlabs/atomic-swap/dleq"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/watcher"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net"
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
	walletScanHeight           uint64 // height of the monero blockchain when the swap is started

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

func newSwapState(
	b backend.Backend,
	offer *types.Offer,
	offerExtra *types.OfferExtra,
	om *offers.Manager,
	providesAmount common.MoneroAmount,
	desiredAmount EthereumAssetAmount,
) (*swapState, error) {
	exchangeRate := types.ExchangeRate(providesAmount.AsMonero() / desiredAmount.AsStandard())

	stage := types.ExpectingKeys
	if offerExtra.StatusCh == nil {
		offerExtra.StatusCh = make(chan types.Status, 7)
	}

	offerExtra.StatusCh <- stage
	info := pswap.NewInfo(
		offer.ID,
		types.ProvidesXMR,
		providesAmount.AsMonero(),
		desiredAmount.AsStandard(),
		exchangeRate,
		offer.EthAsset,
		stage,
		offerExtra.StatusCh,
	)

	if err := b.SwapManager().AddSwap(info); err != nil {
		return nil, err
	}

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

	walletScanHeight, err := b.XMRClient().GetChainHeight()
	if err != nil {
		return nil, err
	}
	// reduce the scan height a little in case there is a block reorg
	if walletScanHeight >= monero.MinSpendConfirmations {
		walletScanHeight -= monero.MinSpendConfirmations
	}

	ethHeader, err := b.ETHClient().Raw().HeaderByNumber(b.Ctx(), nil)
	if err != nil {
		return nil, err
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
		ethHeader.Number,
		readyTopic,
		logReadyCh,
	)

	refundedWatcher := watcher.NewEventFilter(
		ctx,
		b.ETHClient().Raw(),
		b.ContractAddr(),
		ethHeader.Number,
		refundedTopic,
		logRefundedCh,
	)

	err = readyWatcher.Start()
	if err != nil {
		cancel()
		return nil, err
	}

	err = refundedWatcher.Start()
	if err != nil {
		cancel()
		return nil, err
	}

	s := &swapState{
		ctx:               ctx,
		cancel:            cancel,
		Backend:           b,
		sender:            sender,
		offer:             offer,
		offerExtra:        offerExtra,
		offerManager:      om,
		walletScanHeight:  walletScanHeight,
		nextExpectedEvent: EventETHLockedType,
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
func (s *swapState) SendKeysMessage() (*net.SendKeysMessage, error) {
	if err := s.generateAndSetKeys(); err != nil {
		return nil, err
	}

	return &net.SendKeysMessage{
		ProvidedAmount:     s.info.ProvidedAmount,
		PublicSpendKey:     s.pubkeys.SpendKey().Hex(),
		PrivateViewKey:     s.privkeys.ViewKey().Hex(),
		DLEqProof:          hex.EncodeToString(s.dleqProof.Proof()),
		Secp256k1PublicKey: s.secp256k1Pub.String(),
		EthAddress:         s.ETHClient().Address().String(),
	}, nil
}

// InfoFile returns the swap's infoFile path
func (s *swapState) InfoFile() string {
	return s.offerExtra.InfoFile
}

// ReceivedAmount returns the amount received, or expected to be received, at the end of the swap
func (s *swapState) ReceivedAmount() float64 {
	return s.info.ReceivedAmount
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
		err := s.SwapManager().CompleteOngoingSwap(s.offer.ID)
		if err != nil {
			log.Warnf("failed to mark swap %s as completed: %s", s.offer.ID, err)
			return
		}

		// TODO: when recovery from disk is implemented, remove s.offer != nil as
		// it should always be set
		if s.info.Status != types.CompletedSuccess && s.offer.IsSet() {
			// re-add offer, as it wasn't taken successfully
			_, err := s.offerManager.AddOffer(s.offer, s.offerExtra.RelayerEndpoint, s.offerExtra.RelayerCommission)
			if err != nil {
				log.Warnf("failed to re-add offer %s: %s", s.offer.ID, err)
			}

			log.Debugf("re-added offer %s", s.offer.ID)
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

func (s *swapState) reclaimMonero(skA *mcrypto.PrivateSpendKey) (mcrypto.Address, error) {
	vkA, err := skA.View()
	if err != nil {
		return "", err
	}

	skAB := mcrypto.SumPrivateSpendKeys(skA, s.privkeys.SpendKey())
	vkAB := mcrypto.SumPrivateViewKeys(vkA, s.privkeys.ViewKey())
	kpAB := mcrypto.NewPrivateKeyPair(skAB, vkAB)

	// write keys to file in case something goes wrong
	if err = pcommon.WriteSharedSwapKeyPairToFile(s.offerExtra.InfoFile, kpAB, s.Env()); err != nil {
		return "", err
	}

	s.XMRClient().Lock()
	defer s.XMRClient().Unlock()
	return monero.CreateWallet("xmrmaker-swap-wallet", s.Env(), s.XMRClient(), kpAB, s.walletScanHeight)
}

func (s *swapState) filterForRefund() (*mcrypto.PrivateSpendKey, error) {
	const refundedEvent = "Refunded"

	logs, err := s.ETHClient().Raw().FilterLogs(s.ctx, eth.FilterQuery{
		Addresses: []ethcommon.Address{s.ContractAddr()},
		Topics:    [][]ethcommon.Hash{{refundedTopic}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %w", err)
	}

	if len(logs) == 0 {
		return nil, errNoRefundLogsFound
	}

	var (
		foundLog ethtypes.Log
		found    bool
	)

	for _, log := range logs {
		matches, err := contracts.CheckIfLogIDMatches(log, refundedEvent, s.contractSwapID) //nolint:govet
		if err != nil {
			continue
		}

		if matches {
			foundLog = log
			found = true
			break
		}
	}

	if !found {
		return nil, errNoRefundLogsFound
	}

	sa, err := contracts.GetSecretFromLog(&foundLog, refundedEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret from log: %w", err)
	}

	return sa, nil
}

// generateKeys generates XMRMaker's spend and view keys (s_b, v_b)
// It returns XMRMaker's public spend key and his private view key, so that XMRTaker can see
// if the funds are locked.
func (s *swapState) generateAndSetKeys() error {
	if s == nil {
		return errNilSwapState
	}

	if s.privkeys != nil {
		return nil
	}

	keysAndProof, err := generateKeys()
	if err != nil {
		return err
	}

	s.dleqProof = keysAndProof.DLEqProof
	s.secp256k1Pub = keysAndProof.Secp256k1PublicKey
	s.privkeys = keysAndProof.PrivateKeyPair
	s.pubkeys = keysAndProof.PublicKeyPair

	return pcommon.WriteKeysToFile(s.offerExtra.InfoFile, s.privkeys, s.Env())
}

func generateKeys() (*pcommon.KeysAndProof, error) {
	return pcommon.GenerateKeysAndProof()
}

// getSecret secrets returns the current secret scalar used to unlock funds from the contract.
func (s *swapState) getSecret() [32]byte {
	return s.dleqProof.Secret()
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

func (s *swapState) setTimeouts(t0, t1 *big.Int) {
	s.t0 = time.Unix(t0.Int64(), 0)
	s.t1 = time.Unix(t1.Int64(), 0)
}

// checkContract checks the contract's balance and Claim/Refund keys.
// if the balance doesn't match what we're expecting to receive, or the public keys in the contract
// aren't what we expect, we error and abort the swap.
func (s *swapState) checkContract(txHash ethcommon.Hash) error {
	tx, _, err := s.ETHClient().Raw().TransactionByHash(s.ctx, txHash)
	if err != nil {
		return err
	}

	if tx.To() == nil || *(tx.To()) != s.contractAddr {
		return errInvalidETHLockedTransaction
	}

	receipt, err := s.ETHClient().WaitForReceipt(s.ctx, txHash)
	if err != nil {
		return fmt.Errorf("failed to get receipt for New transaction: %w", err)
	}

	if receipt.Status == 0 {
		// swap transaction reverted
		return errLockTxReverted
	}

	// check that New log was emitted
	if len(receipt.Logs) == 0 {
		return errCannotFindNewLog
	}

	var event *contracts.SwapFactoryNew
	for _, log := range receipt.Logs {
		event, err = s.Contract().ParseNew(*log)
		if err == nil {
			break
		}
	}
	if err != nil {
		return errCannotFindNewLog
	}

	if !bytes.Equal(event.SwapID[:], s.contractSwapID[:]) {
		return errUnexpectedSwapID
	}

	// check that contract was constructed with correct secp256k1 keys
	skOurs := s.secp256k1Pub.Keccak256()
	if !bytes.Equal(event.ClaimKey[:], skOurs[:]) {
		return fmt.Errorf("contract claim key is not expected: got 0x%x, expected 0x%x", event.ClaimKey, skOurs)
	}

	skTheirs := s.xmrtakerSecp256K1PublicKey.Keccak256()
	if !bytes.Equal(event.RefundKey[:], skTheirs[:]) {
		return fmt.Errorf("contract refund key is not expected: got 0x%x, expected 0x%x", event.RefundKey, skTheirs)
	}

	// TODO: check timeouts (#161)

	// check asset of created swap
	if types.EthAsset(s.contractSwap.Asset) != types.EthAsset(event.Asset) {
		return fmt.Errorf("swap asset is not expected: got %v, expected %v", event.Asset, s.contractSwap.Asset)
	}

	// check value of created swap
	if s.contractSwap.Value.Cmp(event.Value) != 0 {
		// this should never happen
		return fmt.Errorf("swap value and event value don't match: got %v, expected %v", event.Value, s.contractSwap.Value)
	}

	var receivedAmount *big.Int
	if s.info.EthAsset != types.EthAssetETH {
		_, _, decimals, err := s.ETHClient().ERC20Info(s.ctx, s.contractSwap.Asset)
		if err != nil {
			return fmt.Errorf("failed to get ERC20 info: %w", err)
		}

		receivedAmount = common.NewERC20TokenAmountFromDecimals(s.info.ReceivedAmount, float64(decimals)).BigInt()
	} else {
		receivedAmount = common.EtherToWei(s.info.ReceivedAmount).BigInt()
	}
	if s.contractSwap.Value.Cmp(receivedAmount) != 0 {
		return fmt.Errorf("swap value is not expected: got %v, expected %v", s.contractSwap.Value, receivedAmount)
	}

	return nil
}

// lockFunds locks XMRMaker's funds in the monero account specified by public key
// (S_a + S_b), viewable with (V_a + V_b)
// It accepts the amount to lock as the input
func (s *swapState) lockFunds(amount common.MoneroAmount) (*message.NotifyXMRLock, error) {
	swapDestAddr := mcrypto.SumSpendAndViewKeys(s.xmrtakerPublicKeys, s.pubkeys).Address(s.Env())
	log.Infof("going to lock XMR funds, amount(piconero)=%d", amount)

	s.XMRClient().Lock()
	defer s.XMRClient().Unlock()

	balance, err := s.XMRClient().GetBalance(0)
	if err != nil {
		return nil, err
	}

	log.Debug("total XMR balance: ", balance.Balance)
	log.Info("unlocked XMR balance: ", balance.UnlockedBalance)

	transResp, err := s.XMRClient().Transfer(swapDestAddr, 0, uint64(amount))
	if err != nil {
		return nil, err
	}

	log.Infof("locked %f XMR, txID=%s fee=%d", amount.AsMonero(), transResp.TxHash, transResp.Fee)

	// TODO: It would be friendlier to concurrent swaps if we didn't hold the client lock
	//       for the entire confirmation period. Options to improve this include creating a
	//       separate monero-wallet-rpc instance for A+B wallets or carefully releasing the
	//       lock between confirmations and re-opening the A+B wallet after grabbing the
	//       lock again.
	transfer, err := s.XMRClient().WaitForReceipt(&monero.WaitForReceiptRequest{
		Ctx:              s.ctx,
		TxID:             transResp.TxHash,
		DestAddr:         swapDestAddr,
		NumConfirmations: monero.MinSpendConfirmations,
		AccountIdx:       0,
	})
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
