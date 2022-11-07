package xmrmaker

import (
	"fmt"
	"reflect"

	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/net/message"
)

// getStatus returns the status corresponding to the next expected event.
func getStatus(t Event) types.Status {
	switch t.(type) {
	case *EventETHLocked:
		return types.KeysExchanged
	case *EventContractReady:
		return types.XMRLocked
	default:
		return types.UnknownStatus
	}
}

// Event represents a swap state event.
type Event interface{}

// EventETHLocked is the first expected event. It represents ETH being locked
// on-chain.
type EventETHLocked struct {
	message *message.NotifyETHLocked
	errCh   chan error
}

func newEventETHLocked(msg *message.NotifyETHLocked) *EventETHLocked {
	return &EventETHLocked{
		message: msg,
		errCh:   make(chan error),
	}
}

// EventContractReady is the second expected event. It represents the contract being
// ready for us to claim the ETH.
type EventContractReady struct {
	errCh chan error
}

func newEventContractReady() *EventContractReady {
	return &EventContractReady{
		errCh: make(chan error),
	}
}

// EventETHRefunded is an optional event. It represents the ETH being refunded back
// to the counterparty, and thus we also must refund.
type EventETHRefunded struct {
	sk    *mcrypto.PrivateSpendKey
	errCh chan error
}

func newEventETHRefunded(sk *mcrypto.PrivateSpendKey) *EventETHRefunded {
	return &EventETHRefunded{
		sk:    sk,
		errCh: make(chan error),
	}
}

// EventExit is an optional event. It is sent when the protocol should be stopped,
// for example if the remote peer closes their connection with us before sending all
// required messages, or we decide to cancel the swap.
type EventExit struct {
	errCh chan error
}

func newEventExit() *EventExit {
	return &EventExit{
		errCh: make(chan error),
	}
}

func (s *swapState) runHandleEvents() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case event := <-s.eventCh:
			s.handleEvent(event)
		}
	}
}

func (s *swapState) handleEvent(event Event) {
	// events are only used once, so their error channel can be closed after handling.
	switch e := event.(type) {
	case *EventETHLocked:
		log.Infof("EventETHLocked")
		defer close(e.errCh)

		if reflect.TypeOf(s.nextExpectedEvent) != reflect.TypeOf(&EventETHLocked{}) {
			e.errCh <- fmt.Errorf("nextExpectedEvent was %T, not %T", s.nextExpectedEvent, e)
			return
		}

		err := s.handleEventETHLocked(e)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle EventETHLocked: %w", err)
			return
		}

		s.setNextExpectedEvent(&EventContractReady{})
	case *EventContractReady:
		log.Infof("EventContractReady")
		defer close(e.errCh)

		if reflect.TypeOf(s.nextExpectedEvent) != reflect.TypeOf(&EventContractReady{}) {
			e.errCh <- fmt.Errorf("nextExpectedEvent was %T, not %T", s.nextExpectedEvent, e)
			return
		}

		err := s.handleEventContractReady()
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle EventContractReady: %w", err)
			return
		}

		s.setNextExpectedEvent(&EventExit{})
		err = s.Exit()
		if err != nil {
			log.Warnf("failed to exit swap: %s", err)
		}
	case *EventETHRefunded:
		log.Infof("EventETHRefunded")
		defer close(e.errCh)

		err := s.handleEventETHRefunded(e)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle EventETHRefunded: %w", err)
			return
		}

		s.setNextExpectedEvent(&EventExit{})
		err = s.Exit()
		if err != nil {
			log.Warnf("failed to exit swap: %s", err)
		}
	case *EventExit:
		// this can happen at any stage.
		log.Infof("EventExit")
		defer close(e.errCh)

		err := s.exit()
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle EventExit: %w", err)
		}
	default:
		panic("unhandled event type")
	}
}

func (s *swapState) handleEventETHLocked(e *EventETHLocked) error {
	resp, err := s.handleNotifyETHLocked(e.message)
	if err != nil {
		return err
	}

	return s.SendSwapMessage(resp, s.ID())
}

func (s *swapState) handleEventContractReady() error {
	log.Debug("contract ready, attempting to claim funds...")
	close(s.readyCh)

	// contract ready, let's claim our ether
	txHash, err := s.claimFunds()
	if err != nil {
		log.Warnf("failed to claim funds from contract, attempting to safely exit: %s", err)

		// TODO: retry claim, depending on error (#162)
		if err = s.exit(); err != nil {
			return fmt.Errorf("failed to exit after failing to claim: %w", err)
		}
		return fmt.Errorf("failed to claim: %w", err)
	}

	log.Debug("funds claimed, tx: %s", txHash)
	s.clearNextExpectedEvent(types.CompletedSuccess)
	return nil
}

func (s *swapState) handleEventETHRefunded(e *EventETHRefunded) error {
	// generate monero wallet, regaining control over locked funds
	addr, err := s.reclaimMonero(e.sk)
	if err != nil {
		return err
	}

	s.clearNextExpectedEvent(types.CompletedRefund)
	log.Infof("regained control over monero account %s", addr)
	s.CloseProtocolStream(s.ID())
	return nil
}
