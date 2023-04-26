// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package xmrmaker

import (
	"errors"

	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

func (s *swapState) runContractEventWatcher() {
	readyEventSent := false
	for {
		select {
		case <-s.ctx.Done():
			return
		case l := <-s.logReadyCh:
			if readyEventSent {
				// we already sent the ready event, ignore any Ready logs
				continue
			}

			eventSent, err := s.handleReadyLogs(&l)
			if err != nil {
				log.Errorf("failed to handle ready logs: %s", err)
			}

			readyEventSent = eventSent
		case l := <-s.logRefundedCh:
			eventSent, err := s.handleRefundLogs(&l)
			if err != nil {
				log.Errorf("failed to handle refund logs: %s", err)
			}

			if eventSent {
				log.Debugf("EventETHRefunded sent, returning from event watcher")
				return
			}
		}
	}
}

func (s *swapState) handleReadyLogs(l *ethtypes.Log) (bool, error) {
	err := pcommon.CheckSwapID(l, readyTopic, s.contractSwapID)
	if errors.Is(err, pcommon.ErrLogNotForUs) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// contract was set to ready, send EventReady
	event := newEventContractReady()
	s.eventCh <- event
	go func() {
		err = <-event.errCh
		if err != nil {
			log.Errorf("failed to handle EventReady: %s", err)
		}
	}()
	return true, nil
}

func (s *swapState) handleRefundLogs(ethlog *ethtypes.Log) (bool, error) {
	err := pcommon.CheckSwapID(ethlog, refundedTopic, s.contractSwapID)
	if errors.Is(err, pcommon.ErrLogNotForUs) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	sk, err := contracts.GetSecretFromLog(ethlog, refundedTopic)
	if err != nil {
		return false, err
	}

	// swap was refunded, send EventRefunded
	event := newEventETHRefunded(sk)
	s.eventCh <- event
	return true, <-event.errCh
}
