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
		case l := <-s.logReadyCh:
			err := s.handleReadyLogs(&l)
			if err != nil {
				log.Errorf("failed to handle ready logs: %s", err)
			}
		case l := <-s.logRefundedCh:
			err := s.handleRefundLogs(&l)
			if err != nil {
				log.Errorf("failed to handle refund logs: %s", err)
			}

			// there won't be any more events after this
			return
		}
	}
}

func (s *swapState) handleReadyLogs(l *ethtypes.Log) error {
	log.Infof("got ready log %v", l)

	err := pcommon.CheckSwapID(l, "Ready", s.contractSwapID)
	if errors.Is(err, pcommon.ErrLogNotForUs) {
		log.Infof("log not for us")
		return nil
	}
	if err != nil {
		return err
	}

	// contract was set to ready, send EventReady
	event := newEventContractReady()
	s.eventCh <- event
	return <-event.errCh
}

func (s *swapState) handleRefundLogs(log *ethtypes.Log) error {
	err := pcommon.CheckSwapID(log, "Refunded", s.contractSwapID)
	if errors.Is(err, pcommon.ErrLogNotForUs) {
		return nil
	}
	if err != nil {
		return err
	}

	sk, err := contracts.GetSecretFromLog(log, "Refunded")
	if err != nil {
		return err
	}

	// swap was refunded, send EventRefunded
	event := newEventETHRefunded(sk)
	s.eventCh <- event
	return <-event.errCh
}
