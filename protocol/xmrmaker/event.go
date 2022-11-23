package xmrmaker

import (
	"fmt"

	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/net/message"
)

// EventType ...
type EventType byte

const (
	EventETHLockedType EventType = iota //nolint:revive
	EventContractReadyType
	EventETHRefundedType
	EventExitType
	EventNoneType
)

func nextExpectedEventFromStatus(s types.Status) EventType {
	switch s {
	case types.ExpectingKeys, types.KeysExchanged:
		return EventETHLockedType
	case types.XMRLocked:
		return EventContractReadyType
	default:
		return EventExitType
	}
}

func (t EventType) String() string {
	switch t {
	case EventETHLockedType:
		return "EventETHLockedType"
	case EventContractReadyType:
		return "EventContractReadyType"
	case EventETHRefundedType:
		return "EventETHRefundedType"
	case EventExitType:
		return "EventExitType"
	case EventNoneType:
		return "EventNoneType"
	default:
		panic("invalid EventType")
	}
}

// getStatus returns the status corresponding to the next expected event.
func (t EventType) getStatus() types.Status {
	switch t {
	case EventETHLockedType:
		return types.KeysExchanged
	case EventContractReadyType:
		return types.XMRLocked
	default:
		return types.UnknownStatus
	}
}

// Event represents a swap state event.
type Event interface {
	Type() EventType
}

// EventETHLocked is the first expected event. It represents ETH being locked
// on-chain.
type EventETHLocked struct {
	message *message.NotifyETHLocked
	errCh   chan error
}

// Type ...
func (*EventETHLocked) Type() EventType {
	return EventETHLockedType
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

// Type ...
func (*EventContractReady) Type() EventType {
	return EventContractReadyType
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

// Type ...
func (*EventETHRefunded) Type() EventType {
	return EventETHRefundedType
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
	case *EventETHLocked:
		log.Infof("EventETHLocked")
		defer close(e.errCh)

		if s.nextExpectedEvent != EventETHLockedType {
			e.errCh <- fmt.Errorf("nextExpectedEvent was %s, not %s", s.nextExpectedEvent, e.Type())
			return
		}

		err := s.handleEventETHLocked(e)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle EventETHLocked: %w", err)
			return
		}

		err = s.setNextExpectedEvent(EventContractReadyType)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to set next expected event to EventContractReadyType: %w", err)
			return
		}
	case *EventContractReady:
		log.Infof("EventContractReady")
		defer close(e.errCh)

		if s.nextExpectedEvent != EventContractReadyType {
			e.errCh <- fmt.Errorf("nextExpectedEvent was %s, not %s", s.nextExpectedEvent, e.Type())
			return
		}

		err := s.handleEventContractReady()
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle EventContractReady: %w", err)
			return
		}

		err = s.exit()
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

		err = s.exit()
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
		if err2 := s.exit(); err2 != nil {
			return fmt.Errorf("failed to exit after failing to claim: %w", err2)
		}

		return fmt.Errorf("failed to claim: %w", err)
	}

	log.Debugf("funds claimed, tx: %s", txHash)
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
