package xmrmaker

import (
	"fmt"
	"reflect"

	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/net/message"
)

// getStatus returns the status corresponding to an Event.
func getStatus(t Event) types.Status {
	switch t.(type) {
	case *EventETHLocked:
		return types.KeysExchanged
	case *EventContractReady:
		return types.XMRLocked
	case *EventETHRefunded:
		return types.ContractReady
	default:
		return types.UnknownStatus
	}
}

// Event represents a swap state event.
type Event interface{}

// EventETHLocked is the second expected event. It represents ETH being locked
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

// EventContractReady is the third expected event. It represents the contract being
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

		if reflect.TypeOf(s.nextExpectedEvent) != reflect.TypeOf(&EventETHLocked{}) {
			e.errCh <- fmt.Errorf("nextExpectedEvent was not %T", e)
		}

		err := s.handleEventETHLocked(e)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle EventETHLocked: %w", err)
		}
		close(e.errCh)
		s.setNextExpectedEvent(&EventContractReady{})
	case *EventContractReady:
		log.Infof("EventContractReady")

		if reflect.TypeOf(s.nextExpectedEvent) != reflect.TypeOf(&EventContractReady{}) {
			e.errCh <- fmt.Errorf("nextExpectedEvent was not %T", e)
		}

		err := s.handleEventContractReady(e)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle EventContractReady: %w", err)
		}
		close(e.errCh)
		s.setNextExpectedEvent(&EventExit{})
	case *EventETHRefunded:
		log.Infof("EventETHRefunded")
		err := s.handleEventETHRefunded(e)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle EventETHRefunded: %w", err)
		}

		close(e.errCh)
		s.setNextExpectedEvent(&EventExit{})
	case *EventExit:
		log.Infof("EventExit")

		if reflect.TypeOf(s.nextExpectedEvent) != reflect.TypeOf(&EventExit{}) {
			e.errCh <- fmt.Errorf("nextExpectedEvent was not %T", e)
		}

		err := s.exit()
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle EventExit: %w", err)
		}

		close(e.errCh)
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

func (s *swapState) handleEventContractReady(_ *EventContractReady) error {
	log.Debug("contract ready, attempting to claim funds...")
	close(s.readyCh)

	// contract ready, let's claim our ether
	txHash, err := s.claimFunds()
	if err != nil {
		// TODO: retry claim, depending on error (#162)
		if err = s.exit(); err != nil {
			return fmt.Errorf("failed to exit after failing to claim: %w", err)
		}
		return fmt.Errorf("failed to claim: %w", err)
	}

	log.Debug("funds claimed, tx=%s", txHash)
	out := &message.NotifyClaimed{
		TxHash: txHash.String(),
	}

	s.clearNextExpectedEvent(types.CompletedSuccess)
	return s.SendSwapMessage(out, s.ID())
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
