package xmrtaker

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	"github.com/noot/atomic-swap/crypto/secp256k1"
	"github.com/noot/atomic-swap/dleq"
	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/net/message"
	pcommon "github.com/noot/atomic-swap/protocol"
	"github.com/noot/atomic-swap/protocol/backend"
	pswap "github.com/noot/atomic-swap/protocol/swap"
	"github.com/noot/atomic-swap/swapfactory"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color" //nolint:misspell
)

const revertSwapCompleted = "swap is already completed"
const revertUnableToRefund = "it's the counterparty's turn, unable to refund, try again later"

// swapState is an instance of a swap. it holds the info needed for the swap,
// and its current state.
type swapState struct {
	backend.Backend

	ctx          context.Context
	cancel       context.CancelFunc
	stateMu      sync.Mutex
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

	// swap contract and timeouts in it; set once contract is deployed
	contractSwapID [32]byte
	contractSwap   swapfactory.SwapFactorySwap
	t0, t1         time.Time

	// next expected network message
	nextExpectedMessage net.Message

	// channels
	xmrLockedCh chan struct{}
	claimedCh   chan struct{}
	done        chan struct{}
	exited      bool
}

func newSwapState(b backend.Backend, offerID types.Hash, infofile string, transferBack bool,
	providesAmount common.EtherAmount, receivedAmount common.MoneroAmount,
	exchangeRate types.ExchangeRate) (*swapState, error) {
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
	info := pswap.NewInfo(offerID, types.ProvidesETH, providesAmount.AsEther(), receivedAmount.AsMonero(),
		exchangeRate, stage, statusCh)
	if err := b.SwapManager().AddSwap(info); err != nil {
		return nil, err
	}

	if b.ExternalSender() != nil {
		transferBack = true // front-end must set final deposit address
	}

	ctx, cancel := context.WithCancel(b.Ctx())
	s := &swapState{
		ctx:                 ctx,
		cancel:              cancel,
		Backend:             b,
		infoFile:            infofile,
		transferBack:        transferBack,
		nextExpectedMessage: &net.SendKeysMessage{},
		xmrLockedCh:         make(chan struct{}),
		claimedCh:           make(chan struct{}),
		done:                make(chan struct{}),
		info:                info,
		statusCh:            statusCh,
	}

	if err := pcommon.WriteContractAddressToFile(s.infoFile, b.ContractAddr().String()); err != nil {
		return nil, fmt.Errorf("failed to write contract address to file: %w", err)
	}

	go s.waitForSendKeysMessage()
	return s, nil
}

func (s *swapState) lockState() {
	s.stateMu.Lock()
}

func (s *swapState) unlockState() {
	s.stateMu.Unlock()
}

