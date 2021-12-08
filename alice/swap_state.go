package alice

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/swap-contract"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color" //nolint:misspell
)

var nextID uint64 = 0

var (
	errMissingKeys    = errors.New("did not receive Bob's public spend or private view key")
	errMissingAddress = errors.New("did not receive Bob's address")
)

// swapState is an instance of a swap. it holds the info needed for the swap,
// and its current state.
type swapState struct {
	*alice
	ctx    context.Context
	cancel context.CancelFunc
	sync.Mutex

	id uint64
	// amount of ETH we are providing this swap, and the amount of XMR we should receive.
	providesAmount common.EtherAmount
	desiredAmount  common.MoneroAmount

	// our keys for this session
	privkeys *monero.PrivateKeyPair
	pubkeys  *monero.PublicKeyPair

	// Bob's keys for this session
	bobPublicSpendKey *monero.PublicKey
	bobPrivateViewKey *monero.PrivateViewKey
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
	success bool
}

func newSwapState(a *alice, providesAmount common.EtherAmount, desiredAmount common.MoneroAmount) (*swapState, error) {
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
		desiredAmount:       desiredAmount,
		txOpts:              txOpts,
		nextExpectedMessage: &net.SendKeysMessage{},
		xmrLockedCh:         make(chan struct{}),
		claimedCh:           make(chan struct{}),
	}

	nextID++
	return s, nil
}

func (s *swapState) SendKeysMessage() (*net.SendKeysMessage, error) {
	kp, err := s.generateKeys()
	if err != nil {
		return nil, err
	}

	return &net.SendKeysMessage{
		PublicSpendKey: kp.SpendKey().Hex(),
		PublicViewKey:  kp.ViewKey().Hex(),
	}, nil
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
	case *net.NotifyXMRLock:
		// we already deployed the contract, so we should call Refund().
		if err := s.tryRefund(); err != nil {
			log.Errorf("failed to refund: err=%s", err)
			return err
		}
	case *net.NotifyClaimed:
		// the XMR has been locked, but the ETH hasn't been claimed.
		// we should also refund in this case.
		if err := s.tryRefund(); err != nil {
			log.Errorf("failed to refund: err=%s", err)
			return err
		}
	default:
		log.Errorf("unexpected nextExpectedMessage in ProtocolExited: type=%T", s.nextExpectedMessage)
		return errors.New("unexpected message type")
	}

	return nil
}

func (s *swapState) tryRefund() error {
	untilT0 := time.Until(s.t0)
	untilT1 := time.Until(s.t1)

	// TODO: also check if IsReady == true

	if untilT0 > 0 && untilT1 < 0 {
		// we've passed t0 but aren't past t1 yet, so we need to wait until t1
		log.Infof("waiting until time %s to refund", s.t1)
		<-time.After(untilT1)
	}

	_, err := s.refund()
	return err
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
		vk := monero.SumPrivateViewKeys(s.bobPrivateViewKey, s.privkeys.ViewKey())
		sk := monero.SumPublicKeys(s.bobPublicSpendKey, s.pubkeys.SpendKey())
		kp := monero.NewPublicKeyPair(sk, vk.Public())

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

			if monero.Address(addr) == kp.Address(s.alice.env) {
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

		s.nextExpectedMessage = &net.NotifyClaimed{}
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
				if err = s.net.SendSwapMessage(&net.NotifyRefund{
					TxHash: txhash,
				}); err != nil {
					log.Errorf("failed to send refund message: err=%s", err)
				}
			case <-s.claimedCh:
				return
			}
		}()

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
	if msg.PublicSpendKey == "" || msg.PrivateViewKey == "" {
		return nil, errMissingKeys
	}

	if msg.EthAddress == "" {
		return nil, errMissingAddress
	}

	vk, err := monero.NewPrivateViewKeyFromHex(msg.PrivateViewKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Bob's private view keys: %w", err)
	}

	s.bobAddress = ethcommon.HexToAddress(msg.EthAddress)

	log.Debugf("got Bob's keys and address: address=%s", s.bobAddress)
	s.nextExpectedMessage = &net.NotifyXMRLock{}

	sk, err := monero.NewPublicKeyFromHex(msg.PublicSpendKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Bob's public spend key: %w", err)
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
			if err := s.net.SendSwapMessage(&net.NotifyRefund{
				TxHash: txhash,
			}); err != nil {
				log.Errorf("failed to send refund message: err=%s", err)
			}
		case <-s.xmrLockedCh:
			return
		}

	}()

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
