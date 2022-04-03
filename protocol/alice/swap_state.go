package alice

import (
	"context"
	"encoding/hex"
	"errors"
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
	pswap "github.com/noot/atomic-swap/protocol/swap"
	"github.com/noot/atomic-swap/swapfactory"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color" //nolint:misspell
)

const revertSwapCompleted = "swap is already completed"

// swapState is an instance of a swap. it holds the info needed for the swap,
// and its current state.
type swapState struct {
	alice  *Instance
	ctx    context.Context
	cancel context.CancelFunc
	sync.Mutex
	infofile string

	info     *pswap.Info
	statusCh chan types.Status

	// our keys for this session
	dleqProof    *dleq.Proof
	secp256k1Pub *secp256k1.PublicKey
	privkeys     *mcrypto.PrivateKeyPair
	pubkeys      *mcrypto.PublicKeyPair

	// Bob's keys for this session
	bobPublicSpendKey     *mcrypto.PublicKey
	bobPrivateViewKey     *mcrypto.PrivateViewKey
	bobSecp256k1PublicKey *secp256k1.PublicKey
	bobAddress            ethcommon.Address

	// swap contract and timeouts in it; set once contract is deployed
	contractSwapID *big.Int
	t0, t1         time.Time
	txOpts         *bind.TransactOpts

	// next expected network message
	nextExpectedMessage net.Message

	// channels
	xmrLockedCh chan struct{}
	claimedCh   chan struct{}
}

func newSwapState(a *Instance, infofile string, providesAmount common.EtherAmount,
	receivedAmount common.MoneroAmount, exhangeRate types.ExchangeRate) (*swapState, error) {
	txOpts, err := bind.NewKeyedTransactorWithChainID(a.ethPrivKey, a.chainID)
	if err != nil {
		return nil, err
	}

	txOpts.GasPrice = a.gasPrice
	txOpts.GasLimit = a.gasLimit

	stage := types.ExpectingKeys
	statusCh := make(chan types.Status, 16)
	statusCh <- stage
	info := pswap.NewInfo(types.ProvidesETH, providesAmount.AsEther(), receivedAmount.AsMonero(),
		exhangeRate, stage, statusCh)
	if err := a.swapManager.AddSwap(info); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(a.ctx)
	s := &swapState{
		ctx:                 ctx,
		cancel:              cancel,
		alice:               a,
		infofile:            infofile,
		txOpts:              txOpts,
		nextExpectedMessage: &net.SendKeysMessage{},
		xmrLockedCh:         make(chan struct{}),
		claimedCh:           make(chan struct{}),
		info:                info,
		statusCh:            statusCh,
	}

	if err := pcommon.WriteSwapIDToFile(infofile, info.ID()); err != nil {
		return nil, err
	}

	if err := pcommon.WriteContractAddressToFile(s.infofile, a.contractAddr.String()); err != nil {
		return nil, fmt.Errorf("failed to write contract address to file: %w", err)
	}

	return s, nil
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

// InfoFile returns the swap's infofile path
func (s *swapState) InfoFile() string {
	return s.infofile
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
func (s *swapState) ID() uint64 {
	return s.info.ID()
}

// Exit is called by the network when the protocol stream closes, or if the swap_refund RPC endpoint is called.
// It exists the swap by refunding if necessary. If no locking has been done, it simply aborts the swap.
// If the swap already completed successfully, this function does not doing anything in regards to the protoco.
func (s *swapState) Exit() error {
	s.Lock()
	defer s.Unlock()

	defer func() {
		// stop all running goroutines
		s.cancel()
		s.alice.swapState = nil
		s.alice.swapManager.CompleteOngoingSwap()

		if s.info.Status() == types.CompletedSuccess {
			str := color.New(color.Bold).Sprintf("**swap completed successfully: id=%d**", s.info.ID())
			log.Info(str)
			return
		}

		if s.info.Status() == types.CompletedRefund {
			str := color.New(color.Bold).Sprintf("**swap refunded successfully! id=%d**", s.info.ID())
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
			// seems like Bob claimed already - try to claim monero
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
		if err = s.alice.net.SendSwapMessage(&message.NotifyRefund{
			TxHash: txHash.String(),
		}); err != nil {
			return ethcommon.Hash{}, fmt.Errorf("failed to send refund message: err=%w", err)
		}

		return txHash, nil
	default:
		return ethcommon.Hash{}, errCannotRefund
	}
}

func (s *swapState) tryRefund() (ethcommon.Hash, error) {
	untilT0 := time.Until(s.t0)
	untilT1 := time.Until(s.t1)

	isReady, err := s.alice.contract.IsReady(s.alice.callOpts, s.contractSwapID)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	log.Debugf("tryRefund isReady=%v untilT0=%vs untilT1=%vs", isReady, untilT0.Seconds(), untilT1.Seconds())

	if (untilT0 > 0 || isReady) && untilT1 > 0 {
		// we've passed t0 but aren't past t1 yet, so we need to wait until t1
		log.Infof("waiting until time %s to refund", s.t1)
		<-time.After(untilT1)
	}

	return s.refund()
}

func (s *swapState) setTimeouts() error {
	if s.alice.contract == nil {
		return errors.New("contract is nil")
	}

	if (s.t0 != time.Time{}) && (s.t1 != time.Time{}) {
		return nil
	}

	// TODO: add maxRetries
	for {
		log.Debug("attempting to fetch timestamps from contract")

		info, err := s.alice.contract.Swaps(s.alice.callOpts, s.contractSwapID)
		if err != nil {
			time.Sleep(time.Second * 10)
			continue
		}

		s.t0 = time.Unix(info.Timeout0.Int64(), 0)
		s.t1 = time.Unix(info.Timeout1.Int64(), 0)
		break
	}

	return nil
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

	return pcommon.WriteKeysToFile(s.infofile, s.privkeys, s.alice.env)
}

// generateKeys generates Alice's monero spend and view keys (S_b, V_b), a secp256k1 public key,
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

// setBobKeys sets Bob's public spend key (to be stored in the contract) and Bob's
// private view key (used to check XMR balance before calling Ready())
func (s *swapState) setBobKeys(sk *mcrypto.PublicKey, vk *mcrypto.PrivateViewKey, secp256k1Pub *secp256k1.PublicKey) {
	s.bobPublicSpendKey = sk
	s.bobPrivateViewKey = vk
	s.bobSecp256k1PublicKey = secp256k1Pub
}

// lockETH the Swap contract function new_swap and locks `amount` ether in it.
func (s *swapState) lockETH(amount common.EtherAmount) (ethcommon.Hash, error) {
	if s.pubkeys == nil {
		return ethcommon.Hash{}, errors.New("public keys aren't set")
	}

	if s.bobPublicSpendKey == nil || s.bobPrivateViewKey == nil {
		return ethcommon.Hash{}, errors.New("bob's keys aren't set")
	}

	cmtAlice := s.secp256k1Pub.Keccak256()
	cmtBob := s.bobSecp256k1PublicKey.Keccak256()

	s.txOpts.Value = amount.BigInt()
	defer func() {
		s.txOpts.Value = nil
	}()

	tx, err := s.alice.contract.NewSwap(s.txOpts,
		cmtBob, cmtAlice, s.bobAddress, big.NewInt(int64(s.alice.swapTimeout.Seconds())))
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to instantiate swap on-chain: %w", err)
	}

	log.Debugf("instantiating swap on-chain: amount=%s txHash=%s", amount, tx.Hash())
	receipt, err := common.WaitForReceipt(s.ctx, s.alice.ethClient, tx.Hash())
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to call new_swap in contract: %w", err)
	}

	if len(receipt.Logs) == 0 {
		return ethcommon.Hash{}, errors.New("expected 1 log, got 0")
	}

	s.contractSwapID, err = swapfactory.GetIDFromLog(receipt.Logs[0])
	if err != nil {
		return ethcommon.Hash{}, err
	}

	return tx.Hash(), nil
}

// ready calls the Ready() method on the Swap contract, indicating to Bob he has until time t_1 to
// call Claim(). Ready() should only be called once Alice sees Bob lock his XMR.
// If time t_0 has passed, there is no point of calling Ready().
func (s *swapState) ready() error {
	tx, err := s.alice.contract.SetReady(s.txOpts, s.contractSwapID)
	if err != nil {
		if strings.Contains(err.Error(), revertSwapCompleted) && !s.info.Status().IsOngoing() {
			return nil
		}

		return err
	}

	if _, err := common.WaitForReceipt(s.ctx, s.alice.ethClient, tx.Hash()); err != nil {
		return fmt.Errorf("failed to call is_ready in swap contract: %w", err)
	}

	return nil
}

// refund calls the Refund() method in the Swap contract, revealing Alice's secret
// and returns to her the ether in the contract.
// If time t_1 passes and Claim() has not been called, Alice should call Refund().
func (s *swapState) refund() (ethcommon.Hash, error) {
	if s.alice.contract == nil {
		return ethcommon.Hash{}, errors.New("contract is nil")
	}

	sc := s.getSecret()

	log.Infof("attempting to call Refund()...")
	tx, err := s.alice.contract.Refund(s.txOpts, s.contractSwapID, sc)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	if _, err := common.WaitForReceipt(s.ctx, s.alice.ethClient, tx.Hash()); err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to call Refund function in contract: %w", err)
	}

	s.clearNextExpectedMessage(types.CompletedRefund)
	return tx.Hash(), nil
}

