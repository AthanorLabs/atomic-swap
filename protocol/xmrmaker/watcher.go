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
	err := pcommon.CheckSwapID(l, readyTopic, s.contractSwapID)
	if errors.Is(err, pcommon.ErrLogNotForUs) {
		return nil
	}
	if err != nil {
		return err
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
	return nil
}

func (s *swapState) handleRefundLogs(ethlog *ethtypes.Log) error {
	err := pcommon.CheckSwapID(ethlog, refundedTopic, s.contractSwapID)
	if errors.Is(err, pcommon.ErrLogNotForUs) {
		return nil
	}
	if err != nil {
		return err
	}

	sk, err := contracts.GetSecretFromLog(ethlog, refundedTopic)
	if err != nil {
		return err
	}

	// swap was refunded, send EventRefunded
	event := newEventETHRefunded(sk)
	s.eventCh <- event
	return <-event.errCh
}
