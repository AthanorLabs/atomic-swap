// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package xmrmaker

import (
	"fmt"

	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/net/message"
)

// EventType represents an event that occurs which moves the swap
// "state machine" to its next state.
type EventType byte

const (
	// EventETHLockedType is triggered when the taker notifies us that the ETH
	// is locked in the smart contract. Upon verification, it causes us to lock
	// our XMR. After this event, the other possible events are
	// EventContractReadyType (success), EventETHRefundedType (abort), or
	// EventExitType (abort).
	EventETHLockedType EventType = iota

	// EventContractReadyType is triggered when the taker sets the contract to
	// "ready" or timeout0 is reached. When this event occurs, we can claim ETH
	// from the contract. After this event, the other possible events are
	// EventETHRefundedType (which would only happen if we go offline until
	// timeout1, causing us to refund), or EventExitType (refund).
	EventContractReadyType

	// EventETHRefundedType is triggered when the taker refunds the
	// contract-locked ETH back to themselves. It causes use to try to refund
	// our XMR. After this event, the only possible event is EventExitType.
	EventETHRefundedType

	// EventExitType is triggered by the protocol "exiting", which may happen
	// via a swap cancellation via the RPC endpoint, or from the counterparty
	// disconnecting from us on the p2p network. It causes us to attempt to
	// gracefully exit from the swap, which leads to an abort, refund, or claim,
	// depending on the state we're currently in. No other events can occur
	// after this.
	EventExitType

	// EventNoneType is set as the "nextExpectedEvent" once the swap has exited.
	// It does not trigger any action. No other events can occur after this.
	EventNoneType
)

// nextExpectedEventFromStatus returns the next expected event given the current
// swap status.
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
		// the only possible nextExpectedEvents are EventETHLockedType
		// and EventContractReadyType, so this case shouldn't be hit.
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

		err := s.handleNotifyETHLocked(e.message)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle EventETHLocked: %w", err)
			err = s.exit()
			if err != nil {
				log.Warnf("failed to exit swap: %s", err)
			}
		}

		// close the stream to the remote peer, since we won't be
		// receiving any more messages.
		s.Backend.CloseProtocolStream(s.OfferID())

		// nextExpectedEvent was set in s.lockFunds()
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

func (s *swapState) handleEventContractReady() error {
	log.Debug("contract ready, attempting to claim funds...")
	close(s.readyCh)
	s.readyWatcher.Stop()

	// contract ready, let's claim our ether
	receipt, err := s.claimFunds()
	if err != nil {
		log.Warnf("failed to claim funds from contract, attempting to safely exit: %s", err)

		// TODO: retry claim, depending on error (#162)
		if err2 := s.exit(); err2 != nil {
			return fmt.Errorf("failed to exit after failing to claim: %w", err2)
		}

		return fmt.Errorf("failed to claim: %w", err)
	}

	log.Debugf("funds claimed, tx: %s", receipt.TxHash)
	s.clearNextExpectedEvent(types.CompletedSuccess)
	return nil
}

func (s *swapState) handleEventETHRefunded(e *EventETHRefunded) error {
	// generate monero wallet, regaining control over locked funds
	err := s.reclaimMonero(e.sk)
	if err != nil {
		return err
	}

	s.clearNextExpectedEvent(types.CompletedRefund)
	return nil
}
