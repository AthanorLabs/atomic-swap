package alice

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	"github.com/noot/atomic-swap/crypto/secp256k1"
	"github.com/noot/atomic-swap/dleq"
	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	pcommon "github.com/noot/atomic-swap/protocol"
	"github.com/noot/atomic-swap/swap-contract"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color" //nolint:misspell
)

var nextID uint64

var (
	errMissingKeys    = errors.New("did not receive Bob's public spend or private view key")
	errMissingAddress = errors.New("did not receive Bob's address")
)

// swapState is an instance of a swap. it holds the info needed for the swap,
// and its current state.
type swapState struct {
	alice  *Instance
	ctx    context.Context
	cancel context.CancelFunc
	sync.Mutex

	id uint64
	// amount of ETH we are providing this swap, and the amount of XMR we should receive.
	providesAmount common.EtherAmount
	desiredAmount  common.MoneroAmount

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
	contract *swap.Swap
	t0, t1   time.Time
	txOpts   *bind.TransactOpts

	// next expected network message
	nextExpectedMessage net.Message // TODO: change to type?

	// channels
	xmrLockedCh chan struct{}
	claimedCh   chan struct{}

	// set to true upon creating of the XMR wallet
	success  bool
	refunded bool
}

func newSwapState(a *Instance, providesAmount common.EtherAmount) (*swapState, error) {
	txOpts, err := bind.NewKeyedTransactorWithChainID(a.ethPrivKey, a.chainID)
	if err != nil {
		return nil, err
	}

	txOpts.GasPrice = a.gasPrice
	txOpts.GasLimit = a.gasLimit

	ctx, cancel := context.WithCancel(a.ctx)

	s := &swapState{
		ctx:                 ctx,
		cancel:              cancel,
		alice:               a,
		id:                  nextID,
		providesAmount:      providesAmount,
		txOpts:              txOpts,
		nextExpectedMessage: &net.SendKeysMessage{},
		xmrLockedCh:         make(chan struct{}),
		claimedCh:           make(chan struct{}),
	}

	nextID++
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

// ReceivedAmount returns the amount received, or expected to be received, at the end of the swap
func (s *swapState) ReceivedAmount() float64 {
	return s.desiredAmount.AsMonero()
}

// ProtocolExited is called by the network when the protocol stream closes.
// If it closes prematurely, we need to perform recovery.
func (s *swapState) ProtocolExited() error {
	s.Lock()
	defer s.Unlock()

	defer func() {
		// stop all running goroutines
		s.cancel()
		s.alice.swapState = nil
	}()

	if s.success {
		str := color.New(color.Bold).Sprintf("**swap completed successfully! id=%d**", s.id)
		log.Info(str)
		return nil
	}

	switch s.nextExpectedMessage.(type) {
	case *net.SendKeysMessage:
		// we are fine, as we only just initiated the protocol.
		return errors.New("swap cancelled early, but before any locking happened")
	case *net.NotifyXMRLock:
		// we already deployed the contract, so we should call Refund().
		txHash, err := s.tryRefund()
		if err != nil {
			log.Errorf("failed to refund: err=%s", err)
			return err
		}

		log.Infof("refunded ether: transaction hash=%s", txHash)
	case *net.NotifyClaimed:
		// the XMR has been locked, but the ETH hasn't been claimed.
		// we should also refund in this case.
		txHash, err := s.tryRefund()
		if err != nil {
			log.Errorf("failed to refund: err=%s", err)
			return err
		}

		log.Infof("refunded ether: transaction hash=%s", txHash)
	default:
		log.Errorf("unexpected nextExpectedMessage in ProtocolExited: type=%T", s.nextExpectedMessage)
		return errors.New("unexpected message type")
	}

	if s.refunded {
		return errors.New("swap refunded")
	}

	return nil
}

func (s *swapState) tryRefund() (ethcommon.Hash, error) {
	untilT0 := time.Until(s.t0)
	untilT1 := time.Until(s.t1)

	// TODO: also check if IsReady == true

	if untilT0 > 0 && untilT1 < 0 {
		// we've passed t0 but aren't past t1 yet, so we need to wait until t1
		log.Infof("waiting until time %s to refund", s.t1)
		<-time.After(untilT1)
	}

	return s.refund()
}

func (s *swapState) setTimeouts() error {
	if s.contract == nil {
		return errors.New("contract is nil")
	}

	if (s.t0 != time.Time{}) && (s.t1 != time.Time{}) {
		return nil
	}

	// TODO: add maxRetries
	for {
		log.Debug("attempting to fetch t0 from contract")

		st0, err := s.contract.Timeout0(s.alice.callOpts)
		if err != nil {
			time.Sleep(time.Second * 10)
			continue
		}

		s.t0 = time.Unix(st0.Int64(), 0)
		break
	}

	for {
		log.Debug("attempting to fetch t1 from contract")

		st1, err := s.contract.Timeout1(s.alice.callOpts)
		if err != nil {
			time.Sleep(time.Second * 10)
			continue
		}

		s.t1 = time.Unix(st1.Int64(), 0)
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

	fp := fmt.Sprintf("%s/%d/alice-secret", s.alice.basepath, s.id)
	if err := mcrypto.WriteKeysToFile(fp, s.privkeys, s.alice.env); err != nil {
		return err
	}

	return nil
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

// deployAndLockETH deploys an instance of the Swap contract and locks `amount` ether in it.
func (s *swapState) deployAndLockETH(amount common.EtherAmount) (ethcommon.Address, error) {
	if s.pubkeys == nil {
		return ethcommon.Address{}, errors.New("public keys aren't set")
	}

	if s.bobPublicSpendKey == nil || s.bobPrivateViewKey == nil {
		return ethcommon.Address{}, errors.New("bob's keys aren't set")
	}

	cmtAlice := s.secp256k1Pub.Keccak256()
	cmtBob := s.bobSecp256k1PublicKey.Keccak256()

	s.txOpts.Value = amount.BigInt()
	defer func() {
		s.txOpts.Value = nil
	}()

	address, tx, swap, err := swap.DeploySwap(s.txOpts, s.alice.ethClient,
		cmtBob, cmtAlice, s.bobAddress, defaultTimeoutDuration)
	if err != nil {
		return ethcommon.Address{}, fmt.Errorf("failed to deploy Swap.sol: %w", err)
	}

	log.Debugf("deploying Swap.sol, amount=%s txHash=%s", amount, tx.Hash())
	if _, ok := common.WaitForReceipt(s.ctx, s.alice.ethClient, tx.Hash()); !ok {
		return ethcommon.Address{}, errors.New("failed to deploy Swap.sol")
	}

	fp := fmt.Sprintf("%s/%d/contractaddress", s.alice.basepath, s.id)
	if err = common.WriteContractAddressToFile(fp, address.String()); err != nil {
		return ethcommon.Address{}, fmt.Errorf("failed to write contract address to file: %w", err)
	}

	balance, err := s.alice.ethClient.BalanceAt(s.ctx, address, nil)
	if err != nil {
		return ethcommon.Address{}, err
	}

	log.Debug("contract balance: ", balance)

	s.contract = swap
	return address, nil
}

// ready calls the Ready() method on the Swap contract, indicating to Bob he has until time t_1 to
// call Claim(). Ready() should only be called once Alice sees Bob lock his XMR.
// If time t_0 has passed, there is no point of calling Ready().
func (s *swapState) ready() error {
	tx, err := s.contract.SetReady(s.txOpts)
	if err != nil {
		return err
	}

	if _, ok := common.WaitForReceipt(s.ctx, s.alice.ethClient, tx.Hash()); !ok {
		return errors.New("failed to set IsReady to true in Swap.sol")
	}

	return nil
}

// refund calls the Refund() method in the Swap contract, revealing Alice's secret
// and returns to her the ether in the contract.
// If time t_1 passes and Claim() has not been called, Alice should call Refund().
func (s *swapState) refund() (ethcommon.Hash, error) {
	if s.contract == nil {
		return ethcommon.Hash{}, errors.New("contract is nil")
	}

	sc := s.getSecret()

	log.Infof("attempting to call Refund()...")
	tx, err := s.contract.Refund(s.txOpts, sc)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	if _, ok := common.WaitForReceipt(s.ctx, s.alice.ethClient, tx.Hash()); !ok {
		return ethcommon.Hash{}, errors.New("failed to call Refund in Swap.sol")
	}

	s.success = true
	s.refunded = true
	return tx.Hash(), nil
}

func (s *swapState) claimMonero(skB *mcrypto.PrivateSpendKey) (mcrypto.Address, error) {
	skAB := mcrypto.SumPrivateSpendKeys(skB, s.privkeys.SpendKey())
	vkAB := mcrypto.SumPrivateViewKeys(s.bobPrivateViewKey, s.privkeys.ViewKey())
	kpAB := mcrypto.NewPrivateKeyPair(skAB, vkAB)

	// write keys to file in case something goes wrong
	fp := fmt.Sprintf("%s/%d/swap-secret", s.alice.basepath, s.id)
	if err := mcrypto.WriteKeysToFile(fp, kpAB, s.alice.env); err != nil {
		return "", err
	}

	return monero.CreateMoneroWallet("alice-swap-wallet", s.alice.env, s.alice.client, kpAB)
}
