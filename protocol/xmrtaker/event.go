package xmrtaker

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/net/message"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

// EventType ...
type EventType byte

const (
	EventKeysReceivedType EventType = iota //nolint:revive
	EventXMRLockedType
	EventETHClaimedType
	EventShouldRefundType
	EventExitType
	EventNoneType
)

// getStatus returns the status corresponding to the next expected event.
func getStatus(t EventType) types.Status {
	switch t {
	case EventXMRLockedType:
		return types.ETHLocked
	case EventETHClaimedType:
		return types.ContractReady
	default:
		return types.UnknownStatus
	}
}

// Event represents a swap state event.
type Event interface {
	Type() EventType
}

// EventKeysReceived is the first expected event.
type EventKeysReceived struct {
	message *message.SendKeysMessage
	errCh   chan error
}

// Type ...
func (*EventKeysReceived) Type() EventType {
	return EventKeysReceivedType
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

// Type ...
func (*EventXMRLocked) Type() EventType {
	return EventXMRLockedType
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

// Type ...
func (*EventETHClaimed) Type() EventType {
	return EventETHClaimedType
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
	errCh    chan error
	txHashCh chan ethcommon.Hash // contains the refund tx hash, if successful
}

// Type ...
func (*EventShouldRefund) Type() EventType {
	return EventShouldRefundType
}

func newEventShouldRefund() *EventShouldRefund {
	return &EventShouldRefund{
		errCh:    make(chan error),
		txHashCh: make(chan ethcommon.Hash, 1),
	}
}

// EventExit is an optional event. It is sent when the protocol should be stopped,
// for example if the remote peer closes their connection with us before sending all
// required messages, or we decide to cancel the swap.
type EventExit struct {
	errCh chan error
}

// Type ...
func (*EventExit) Type() EventType {
	return EventExitType
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

		if s.nextExpectedEvent != EventKeysReceivedType {
			e.errCh <- fmt.Errorf("nextExpectedEvent was %T, not %T", s.nextExpectedEvent, e.Type())
			return
		}

		err := s.handleEventKeysReceived(e)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle %T: %w", e, err)
			return
		}

		s.setNextExpectedEvent(EventXMRLockedType)
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

		s.setNextExpectedEvent(EventETHClaimedType)
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
		defer close(e.txHashCh)

		// either EventXMRLocked or EventETHClaimed next is ok
		if (reflect.TypeOf(s.nextExpectedEvent) != reflect.TypeOf(&EventXMRLocked{}) &&
			reflect.TypeOf(s.nextExpectedEvent) != reflect.TypeOf(&EventETHClaimed{})) {
			e.errCh <- fmt.Errorf("nextExpectedEvent was not %T", e)
		}

		err := s.handleEventShouldRefund(e)
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

func (s *swapState) handleEventShouldRefund(event *EventShouldRefund) error {
	if !s.info.Status.IsOngoing() {
		return nil
	}

	txHash, err := s.refund()
	if err != nil {
		// TODO: could this ever happen anymore?
		if !strings.Contains(err.Error(), revertSwapCompleted) {
			return err
		}

		log.Debugf("failed to refund (okay): err=%s", err)
		return nil
	}

	log.Infof("got our ETH back: tx hash=%s", txHash)
	event.txHashCh <- txHash
	return nil
}
