package xmrmaker

import (
	contracts "github.com/athanorlabs/atomic-swap/ethereum"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

func (s *swapState) runContractEventWatcher() {
	select {
	case <-s.ctx.Done():
		return
	case logs := <-s.logReadyCh:
		// contract was set to ready, send EventReady
		err := s.handleReadyLogs(logs)
		if err != nil {
			log.Errorf("failed to handle ready logs: %s", err)
		}
	case logs := <-s.logRefundedCh:
		// swap was refunded, send EventRefunded
		err := s.handleRefundLogs(logs)
		if err != nil {
			// TODO what should we actually do here? this shouldn't happen ever
			log.Errorf("failed to handle refund logs: %s", err)
		}
	}
}

func (s *swapState) handleReadyLogs(logs []ethtypes.Log) error {
	if len(logs) == 0 {
		return errNoReadyLogs
	}

	event := newEventContractReady()
	s.eventCh <- event
	return <-event.errCh
}

func (s *swapState) handleRefundLogs(logs []ethtypes.Log) error {
	if len(logs) == 0 {
		return errNoRefundLogs
	}

	sk, err := contracts.GetSecretFromLog(&logs[0], "Refunded")
	if err != nil {
		return err
	}

	event := newEventETHRefunded(sk)
	s.eventCh <- event
	return <-event.errCh
}