func (s *swapState) waitForSendKeysMessage() {
	waitDuration := time.Minute
	timer := time.After(waitDuration)
	select {
	case <-s.ctx.Done():
		return
	case <-timer:
	}

	// check if we've received a response from the counterparty yet
	if s.nextExpectedMessage != (&net.SendKeysMessage{}) {
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
	return s.info.ReceivedAmount()
}

func (s *swapState) providedAmountInWei() common.EtherAmount {
	return common.EtherToWei(s.info.ProvidedAmount())
}

func (s *swapState) receivedAmountInPiconero() common.MoneroAmount {
	return common.MoneroToPiconero(s.info.ReceivedAmount())
}

// ID returns the ID of the swap
func (s *swapState) ID() types.Hash {
	return s.info.ID()
}

// Exit is called by the network when the protocol stream closes, or if the swap_refund RPC endpoint is called.
// It exists the swap by refunding if necessary. If no locking has been done, it simply aborts the swap.
// If the swap already completed successfully, this function does not do anything regarding the protocol.
func (s *swapState) Exit() error {
	s.lockState()
	defer s.unlockState()
	return s.exit()
}

// exit is the same as Exit, but assumes the calling code block already holds the swapState lock.
func (s *swapState) exit() error {
	if s.exited {
		return nil
	}
	s.exited = true

	defer func() {
		// stop all running goroutines
		s.cancel()
		s.SwapManager().CompleteOngoingSwap(s.info.ID())
		close(s.done)

		if s.info.Status() == types.CompletedSuccess {
			str := color.New(color.Bold).Sprintf("**swap completed successfully: id=%s**", s.info.ID())
			log.Info(str)
			return
		}

		if s.info.Status() == types.CompletedRefund {
			str := color.New(color.Bold).Sprintf("**swap refunded successfully! id=%s**", s.info.ID())
			log.Info(str)
			return
		}
	}()

	log.Debugf("attempting to exit swap: nextExpectedMessage=%s", s.nextExpectedMessage)

	switch s.nextExpectedMessage.(type) {
	case *net.SendKeysMessage:
		// we are fine, as we only just initiated the protocol.
		s.clearNextExpectedMessage(types.CompletedAbort)
		return nil
	case *message.NotifyXMRLock:
		// we already deployed the contract, so we should call Refund().
		txHash, err := s.tryRefund()
		if err != nil {
			if strings.Contains(err.Error(), revertSwapCompleted) {
				return s.tryClaim()
			}

			s.clearNextExpectedMessage(types.CompletedAbort)
			log.Errorf("failed to refund: err=%s", err)
			return err
		}

		s.clearNextExpectedMessage(types.CompletedRefund)
		log.Infof("refunded ether: transaction hash=%s", txHash)
	case *message.NotifyClaimed:
		// the XMR has been locked, but the ETH hasn't been claimed.
		// we should also refund in this case.
		txHash, err := s.tryRefund()
		if err != nil {
			// seems like XMRMaker claimed already - try to claim monero
			if strings.Contains(err.Error(), revertSwapCompleted) {
				return s.tryClaim()
			}

			s.clearNextExpectedMessage(types.CompletedAbort)
			log.Errorf("failed to refund: err=%s", err)
			return err
		}

		s.clearNextExpectedMessage(types.CompletedRefund)
		log.Infof("refunded ether: transaction hash=%s", txHash)
	case nil:
		return s.tryClaim()
	default:
		log.Errorf("unexpected nextExpectedMessage in Exit: type=%T", s.nextExpectedMessage)
		s.clearNextExpectedMessage(types.CompletedAbort)
		return errUnexpectedMessageType
	}

	return nil
}

func (s *swapState) tryClaim() error {
	if !s.info.Status().IsOngoing() {
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
	s.clearNextExpectedMessage(types.CompletedSuccess)
	return nil
}

// doRefund is called by the RPC function swap_refund.
// If it's possible to refund the ongoing swap, it does that, then notifies the counterparty.
func (s *swapState) doRefund() (ethcommon.Hash, error) {
	switch s.nextExpectedMessage.(type) {
	case *message.NotifyXMRLock, *message.NotifyClaimed:
		// the XMR has been locked, but the ETH hasn't been claimed.
		// we can refund in this case.
		txHash, err := s.tryRefund()
		if err != nil {
			s.clearNextExpectedMessage(types.CompletedAbort)
			log.Errorf("failed to refund: err=%s", err)
			return ethcommon.Hash{}, err
		}

		s.clearNextExpectedMessage(types.CompletedRefund)
		log.Infof("refunded ether: transaction hash=%s", txHash)

		// send NotifyRefund msg
		if err = s.SendSwapMessage(&message.NotifyRefund{
			TxHash: txHash.String(),
		}, s.ID()); err != nil {
			return ethcommon.Hash{}, fmt.Errorf("failed to send refund message: err=%w", err)
		}

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
	case swapfactory.StageInvalid:
		return ethcommon.Hash{}, errRefundInvalid
	case swapfactory.StageCompleted:
		return ethcommon.Hash{}, errRefundSwapCompleted
	case swapfactory.StagePending, swapfactory.StageReady:
		// do nothing
	default:
		panic("Unhandled stage value")
	}
	isReady := stage == swapfactory.StageReady

	ts, err := s.LatestBlockTimestamp(s.ctx)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	log.Debugf("tryRefund isReady=%v untilT0=%vs untilT1=%vs",
		isReady, s.t0.Sub(ts).Seconds(), s.t1.Sub(ts).Seconds())

	if ts.Before(s.t0) && !isReady {
		txHash, err := s.refund() //nolint:govet
		// TODO: Have refund() return errors that we can use errors.Is to check against
		if err == nil || !strings.Contains(err.Error(), revertUnableToRefund) {
			return txHash, err
		}
		// There is a small, but non-zero chance that our transaction gets placed in a block that is after T0
		// even though the current block is before T0. In this case, the transaction will be reverted, the
		// gas fee is lost, but we can wait until T1 and try again.
		log.Warnf("First refund attempt failed: err=%s", err)
	}

	if ts.Before(s.t1) {
		// we've passed t0 but aren't past t1 yet, so wait until t1
		log.Infof("Waiting until time %s to refund", s.t1)
		err = s.WaitForTimestamp(s.ctx, s.t1)
		if err != nil {
			return ethcommon.Hash{}, err
		}
	}
	return s.refund()
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

// lockETH the Swap contract function new_swap and locks `amount` ether in it.
func (s *swapState) lockETH(amount common.EtherAmount) (ethcommon.Hash, error) {
	if s.pubkeys == nil {
		return ethcommon.Hash{}, errNoPublicKeysSet
	}

	if s.xmrmakerPublicSpendKey == nil || s.xmrmakerPrivateViewKey == nil {
		return ethcommon.Hash{}, errCounterpartyKeysNotSet
	}

	cmtXMRTaker := s.secp256k1Pub.Keccak256()
	cmtXMRMaker := s.xmrmakerSecp256k1PublicKey.Keccak256()

	nonce := generateNonce()
	txHash, receipt, err := s.NewSwap(s.ID(), cmtXMRMaker, cmtXMRTaker,
		s.xmrmakerAddress, big.NewInt(int64(s.SwapTimeout().Seconds())), nonce, amount.BigInt())
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to instantiate swap on-chain: %w", err)
	}

	log.Debugf("instantiated swap on-chain: amount=%s txHash=%s", amount, txHash)

	if len(receipt.Logs) == 0 {
		return ethcommon.Hash{}, errSwapInstantiationNoLogs
	}

	s.contractSwapID, err = swapfactory.GetIDFromLog(receipt.Logs[0])
	if err != nil {
		return ethcommon.Hash{}, err
	}

	t0, t1, err := swapfactory.GetTimeoutsFromLog(receipt.Logs[0])
	if err != nil {
		return ethcommon.Hash{}, err
	}

	s.setTimeouts(t0, t1)

	s.contractSwap = swapfactory.SwapFactorySwap{
		Owner:        s.EthAddress(),
		Claimer:      s.xmrmakerAddress,
		PubKeyClaim:  cmtXMRMaker,
		PubKeyRefund: cmtXMRTaker,
		Timeout0:     t0,
		Timeout1:     t1,
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
	if stage != swapfactory.StagePending {
		return fmt.Errorf("can not set contract to ready when swap stage is %s", swapfactory.StageToString(stage))
	}
	_, _, err = s.SetReady(s.ID(), s.contractSwap)
	if err != nil {
		if strings.Contains(err.Error(), revertSwapCompleted) && !s.info.Status().IsOngoing() {
			return nil
		}
		return err
	}

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
	txHash, _, err := s.Refund(s.ID(), s.contractSwap, sc)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	s.clearNextExpectedMessage(types.CompletedRefund)
	return txHash, nil
}

func (s *swapState) claimMonero(skB *mcrypto.PrivateSpendKey) (mcrypto.Address, error) {
	if !s.info.Status().IsOngoing() {
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

	addr, err := monero.CreateMoneroWallet("xmrtaker-swap-wallet", s.Env(), s.Backend, kpAB)
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

	err = mcrypto.ValidateAddress(string(depositAddr))
	if err != nil {
		log.Errorf("failed to transfer to original account, address %s is invalid", addr)
		return addr, nil
	}

	err = s.waitUntilBalanceUnlocks(depositAddr)
	if err != nil {
		return "", fmt.Errorf("failed to wait for balance to unlock: %w", err)
	}

	res, err := s.SweepAll(depositAddr, 0)
	if err != nil {
		return "", fmt.Errorf("failed to send funds to original account: %w", err)
	}

	if len(res.AmountList) == 0 {
		return "", fmt.Errorf("sweep all did not return any amounts")
	}

	amount := res.AmountList[0]
	log.Infof("transferred %v XMR to %s",
		common.MoneroAmount(amount).AsMonero(),
		depositAddr,
	)

	close(s.claimedCh)
	return addr, nil
}

func (s *swapState) waitUntilBalanceUnlocks(depositAddr mcrypto.Address) error {
	for {
		if s.ctx.Err() != nil {
			return s.ctx.Err()
		}

		log.Infof("checking if balance unlocked...")

		if s.Env() == common.Development {
			_ = s.GenerateBlocks(string(depositAddr), 64)
			_ = s.Refresh()
		}

		balance, err := s.GetBalance(0)
		if err != nil {
			return fmt.Errorf("failed to get balance: %w", err)
		}

		if balance.Balance == balance.UnlockedBalance {
			return nil
		}

		time.Sleep(time.Second * 30)
	}
}

func generateNonce() *big.Int {
	u256PlusOne := big.NewInt(0).Lsh(big.NewInt(1), 256)
	maxU256 := big.NewInt(0).Sub(u256PlusOne, big.NewInt(1))
	n, _ := rand.Int(rand.Reader, maxU256)
	return n
}
