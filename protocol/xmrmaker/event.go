package xmrmaker

import (
	"fmt"

	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/net/message"
)

// getStatus returns the status corresponding to an Event.
func getStatus(t Event) types.Status {
	switch t.(type) {
	case *EventKeysReceived:
		return types.ExpectingKeys
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

type Event interface{}

type EventKeysReceived struct {
	message *message.SendKeysMessage
	errCh   chan error
}

func newEventKeysSent(msg *message.SendKeysMessage) *EventKeysReceived {
	return &EventKeysReceived{
		message: msg,
		errCh:   make(chan error),
	}
}

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

type EventContractReady struct {
	errCh chan error
}

func newEventContractReady() *EventContractReady {
	return &EventContractReady{
		errCh: make(chan error),
	}
}

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

// EventExit is sent when the protocol should be stopped, for example
// if the remote peer closes their connection with us before sending all
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
	// TODO event type checks, check that event isn't unexpected/out of order

	// events are only used once, so their error channel can be closed after handling.
	switch e := event.(type) {
	case *EventKeysReceived:
		err := s.handleSendKeysMessage(e.message)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle EventKeysReceived: %w", err)
		}
		close(e.errCh)
	case *EventETHLocked:
		err := s.handleEventETHLocked(e)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle EventETHLocked: %w", err)
		}
		close(e.errCh)
	case *EventContractReady:
		err := s.handleEventContractReady(e)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle EventContractReady: %w", err)
		}
		close(e.errCh)
	case *EventETHRefunded:
		err := s.handleEventETHRefunded(e)
		if err != nil {
			e.errCh <- fmt.Errorf("failed to handle EventETHRefunded: %w", err)
		}

		close(e.errCh)
	case *EventExit:
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

func (s *swapState) handleEventContractReady(e *EventContractReady) error {
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
