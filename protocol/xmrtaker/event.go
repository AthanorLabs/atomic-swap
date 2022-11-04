package xmrtaker

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/net/message"
)

// getStatus returns the status corresponding to an Event.
func getStatus(t Event) types.Status {
	switch t.(type) {
	case *EventKeysReceived:
		return types.KeysExchanged
	case *EventXMRLocked:
		return types.XMRLocked
	case *EventETHClaimed:
		return types.ContractReady
	case *EventShouldRefund:
		return types.XMRLocked
	default:
		return types.UnknownStatus
	}
}

// Event represents a swap state event.
type Event interface{}

// EventKeysReceived is the first expected event.
type EventKeysReceived struct {
	message *message.SendKeysMessage
	errCh   chan error
}

func newEventKeysReceived(msg *message.SendKeysMessage) *EventKeysReceived {
	return &EventKeysReceived{
		message: msg,
		errCh:   make(chan error),
	}
}

// EventXMRLocked is the second expected event. It represents XMR being locked
// on-chain.
type EventXMRLocked struct {
	message *message.NotifyXMRLock
	errCh   chan error
}

func newEventXMRLocked(msg *message.NotifyXMRLock) *EventXMRLocked {
	return &EventXMRLocked{
		message: msg,
		errCh:   make(chan error),
	}
}

// EventETHClaimed is the third expected event. It represents the ETH being claimed
// tbyo the counterparty, and thus we can also claim the XMR.
type EventETHClaimed struct {
	sk    *mcrypto.PrivateSpendKey
	errCh chan error
}

func newEventETHClaimed(sk *mcrypto.PrivateSpendKey) *EventETHClaimed {
	return &EventETHClaimed{
		sk:    sk,
		errCh: make(chan error),
	}
}

// EventShouldRefund is an optional event. It occurs when the XMR-maker doesn't
// lock before t0, so we should refund the ETH.
type EventShouldRefund struct {
	errCh chan error
}

func newEventShouldRefund() *EventShouldRefund {
	return &EventShouldRefund{
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
	case *EventKeysReceived:
		log.Infof("EventKeysReceived")
		defer close(e.errCh)

		if reflect.TypeOf(s.nextExpectedEvent) != reflect.TypeOf(&EventKeysReceived{}) {
			e.errCh <- fmt.Errorf("nextExpectedEvent was %T, not %T", s.nextExpectedEvent, e)
			return
		}

		err := s.handleEventKeysReceived(e)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle %T: %w", e, err)
			return
		}

		s.setNextExpectedEvent(&EventXMRLocked{})
	case *EventXMRLocked:
		log.Infof("EventXMRLocked")
		defer close(e.errCh)

		if reflect.TypeOf(s.nextExpectedEvent) != reflect.TypeOf(&EventXMRLocked{}) {
			e.errCh <- fmt.Errorf("nextExpectedEvent was %T, not %T", s.nextExpectedEvent, e)
			return
		}

		err := s.handleEventXMRLocked(e)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle %T: %w", e, err)
			return
		}

		s.setNextExpectedEvent(&EventETHClaimed{})
	case *EventETHClaimed:
		log.Infof("EventETHClaimed")
		defer close(e.errCh)

		if reflect.TypeOf(s.nextExpectedEvent) != reflect.TypeOf(&EventETHClaimed{}) {
			e.errCh <- fmt.Errorf("nextExpectedEvent was %T, not %T", s.nextExpectedEvent, e)
			return
		}

		err := s.handleEventETHClaimed(e)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle %T: %w", e, err)
		}
	case *EventShouldRefund:
		log.Infof("EventShouldRefund")
		defer close(e.errCh)

		// either EventXMRLocked or EventETHClaimed next is ok
		// if reflect.TypeOf(s.nextExpectedEvent) != reflect.TypeOf(&EventXMRLocked{}) {
		// 	e.errCh <- fmt.Errorf("nextExpectedEvent was not %T", e)
		// }

		err := s.handleEventShouldRefund()
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle %T: %w", e, err)
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

func (s *swapState) handleEventKeysReceived(event *EventKeysReceived) error {
	resp, err := s.handleSendKeysMessage(event.message)
	if err != nil {
		return err
	}

	return s.SendSwapMessage(resp, s.ID())
}

func (s *swapState) handleEventXMRLocked(event *EventXMRLocked) error {
	return s.handleNotifyXMRLock(event.message)
}

func (s *swapState) handleEventETHClaimed(event *EventETHClaimed) error {
	_, err := s.claimMonero(event.sk)
	if err != nil {
		return err
	}

	s.clearNextExpectedEvent(types.CompletedSuccess)
	s.CloseProtocolStream(s.ID())
	return nil
}

func (s *swapState) handleEventShouldRefund() error {
	// TODO could this happen still?
	if !s.info.Status().IsOngoing() {
		return nil
	}

	txhash, err := s.refund()
	if err != nil {
		// TODO could this happen either?
		if !strings.Contains(err.Error(), revertSwapCompleted) {
			return err
		}

		log.Debugf("failed to refund (okay): err=%s", err)
		return nil
	}

	log.Infof("got our ETH back: tx hash=%s", txhash)
	return nil
}
