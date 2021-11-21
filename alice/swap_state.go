package alice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/swap-contract"
)

var nextID uint64 = 0

// swapState is an instance of a swap. it holds the info needed for the swap,
// and its current state.
type swapState struct {
	*alice
	ctx    context.Context
	cancel context.CancelFunc

	id uint64
	// amount of ETH we are providing this swap, and the amount of XMR we should receive.
	providesAmount, desiredAmount uint64

	// our keys for this session
	privkeys *monero.PrivateKeyPair
	pubkeys  *monero.PublicKeyPair

	// Bob's keys for this session
	bobSpendKey *monero.PublicKey
	bobViewKey  *monero.PrivateViewKey

	// swap contract and timeouts in it; set once contract is deployed
	contract *swap.Swap
	t0, t1   time.Time

	// next expected network message
	nextExpectedMessage net.Message // TODO: change to type?

	// channels
	xmrLockedCh chan struct{}
	claimedCh   chan struct{}

	// set to true upon creating of the XMR wallet
	success bool
}

func newSwapState(a *alice, providesAmount, desiredAmount uint64) *swapState {
	ctx, cancel := context.WithCancel(a.ctx)

	s := &swapState{
		ctx:                 ctx,
		cancel:              cancel,
		alice:               a,
		id:                  nextID,
		providesAmount:      providesAmount,
		desiredAmount:       desiredAmount,
		nextExpectedMessage: &net.SendKeysMessage{}, // should this be &net.InitiateMessage{}?
		xmrLockedCh:         make(chan struct{}),
		claimedCh:           make(chan struct{}),
	}

	nextID++
	return s
}

// ProtocolComplete is called by the network when the protocol stream closes.
// If it closes prematurely, we need to perform recovery.
func (s *swapState) ProtocolComplete() {
	// stop all running goroutines
	s.cancel()

	defer func() {
		s.alice.swapState = nil
	}()

	if s.success {
		return
	}

	switch s.nextExpectedMessage.(type) {
	case *net.SendKeysMessage:
		// we are fine, as we only just initiated the protocol.
	case *net.NotifyXMRLock:
		// we already deployed the contract, so we should call Refund().

		// TODO: check t0 and t1
		if err := s.refund(); err != nil {
			log.Errorf("failed to refund: err=%s", err)
			return
		}
	case *net.NotifyClaimed:
		// the XMR has been locked, but the ETH hasn't been claimed.
		// we should also refund in this case.

		// TODO: check t0 and t1
		if err := s.refund(); err != nil {
			log.Errorf("failed to refund: err=%s", err)
			return
		}
	default:
		log.Errorf("unexpected nextExpectedMessage in ProtocolComplete: type=%T", s.nextExpectedMessage)
	}

}

// HandleProtocolMessage is called by the network to handle an incoming message.
// If the message received is not the expected type for the point in the protocol we're at,
// this function will return an error.
func (s *swapState) HandleProtocolMessage(msg net.Message) (net.Message, bool, error) {
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

		// TODO: check that XMR was locked in expected account, and confirm amount
		s.nextExpectedMessage = &net.NotifyClaimed{}
		close(s.xmrLockedCh)

		if err := s.ready(); err != nil {
			return nil, true, fmt.Errorf("failed to call Ready: %w", err)
		}

		log.Debug("set swap.IsReady == true")

		go func() {
			st1, err := s.contract.Timeout1(s.alice.callOpts)
			if err != nil {
				log.Errorf("failed to get timeout1 from contract: err=%s", err)
				return
			}

			t1 := time.Unix(st1.Int64(), 0)
			until := time.Until(t1)

			select {
			case <-s.ctx.Done():
				return
			case <-time.After(until):
				// Bob hasn't claimed, and we're after t_1. let's call Refund
				if err = s.refund(); err != nil {
					log.Errorf("failed to refund: err=%s", err)
					return
				}

				// TODO: send NotifyRefund msg
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

func (s *swapState) handleSendKeysMessage(msg *net.SendKeysMessage) (net.Message, error) {
	if msg.PublicSpendKey == "" || msg.PrivateViewKey == "" {
		return nil, errors.New("did not receive Bob's public spend or private view key")
	}

	log.Debug("got Bob's keys")
	s.nextExpectedMessage = &net.NotifyXMRLock{}

	sk, err := monero.NewPublicKeyFromHex(msg.PublicSpendKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Bob's public spend key: %w", err)
	}

	vk, err := monero.NewPrivateViewKeyFromHex(msg.PrivateViewKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Bob's private view keys: %w", err)
	}

	s.setBobKeys(sk, vk)
	address, err := s.deployAndLockETH(s.providesAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy contract: %w", err)
	}

	log.Info("deployed Swap contract, waiting for XMR to be locked: address=", address)

	// start goroutine to check that Bob locks before t_0
	go func() {
		const timeoutBuffer = time.Minute * 5

		st0, err := s.contract.Timeout0(s.alice.callOpts)
		if err != nil {
			log.Errorf("failed to get timeout0 from contract: err=%s", err)
			return
		}

		t0 := time.Unix(st0.Int64(), 0)
		until := time.Until(t0)

		select {
		case <-s.ctx.Done():
			return
		case <-time.After(until - timeoutBuffer):
			// Bob hasn't locked yet, let's call refund
			if err = s.refund(); err != nil {
				log.Errorf("failed to refund: err=%s", err)
				return
			}

			// TODO: send NotifyRefund msg
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
