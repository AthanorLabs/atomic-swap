package alice

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero"
	mcrypto "github.com/noot/atomic-swap/monero/crypto"
	"github.com/noot/atomic-swap/net"
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
	privkeys *mcrypto.PrivateKeyPair
	pubkeys  *mcrypto.PublicKeyPair

	// Bob's keys for this session
	bobPublicSpendKey *mcrypto.PublicKey
	bobPrivateViewKey *mcrypto.PrivateViewKey
	bobAddress        ethcommon.Address

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
	kp, err := s.generateKeys()
	if err != nil {
		return nil, err
	}

	sig, err := s.privkeys.SpendKey().Sign(kp.SpendKey().Bytes())
	if err != nil {
		return nil, err
	}

	return &net.SendKeysMessage{
		PublicSpendKey:  kp.SpendKey().Hex(),
		PublicViewKey:   kp.ViewKey().Hex(),
		PrivateKeyProof: sig.Hex(),
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

// HandleProtocolMessage is called by the network to handle an incoming message.
// If the message received is not the expected type for the point in the protocol we're at,
// this function will return an error.
func (s *swapState) HandleProtocolMessage(msg net.Message) (net.Message, bool, error) {
	s.Lock()
	defer s.Unlock()

	if err := s.checkMessageType(msg); err != nil {
		return nil, true, err
	}

	switch msg := msg.(type) {
	case *net.SendKeysMessage:
		resp, err := s.handleSendKeysMessage(msg)
		if err != nil {
			return nil, true, err
		}

		return resp, false, nil
	case *net.NotifyXMRLock:
		if msg.Address == "" {
			return nil, true, errors.New("got empty address for locked XMR")
		}
		// check that XMR was locked in expected account, and confirm amount
		vk := mcrypto.SumPrivateViewKeys(s.bobPrivateViewKey, s.privkeys.ViewKey())
		sk := mcrypto.SumPublicKeys(s.bobPublicSpendKey, s.pubkeys.SpendKey())
		kp := mcrypto.NewPublicKeyPair(sk, vk.Public())

		if msg.Address != string(kp.Address(s.alice.env)) {
			return nil, true, fmt.Errorf("address received in message does not match expected address")
		}

		t := time.Now().Format("2006-Jan-2-15:04:05")
		walletName := fmt.Sprintf("alice-viewonly-wallet-%s", t)
		log.Debugf("generating view-only wallet to check funds: %s", walletName)
		if err := s.alice.client.GenerateViewOnlyWalletFromKeys(vk, kp.Address(s.alice.env), walletName, ""); err != nil {
			return nil, true, fmt.Errorf("failed to generate view-only wallet to verify locked XMR: %w", err)
		}

		log.Debugf("generated view-only wallet to check funds: %s", walletName)

		if s.alice.env != common.Development {
			// wait for 2 new blocks, otherwise balance might be 0
			// TODO: check transaction hash
			if err := monero.WaitForBlocks(s.alice.client); err != nil {
				return nil, true, err
			}

			if err := monero.WaitForBlocks(s.alice.client); err != nil {
				return nil, true, err
			}
		}

		if err := s.alice.client.Refresh(); err != nil {
			return nil, true, fmt.Errorf("failed to refresh client: %w", err)
		}

		accounts, err := s.alice.client.GetAccounts()
		if err != nil {
			return nil, true, fmt.Errorf("failed to get accounts: %w", err)
		}

		var (
			balance *monero.GetBalanceResponse
		)

		for i, acc := range accounts.SubaddressAccounts {
			addr, ok := acc["base_address"].(string)
			if !ok {
				panic("address is not a string!")
			}

			if mcrypto.Address(addr) == kp.Address(s.alice.env) {
				balance, err = s.alice.client.GetBalance(uint(i))
				if err != nil {
					return nil, true, fmt.Errorf("failed to get balance: %w", err)
				}

				break
			}
		}

		if balance == nil {
			return nil, true, fmt.Errorf("failed to find account with address %s", kp.Address(s.alice.env))
		}

		log.Debugf("checking locked wallet, address=%s balance=%v", kp.Address(s.alice.env), balance.Balance)

		// TODO: also check that the balance isn't unlocked only after an unreasonable amount of blocks
		if balance.Balance < float64(s.desiredAmount) {
			return nil, true, fmt.Errorf("locked XMR amount is less than expected: got %v, expected %v",
				balance.Balance, float64(s.desiredAmount))
		}

		if err := s.alice.client.CloseWallet(); err != nil {
			return nil, true, fmt.Errorf("failed to close wallet: %w", err)
		}

		close(s.xmrLockedCh)

		if err := s.ready(); err != nil {
			return nil, true, fmt.Errorf("failed to call Ready: %w", err)
		}

		log.Debug("set swap.IsReady to true")

		if err := s.setTimeouts(); err != nil {
			return nil, true, fmt.Errorf("failed to set timeouts: %w", err)
		}

		go func() {
			until := time.Until(s.t1)

			select {
			case <-s.ctx.Done():
				return
			case <-time.After(until + time.Second):
				// Bob hasn't claimed, and we're after t_1. let's call Refund
				txhash, err := s.refund()
				if err != nil {
					log.Errorf("failed to refund: err=%s", err)
					return
				}

				log.Infof("got our ETH back: tx hash=%s", txhash)

				// send NotifyRefund msg
				if err = s.alice.net.SendSwapMessage(&net.NotifyRefund{
					TxHash: txhash.String(),
				}); err != nil {
					log.Errorf("failed to send refund message: err=%s", err)
				}
			case <-s.claimedCh:
				return
			}
		}()

		s.nextExpectedMessage = &net.NotifyClaimed{}
		out := &net.NotifyReady{}
		return out, false, nil
	case *net.NotifyClaimed:
		address, err := s.handleNotifyClaimed(msg.TxHash)
		if err != nil {
			log.Error("failed to create monero address: err=", err)
			return nil, true, err
		}

		close(s.claimedCh)

		log.Info("successfully created monero wallet from our secrets: address=", address)
		return nil, true, nil
	default:
		return nil, false, errors.New("unexpected message type")
	}
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

func (s *swapState) handleSendKeysMessage(msg *net.SendKeysMessage) (net.Message, error) {
	// TODO: get user to confirm amount they will receive!!
	s.desiredAmount = common.MoneroToPiconero(msg.ProvidedAmount)
	log.Infof(color.New(color.Bold).Sprintf("you will be receiving %v XMR", msg.ProvidedAmount))

	if msg.PublicSpendKey == "" || msg.PrivateViewKey == "" {
		return nil, errMissingKeys
	}

	if msg.EthAddress == "" {
		return nil, errMissingAddress
	}

	vk, err := mcrypto.NewPrivateViewKeyFromHex(msg.PrivateViewKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Bob's private view keys: %w", err)
	}

	s.bobAddress = ethcommon.HexToAddress(msg.EthAddress)

	log.Debugf("got Bob's keys and address: address=%s", s.bobAddress)

	sk, err := mcrypto.NewPublicKeyFromHex(msg.PublicSpendKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Bob's public spend key: %w", err)
	}

	// verify that counterparty really has the private key to the public key
	sig, err := mcrypto.NewSignatureFromHex(msg.PrivateKeyProof)
	if err != nil {
		return nil, err
	}

	ok := sk.Verify(sk.Bytes(), sig)
	if !ok {
		return nil, errors.New("failed to verify proof of private key")
	}

	s.setBobKeys(sk, vk)
	address, err := s.deployAndLockETH(s.providesAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy contract: %w", err)
	}

	log.Info("deployed Swap contract, waiting for XMR to be locked: contract address=", address)

	// set t0 and t1
	// TODO: these sometimes fail with "attempting to unmarshall an empty string while arguments are expected"
	if err := s.setTimeouts(); err != nil {
		return nil, err
	}

	// start goroutine to check that Bob locks before t_0
	go func() {
		const timeoutBuffer = time.Minute * 5
		until := time.Until(s.t0)

		select {
		case <-s.ctx.Done():
			return
		case <-time.After(until - timeoutBuffer):
			// Bob hasn't locked yet, let's call refund
			txhash, err := s.refund()
			if err != nil {
				log.Errorf("failed to refund: err=%s", err)
				return
			}

			log.Infof("got our ETH back: tx hash=%s", txhash)

			// send NotifyRefund msg
			if err := s.alice.net.SendSwapMessage(&net.NotifyRefund{
				TxHash: txhash.String(),
			}); err != nil {
				log.Errorf("failed to send refund message: err=%s", err)
			}
		case <-s.xmrLockedCh:
			return
		}

	}()

	s.nextExpectedMessage = &net.NotifyXMRLock{}

	out := &net.NotifyContractDeployed{
		Address: address.String(),
	}

	return out, nil
}

func (s *swapState) checkMessageType(msg net.Message) error {
	if msg.Type() != s.nextExpectedMessage.Type() {
		return errors.New("received unexpected message")
	}

	return nil
}

// generateKeys generates Alice's monero spend and view keys (S_b, V_b)
// It returns Alice's public spend key
func (s *swapState) generateKeys() (*mcrypto.PublicKeyPair, error) {
	if s.privkeys != nil {
		return s.pubkeys, nil
	}

	var err error
	s.privkeys, err = mcrypto.GenerateKeys()
	if err != nil {
		return nil, err
	}

	fp := fmt.Sprintf("%s/%d/alice-secret", s.alice.basepath, s.id)
	if err := mcrypto.WriteKeysToFile(fp, s.privkeys, s.alice.env); err != nil {
		return nil, err
	}

	s.pubkeys = s.privkeys.PublicKeyPair()
	return s.pubkeys, nil
}

// setBobKeys sets Bob's public spend key (to be stored in the contract) and Bob's
// private view key (used to check XMR balance before calling Ready())
func (s *swapState) setBobKeys(sk *mcrypto.PublicKey, vk *mcrypto.PrivateViewKey) {
	s.bobPublicSpendKey = sk
	s.bobPrivateViewKey = vk
}

// deployAndLockETH deploys an instance of the Swap contract and locks `amount` ether in it.
func (s *swapState) deployAndLockETH(amount common.EtherAmount) (ethcommon.Address, error) {
	if s.pubkeys == nil {
		return ethcommon.Address{}, errors.New("public keys aren't set")
	}

	if s.bobPublicSpendKey == nil || s.bobPrivateViewKey == nil {
		return ethcommon.Address{}, errors.New("bob's keys aren't set")
	}

	pkAlice := s.pubkeys.SpendKey().Bytes()
	pkBob := s.bobPublicSpendKey.Bytes()

	var pka, pkb [32]byte
	copy(pka[:], common.Reverse(pkAlice))
	copy(pkb[:], common.Reverse(pkBob))

	// TODO: put auth in swapState
	s.txOpts.Value = amount.BigInt()
	defer func() {
		s.txOpts.Value = nil
	}()

	address, tx, swap, err := swap.DeploySwap(s.txOpts, s.alice.ethClient, pkb, pka, s.bobAddress, defaultTimeoutDuration)
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
	secret := s.privkeys.SpendKeyBytes()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret))

	if s.contract == nil {
		return ethcommon.Hash{}, errors.New("contract is nil")
	}

	log.Infof("atte`mpting to call Refund()...")
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

// handleNotifyClaimed handles Bob's reveal after he calls Claim().
// it calls `createMoneroWallet` to create Alice's wallet, allowing her to own the XMR.
func (s *swapState) handleNotifyClaimed(txHash string) (mcrypto.Address, error) {
	receipt, ok := common.WaitForReceipt(s.ctx, s.alice.ethClient, ethcommon.HexToHash(txHash))
	if !ok {
		return "", errors.New("failed check Claim transaction receipt")
	}

	if len(receipt.Logs) == 0 {
		return "", errors.New("claim transaction has no logs")
	}

	skB, err := swap.GetSecretFromLog(receipt.Logs[0], "Claimed")
	if err != nil {
		return "", fmt.Errorf("failed to get secret from log: %w", err)
	}

	return s.claimMonero(skB)
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

	s.success = true
	return monero.CreateMoneroWallet("alice-swap-wallet", s.alice.env, s.alice.client, kpAB)
}
