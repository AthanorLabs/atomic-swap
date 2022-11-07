package xmrtaker

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/crypto/secp256k1"
	"github.com/athanorlabs/atomic-swap/dleq"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/watcher"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	pswap "github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/txsender"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/fatih/color" //nolint:misspell
)

const revertSwapCompleted = "swap is already completed"
const revertUnableToRefund = "it's the counterparty's turn, unable to refund, try again later"

// swapState is an instance of a swap. it holds the info needed for the swap,
// and its current state.
type swapState struct {
	backend.Backend
	sender txsender.Sender

	ctx          context.Context
	cancel       context.CancelFunc
	infoFile     string
	transferBack bool

	info     *pswap.Info
	statusCh chan types.Status

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

	// ETH asset being swapped
	ethAsset types.EthAsset

	// swap contract and timeouts in it; set once contract is deployed
	contractSwapID [32]byte
	contractSwap   contracts.SwapFactorySwap
	t0, t1         time.Time

	// tracks the state of the swap
	nextExpectedEvent Event

	// channels

	// channel for swap events
	// the event handler in event.go ensures only one event is being handled at a time
	eventCh chan Event
	// channel for `Claimed` logs seen on-chain
	logClaimedCh chan []ethtypes.Log
	// signals the t0 expiration handler to return
	xmrLockedCh chan struct{}
	// signals the t1 expiration handler to return
	claimedCh chan struct{}
	// signals to the creator xmrmaker instance that it can delete this swap
	done chan struct{}
}

