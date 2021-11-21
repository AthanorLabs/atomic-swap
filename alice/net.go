package alice

import (
	"errors"
	"fmt"
	"time"

	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
)

func (a *alice) Provides() net.ProvidesCoin {
	return net.ProvidesETH
}

func (a *alice) SendKeysMessage() (*net.SendKeysMessage, error) {
	if a.swapState == nil {
		return nil, errors.New("must initiate swap before generating keys")
	}

	kp, err := a.swapState.generateKeys()
	if err != nil {
		return nil, err
	}

	return &net.SendKeysMessage{
		PublicSpendKey: kp.SpendKey().Hex(),
		PublicViewKey:  kp.ViewKey().Hex(),
	}, nil
}

func (a *alice) InitiateProtocol(providesAmount, desiredAmount uint64) (net.SwapState, error) {
	// if err := a.initiate(providesAmount, desiredAmount); err != nil {
	// 	return err
	// }

	// a.setNextExpectedMessage(&net.SendKeysMessage{})
	// return nil
	return a.initiate(providesAmount, desiredAmount)
}

func (a *alice) initiate(providesAmount, desiredAmount uint64) (*swapState, error) {
	if a.swapState != nil {
		return nil, errors.New("protocol already in progress")
	}

	balance, err := a.ethClient.BalanceAt(a.ctx, a.auth.From, nil)
	if err != nil {
		return nil, err
	}

	// check user's balance and that they actualy have what they will provide
	if balance.Uint64() <= providesAmount {
		return nil, errors.New("balance lower than amount to be provided")
	}

	a.swapState = newSwapState(a, providesAmount, desiredAmount)
	return a.swapState, nil
}

func (s *swapState) ProtocolComplete() {
	s.alice.swapState = nil
}

func (a *alice) HandleInitiateMessage(msg *net.InitiateMessage) (net.SwapState, net.Message, error) {
	if msg.Provides != net.ProvidesXMR {
		return nil, nil, errors.New("peer does not provide XMR")
	}

	// TODO: notify the user via the CLI/websockets that someone wishes to initiate a swap with them.

	// the other party initiated, saying what they will provide and what they desire.
	// we initiate our protocol, saying we will provide what they desire and vice versa.
	swapState, err := a.initiate(msg.DesiredAmount, msg.ProvidesAmount)
	if err != nil {
		return nil, nil, err
	}

	resp, err := swapState.handleSendKeysMessage(msg.SendKeysMessage)
	if err != nil {
		return nil, nil, err
	}

	return swapState, resp, nil
}

func (s *swapState) HandleProtocolMessage(msg net.Message) (net.Message, bool, error) {
	if err := s.checkMessageType(msg); err != nil {
		return nil, true, err
	}

	switch msg := msg.(type) {
	// case *net.InitiateMessage:
	// 	if msg.Provides != net.ProvidesXMR {
	// 		return nil, true, errors.New("peer does not provide XMR")
	// 	}

	// 	// TODO: notify the user via the CLI/websockets that someone wishes to initiate a swap with them.

	// 	// the other party initiated, saying what they will provide and what they desire.
	// 	// we initiate our protocol, saying we will provide what they desire and vice versa.
	// 	swapState, err := a.initiate(msg.DesiredAmount, msg.ProvidesAmount)
	// 	if err != nil {
	// 		return nil, true, err
	// 	}

	// 	resp, err := a.handleSendKeysMessage(msg.SendKeysMessage, xmrLockedCh)
	// 	if err != nil {
	// 		return nil, true, err
	// 	}

	// 	return resp, false, nil
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
			case <-s.alice.ctx.Done():
				return
			case <-time.After(until):
				// Bob hasn't claimed, and we're after t_1. let's call Refund
				if err = s.refund(); err != nil {
					log.Errorf("failed to refund: err=%s", err)
					return
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
