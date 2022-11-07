package xmrmaker

import (
	"errors"

	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

func (s *swapState) runContractEventWatcher() {
	for {
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
				log.Errorf("failed to handle refund logs: %s", err)
			}

			// there won't be any more events after this
			return
		}
	}
}

func (s *swapState) handleReadyLogs(logs []ethtypes.Log) error {
	if len(logs) == 0 {
		return errNoReadyLogs
	}

	err := pcommon.CheckSwapID(logs[0], "Ready", s.contractSwapID)
	if errors.Is(err, pcommon.ErrLogNotForUs) {
		return nil
	}
	if err != nil {
		return err
	}

	log.Debugf("got Ready log: %v", logs[0])
	event := newEventContractReady()
	s.eventCh <- event
	return <-event.errCh
}

func (s *swapState) handleRefundLogs(logs []ethtypes.Log) error {
	if len(logs) == 0 {
		return errNoRefundLogs
	}

	err := pcommon.CheckSwapID(logs[0], "Refunded", s.contractSwapID)
	if errors.Is(err, pcommon.ErrLogNotForUs) {
		return nil
	}
	if err != nil {
		return err
	}

	sk, err := contracts.GetSecretFromLog(&logs[0], "Refunded")
	if err != nil {
		return err
	}

	log.Debugf("got Refunded log: %v", logs[0])
	event := newEventETHRefunded(sk)
	s.eventCh <- event
	return <-event.errCh
}