func newSwapState(b backend.Backend, offerID types.Hash, infofile string, transferBack bool,
	providesAmount common.EtherAmount, receivedAmount common.MoneroAmount,
	exchangeRate types.ExchangeRate, ethAsset types.EthAsset) (*swapState, error) {
	if b.Contract() == nil {
		return nil, errNoSwapContractSet
	}

	_, err := b.XMRDepositAddress(nil)
	if transferBack && err != nil {
		return nil, errMustProvideWalletAddress
	}

	stage := types.ExpectingKeys
	statusCh := make(chan types.Status, 16)
	statusCh <- stage
	info := pswap.NewInfo(
		offerID,
		types.ProvidesETH,
		providesAmount.AsEther(),
		receivedAmount.AsMonero(),
		exchangeRate,
		ethAsset,
		stage,
		statusCh,
	)
	if err = b.SwapManager().AddSwap(info); err != nil {
		return nil, err
	}

	if !b.HasEthereumPrivateKey() {
		transferBack = true // front-end must set final deposit address
	}

	var sender txsender.Sender
	if ethAsset != types.EthAssetETH {
		erc20Contract, err := contracts.NewIERC20(ethAsset.Address(), b.EthClient()) //nolint:govet
		if err != nil {
			return nil, err
		}

		sender, err = b.NewTxSender(ethAsset.Address(), erc20Contract)
		if err != nil {
			return nil, err
		}
	} else {
		sender, err = b.NewTxSender(ethAsset.Address(), nil)
		if err != nil {
			return nil, err
		}
	}

	walletScanHeight, err := b.GetChainHeight()
	if err != nil {
		return nil, err
	}
	// reduce the scan height a little in case there is a block reorg
	if walletScanHeight >= monero.MinSpendConfirmations {
		walletScanHeight -= monero.MinSpendConfirmations
	}

	ethHeader, err := b.EthClient().HeaderByNumber(b.Ctx(), nil)
	if err != nil {
		return nil, err
	}

	// set up ethereum event watchers
	logClaimedCh := make(chan []ethtypes.Log)

	ctx, cancel := context.WithCancel(b.Ctx())

	claimedWatcher := watcher.NewEventFilterer(
		ctx,
		b.EthClient(),
		b.ContractAddr(),
		ethHeader.Number,
		claimedTopic,
		logClaimedCh,
	)

	err = claimedWatcher.Start()
	if err != nil {
		cancel()
		return nil, err
	}

	s := &swapState{
		ctx:               ctx,
		cancel:            cancel,
		Backend:           b,
		sender:            sender,
		infoFile:          infofile,
		transferBack:      transferBack,
		walletScanHeight:  walletScanHeight,
		nextExpectedEvent: &EventKeysReceived{},
		eventCh:           make(chan Event),
		logClaimedCh:      logClaimedCh,
		xmrLockedCh:       make(chan struct{}),
		claimedCh:         make(chan struct{}),
		done:              make(chan struct{}),
		info:              info,
		statusCh:          statusCh,
		ethAsset:          ethAsset,
	}

	if err := pcommon.WriteContractAddressToFile(s.infoFile, b.ContractAddr().String()); err != nil {
		return nil, fmt.Errorf("failed to write contract address to file: %w", err)
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
	_ = s.Exit()
}

// SendKeysMessage ...
func (s *swapState) SendKeysMessage() (*net.SendKeysMessage, error) {
	if err := s.generateAndSetKeys(); err != nil {
		return nil, err
	}

	return &net.SendKeysMessage{
		PublicSpendKey:     s.pubkeys.SpendKey().Hex(),
		PublicViewKey:      s.pubkeys.ViewKey().Hex(),
		DLEqProof:          hex.EncodeToString(s.dleqProof.Proof()),
		Secp256k1PublicKey: s.secp256k1Pub.String(),
	}, nil
}

// InfoFile returns the swap's infoFile path
func (s *swapState) InfoFile() string {
	return s.infoFile
}

// ReceivedAmount returns the amount received, or expected to be received, at the end of the swap
func (s *swapState) ReceivedAmount() float64 {
	return s.info.ReceivedAmount
}

func (s *swapState) providedAmountInWei() common.EtherAmount {
	return common.EtherToWei(s.info.ProvidedAmount)
}

func (s *swapState) receivedAmountInPiconero() common.MoneroAmount {
	return common.MoneroToPiconero(s.info.ReceivedAmount)
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
		// stop all running goroutines
		s.cancel()
		close(s.done)

		err := s.SwapManager().CompleteOngoingSwap(s.info.ID)
		if err != nil {
			log.Warnf("failed to mark swap %s as completed: %s", s.info.ID, err)
			return
		}

		if s.info.Status == types.CompletedSuccess {
			str := color.New(color.Bold).Sprintf("**swap completed successfully: id=%s**", s.info.ID)
			log.Info(str)
			return
		}

		if s.info.Status == types.CompletedRefund {
			str := color.New(color.Bold).Sprintf("**swap refunded successfully: id=%s**", s.info.ID)
			log.Info(str)
			return
		}
	}()

	log.Debugf("attempting to exit swap: nextExpectedEvent=%T", s.nextExpectedEvent)

	switch s.nextExpectedEvent.(type) {
	case *EventKeysReceived:
		// we are fine, as we only just initiated the protocol.
		s.clearNextExpectedEvent(types.CompletedAbort)
		return nil
	case *EventXMRLocked, *EventETHClaimed:
		// for EventXMRLocked, we already deployed the contract,
		// so we should call Refund().
		//
		// for EventETHClaimed, the XMR has been locked, but the
		// ETH hasn't been claimed.
		// we should also refund in this case.
		txHash, err := s.tryRefund()
		if err != nil {
			if strings.Contains(err.Error(), revertSwapCompleted) {
				return s.tryClaim()
			}

			s.clearNextExpectedEvent(types.CompletedAbort)
			log.Errorf("failed to refund: err=%s", err)
			return err
		}

		s.clearNextExpectedEvent(types.CompletedRefund)
		log.Infof("refunded ether: transaction hash=%s", txHash)
	case nil:
		// the swap completed already, do nothing
		return nil
	default:
		log.Errorf("unexpected nextExpectedEvent: %T", s.nextExpectedEvent)
		s.clearNextExpectedEvent(types.CompletedAbort)
		return errUnexpectedEventType
	}

	return nil
}

func (s *swapState) tryClaim() error {
	if !s.info.Status.IsOngoing() {
		return nil
	}

	skA, err := s.filterForClaim()
	if err != nil {
		return err
	}

	addr, err := s.claimMonero(skA)
	if err != nil {
		return err
	}

	log.Infof("claimed monero: address=%s", addr)
	s.clearNextExpectedEvent(types.CompletedSuccess)
	return nil
}