func (s *swapState) claimMonero(skB *mcrypto.PrivateSpendKey) (mcrypto.Address, error) {
	if !s.info.Status().IsOngoing() {
		return "", errors.New("swap has already completed")
	}

	skAB := mcrypto.SumPrivateSpendKeys(skB, s.privkeys.SpendKey())
	vkAB := mcrypto.SumPrivateViewKeys(s.bobPrivateViewKey, s.privkeys.ViewKey())
	kpAB := mcrypto.NewPrivateKeyPair(skAB, vkAB)

	// write keys to file in case something goes wrong
	if err := pcommon.WriteSharedSwapKeyPairToFile(s.infofile, kpAB, s.alice.env); err != nil {
		return "", err
	}

	addr, err := monero.CreateMoneroWallet("alice-swap-wallet", s.alice.env, s.alice.client, kpAB)
	if err != nil {
		return "", err
	}

	log.Infof("monero claimed in account %s; transferring to original account %s",
		addr, s.alice.walletAddress)

	err = s.waitUntilBalanceUnlocks()
	if err != nil {
		return "", fmt.Errorf("failed to wait for balance to unlock: %w", err)
	}

	_, err = s.alice.client.SweepAll(s.alice.walletAddress, 0)
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

		balance, err := s.alice.client.GetBalance(0)
		if err != nil {
			return fmt.Errorf("failed to get balance: %w", err)
		}

		if balance.Balance == balance.UnlockedBalance {
			return nil
		}

		time.Sleep(time.Second)
	}
}
