// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package xmrtaker

import (
	"fmt"
	"strings"

	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/net/message"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

// EventType represents an event that occurs which moves the swap
// "state machine" to its next state.
type EventType byte

const (
	// EventKeysReceivedType is triggered when we receive the XMR maker's
	// swap keys, allowing us to initiate the swap on-chain. It causes us to
	// lock our ETH and store keys in the smart contract. After this event,
	// the other possible events are EventXMRLockedType (success path) or
	// EventExitType (abort path).
	EventKeysReceivedType EventType = iota

	// EventXMRLockedType is triggered after we verify that the maker locked the
	// XMR for the swap. It causes us to set the contract to "ready", so that
	// the maker can claim his ETH. After this event, the other possible events
	// are EventETHClaimedType (success path), EventShouldRefundType (refund
	// path), or EventExitType (refund path).
	EventXMRLockedType

	// EventETHClaimedType is triggered when the maker claims their ETH from the
	// contract. It causes us to claim the XMR. After this event, the other
	// possible event is EventExitType (success path).
	EventETHClaimedType

	// EventShouldRefundType is triggered when we should refund, either because
	// we are nearing the timeout0 threshold and the maker hasn't locked XMR, or
	// because we've reached the timeout1 threshold and the maker hasn't claimed
	// the ETH. It causes us to refund the contract locked ETH locked to
	// ourselves. After this event, the only possible event is EventExitType
	// (refund path).
	//
	// Note: this constant is a placeholder for clarity. While the event it
	// represents is used, we never actually use the constant for its type.
	EventShouldRefundType

	// EventExitType is triggered by the protocol "exiting", which may happen
	// via a swap cancellation via the RPC endpoint, or from the counterparty
	// disconnecting from us on the p2p network. It causes us to attempt to
	// gracefully exit from the swap, which causes either an abort, refund, or
	// claim, depending on the state we're currently in. No other events can
	// occur after this.
	EventExitType

	// EventNoneType is set as the "nextExpectedEvent" once the swap has exited.
	// It does not trigger any action. No other events can occur after this.
	EventNoneType
)

func nextExpectedEventFromStatus(s types.Status) EventType {
	switch s {
	case types.ExpectingKeys:
		return EventKeysReceivedType
	case types.ETHLocked:
		return EventXMRLockedType
	case types.ContractReady:
		return EventETHClaimedType
	default:
		return EventExitType
	}
}

func (t EventType) String() string {
	switch t {
	case EventKeysReceivedType:
		return "EventKeysReceivedType"
	case EventXMRLockedType:
		return "EventXMRLockedType"
	case EventETHClaimedType:
		return "EventETHClaimedType"
	case EventShouldRefundType:
		return "EventShouldRefundType"
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
	case EventXMRLockedType:
		return types.ETHLocked
	case EventETHClaimedType:
		return types.ContractReady
	default:
		// the only possible nextExpectedEvents are EventXMRLockedType
		// and EventETHClaimedType, so this case shouldn't be hit.
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
	errCh chan error
}

// Type ...
func (*EventXMRLocked) Type() EventType {
	return EventXMRLockedType
}

func newEventXMRLocked() *EventXMRLocked {
	return &EventXMRLocked{
		errCh: make(chan error),
	}
}

// EventETHClaimed is the third expected event. It represents the ETH being claimed
// by the counterparty, and thus we can also claim the XMR.
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
			e.errCh <- fmt.Errorf("nextExpectedEvent was %s, not %s", s.nextExpectedEvent, e.Type())
			return
		}

		err := s.handleEventKeysReceived(e)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle %s: %w", e.Type(), err)
			if !s.fundsLocked {
				return
			}
		}

		err = s.setNextExpectedEvent(EventXMRLockedType)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to set next expected event to EventXMRLockedType: %w", err)
			return
		}
	case *EventXMRLocked:
		log.Infof("EventXMRLocked")
		defer close(e.errCh)

		if s.nextExpectedEvent != EventXMRLockedType {
			e.errCh <- fmt.Errorf("nextExpectedEvent was %s, not %s", s.nextExpectedEvent, e.Type())
			return
		}

		err := s.handleNotifyXMRLock()
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle %s: %w", e.Type(), err)
			return
		}

		err = s.setNextExpectedEvent(EventETHClaimedType)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to set next expected event to EventETHClaimedType: %w", err)
			return
		}
	case *EventETHClaimed:
		log.Infof("EventETHClaimed")
		defer close(e.errCh)

		if s.nextExpectedEvent != EventETHClaimedType {
			e.errCh <- fmt.Errorf("nextExpectedEvent was %s, not %s", s.nextExpectedEvent, e.Type())
			return
		}

		err := s.handleEventETHClaimed(e)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle %s: %w", e.Type(), err)
		}
	case *EventShouldRefund:
		log.Infof("EventShouldRefund")
		defer close(e.errCh)
		defer close(e.txHashCh)

		// either EventXMRLocked or EventETHClaimed next is ok
		if s.nextExpectedEvent != EventXMRLockedType &&
			s.nextExpectedEvent != EventETHClaimedType &&
			s.nextExpectedEvent != EventKeysReceivedType {
			e.errCh <- fmt.Errorf("nextExpectedEvent was %s", e.Type())
		}

		err := s.handleEventShouldRefund(e)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle %s: %w", e.Type(), err)
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

func (s *swapState) handleEventKeysReceived(event *EventKeysReceived) error {
	resp, err := s.handleSendKeysMessage(event.message)
	if err != nil {
		return err
	}

	return s.SendSwapMessage(resp, s.OfferID())
}

func (s *swapState) handleEventETHClaimed(event *EventETHClaimed) error {
	_, err := s.claimMonero(event.sk)
	if err != nil {
		return err
	}

	s.clearNextExpectedEvent(types.CompletedSuccess)
	s.CloseProtocolStream(s.OfferID())
	return nil
}

func (s *swapState) handleEventShouldRefund(event *EventShouldRefund) error {
	if !s.info.Status.IsOngoing() {
		return nil
	}

	receipt, err := s.refund()
	if err != nil {
		// TODO: could this ever happen anymore?
		if !strings.Contains(err.Error(), revertSwapCompleted) {
			return err
		}

		log.Debugf("failed to refund (okay): err=%s", err)
		return nil
	}

	log.Infof("got our ETH back: tx hash=%s", receipt.TxHash)
	event.txHashCh <- receipt.TxHash
	return nil
}