// doRefund is called by the RPC function swap_refund.
// If it's possible to refund the ongoing swap, it does that, then notifies the counterparty.
func (s *swapState) doRefund() (ethcommon.Hash, error) {
	switch s.nextExpectedEvent.(type) {
	case *EventXMRLocked, *EventETHClaimed:
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
	stage, err := s.Contract().Swaps(s.CallOpts(), s.contractSwapID)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	switch stage {
	case contracts.StageInvalid:
		return ethcommon.Hash{}, errRefundInvalid
	case contracts.StageCompleted:
		return ethcommon.Hash{}, errRefundSwapCompleted
	case contracts.StagePending, contracts.StageReady:
		// do nothing
	default:
		panic("Unhandled stage value")
	}

	isReady := stage == contracts.StageReady

	ts, err := s.LatestBlockTimestamp(s.ctx)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	log.Debugf("tryRefund isReady=%v untilT0=%vs untilT1=%vs",
		isReady, s.t0.Sub(ts).Seconds(), s.t1.Sub(ts).Seconds())

	if ts.Before(s.t0) && !isReady {
		txHash, err := s.refund()
		// TODO: Have refund() return errors that we can use errors.Is to check against
		if err == nil || !strings.Contains(err.Error(), revertUnableToRefund) {
			return txHash, err
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

	event := <-s.eventCh
	log.Debugf("got event %T", event)
	switch event.(type) {
	case *EventShouldRefund:
		return s.refund()
	case *EventETHClaimed:
		// we should claim
		// this causes the caling function to claim
		return ethcommon.Hash{}, fmt.Errorf(revertSwapCompleted)
	default:
		panic(fmt.Sprintf("got unexpected event while waiting for Claimed/T1: %T", event))
	}
}

func (s *swapState) setTimeouts(t0, t1 *big.Int) {
	s.t0 = time.Unix(t0.Int64(), 0)
	s.t1 = time.Unix(t1.Int64(), 0)
}

func (s *swapState) generateAndSetKeys() error {
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

	return pcommon.WriteKeysToFile(s.infoFile, s.privkeys, s.Env())
}

// generateKeys generates XMRTaker's monero spend and view keys (S_b, V_b), a secp256k1 public key,
// and a DLEq proof proving that the two keys correspond.
func generateKeys() (*pcommon.KeysAndProof, error) {
	return pcommon.GenerateKeysAndProof()
}

// getSecret secrets returns the current secret scalar used to unlock funds from the contract.
func (s *swapState) getSecret() [32]byte {
	secret := s.dleqProof.Secret()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret[:]))
	return sc
}

// setXMRMakerKeys sets XMRMaker's public spend key (to be stored in the contract) and XMRMaker's
// private view key (used to check XMR balance before calling Ready())
func (s *swapState) setXMRMakerKeys(sk *mcrypto.PublicKey, vk *mcrypto.PrivateViewKey,
	secp256k1Pub *secp256k1.PublicKey) {
	s.xmrmakerPublicSpendKey = sk
	s.xmrmakerPrivateViewKey = vk
	s.xmrmakerSecp256k1PublicKey = secp256k1Pub
}

func (s *swapState) approveToken() error {
	token, err := contracts.NewIERC20(s.ethAsset.Address(), s.EthClient())
	if err != nil {
		return fmt.Errorf("failed to instantiate IERC20: %w", err)
	}

	balance, err := token.BalanceOf(s.CallOpts(), s.EthAddress())
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
// TODO: update units to not necessarily be an EtherAmount
func (s *swapState) lockAsset(amount common.EtherAmount) (ethcommon.Hash, error) {
	if s.pubkeys == nil {
		return ethcommon.Hash{}, errNoPublicKeysSet
	}

	if s.xmrmakerPublicSpendKey == nil || s.xmrmakerPrivateViewKey == nil {
		return ethcommon.Hash{}, errCounterpartyKeysNotSet
	}

	if s.ethAsset != types.EthAssetETH {
		err := s.approveToken()
		if err != nil {
			return ethcommon.Hash{}, err
		}
	}

	cmtXMRTaker := s.secp256k1Pub.Keccak256()
	cmtXMRMaker := s.xmrmakerSecp256k1PublicKey.Keccak256()

	nonce := generateNonce()
	txHash, receipt, err := s.sender.NewSwap(cmtXMRMaker, cmtXMRTaker,
		s.xmrmakerAddress, big.NewInt(int64(s.SwapTimeout().Seconds())), nonce,
		s.ethAsset, amount.BigInt())
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to instantiate swap on-chain: %w", err)
	}

	log.Debugf("instantiated swap on-chain: amount=%s asset=%s txHash=%s", amount, s.ethAsset, txHash)

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

	s.setTimeouts(t0, t1)

	s.contractSwap = contracts.SwapFactorySwap{
		Owner:        s.EthAddress(),
		Claimer:      s.xmrmakerAddress,
		PubKeyClaim:  cmtXMRMaker,
		PubKeyRefund: cmtXMRTaker,
		Timeout0:     t0,
		Timeout1:     t1,
		Asset:        ethcommon.Address(s.ethAsset),
		Value:        amount.BigInt(),
		Nonce:        nonce,
	}

	if err := pcommon.WriteContractSwapToFile(s.infoFile, s.contractSwapID, s.contractSwap); err != nil {
		return ethcommon.Hash{}, err
	}

	return txHash, nil
}

// ready calls the Ready() method on the Swap contract, indicating to XMRMaker he has until time t_1 to
// call Claim(). Ready() should only be called once XMRTaker sees XMRMaker lock his XMR.
// If time t_0 has passed, there is no point of calling Ready().
func (s *swapState) ready() error {
	stage, err := s.Contract().Swaps(s.CallOpts(), s.contractSwapID)
	if err != nil {
		return err
	}
	if stage != contracts.StagePending {
		return fmt.Errorf("can not set contract to ready when swap stage is %s", contracts.StageToString(stage))
	}
	_, receipt, err := s.sender.SetReady(s.contractSwap)
	if err != nil {
		if strings.Contains(err.Error(), revertSwapCompleted) && !s.info.Status.IsOngoing() {
			return nil
		}
		return err
	}

	log.Debugf("contract set to ready in block %d", receipt.BlockNumber)
	return nil
}

// refund calls the Refund() method in the Swap contract, revealing XMRTaker's secret
// and returns to her the ether in the contract.
// If time t_1 passes and Claim() has not been called, XMRTaker should call Refund().
func (s *swapState) refund() (ethcommon.Hash, error) {
	if s.Contract() == nil {
		return ethcommon.Hash{}, errNoSwapContractSet
	}

	sc := s.getSecret()

	log.Infof("attempting to call Refund()...")
	txHash, _, err := s.sender.Refund(s.contractSwap, sc)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	s.clearNextExpectedEvent(types.CompletedRefund)
	return txHash, nil
}

func (s *swapState) claimMonero(skB *mcrypto.PrivateSpendKey) (mcrypto.Address, error) {
	if !s.info.Status.IsOngoing() {
		return "", errSwapCompleted
	}

	skAB := mcrypto.SumPrivateSpendKeys(skB, s.privkeys.SpendKey())
	vkAB := mcrypto.SumPrivateViewKeys(s.xmrmakerPrivateViewKey, s.privkeys.ViewKey())
	kpAB := mcrypto.NewPrivateKeyPair(skAB, vkAB)

	// write keys to file in case something goes wrong
	if err := pcommon.WriteSharedSwapKeyPairToFile(s.infoFile, kpAB, s.Env()); err != nil {
		return "", err
	}

	s.LockClient()
	defer s.UnlockClient()

	addr, err := monero.CreateWallet("xmrtaker-swap-wallet", s.Env(), s.Backend, kpAB, s.walletScanHeight)
	if err != nil {
		return "", err
	}

	if !s.transferBack {
		log.Infof("monero claimed in account %s", addr)
		return addr, nil
	}

	id := s.ID()
	depositAddr, err := s.XMRDepositAddress(&id)
	if err != nil {
		return "", err
	}

	log.Infof("monero claimed in account %s; transferring to original account %s",
		addr, depositAddr)

	err = mcrypto.ValidateAddress(string(depositAddr), s.Env())
	if err != nil {
		log.Errorf("failed to transfer to original account, address %s is invalid", addr)
		return addr, nil
	}

	err = s.waitUntilBalanceUnlocks()
	if err != nil {
		return "", fmt.Errorf("failed to wait for balance to unlock: %w", err)
	}

	_, err = s.SweepAll(depositAddr, 0)
	if err != nil {
		return "", fmt.Errorf("failed to send funds to original account: %w", err)
	}

	close(s.claimedCh)
	return addr, nil
}

func (s *swapState) waitUntilBalanceUnlocks() error {
	for {
		if s.ctx.Err() != nil {
			return s.ctx.Err()
		}

		log.Infof("checking if balance unlocked...")
		balance, err := s.GetBalance(0)
		if err != nil {
			return fmt.Errorf("failed to get balance: %w", err)
		}

		if balance.Balance == balance.UnlockedBalance {
			return nil
		}
		if _, err = monero.WaitForBlocks(s.ctx, s, int(balance.BlocksToUnlock)); err != nil {
			log.Warnf("Waiting for %d monero blocks failed: %s", balance.BlocksToUnlock, err)
		}
	}
}

func generateNonce() *big.Int {
	u256PlusOne := big.NewInt(0).Lsh(big.NewInt(1), 256)
	maxU256 := big.NewInt(0).Sub(u256PlusOne, big.NewInt(1))
	n, _ := rand.Int(rand.Reader, maxU256)
	return n
}
