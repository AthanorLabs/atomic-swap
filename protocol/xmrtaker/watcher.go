// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package xmrtaker

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
		case l := <-s.logClaimedCh:
			eventSent, err := s.handleClaimedLogs(&l)
			if err != nil {
				log.Errorf("failed to handle ready logs: %s", err)
			}

			if eventSent {
				log.Debugf("EventETHClaimed sent, returning from event watcher")
				return
			}
		}
	}
}

func (s *swapState) handleClaimedLogs(l *ethtypes.Log) (bool, error) {
	err := pcommon.CheckSwapID(l, claimedTopic, s.contractSwapID)
	if errors.Is(err, pcommon.ErrLogNotForUs) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	sk, err := contracts.GetSecretFromLog(l, claimedTopic)
	if err != nil {
		return false, err
	}

	// contract was set to ready, send EventReady
	event := newEventETHClaimed(sk)
	s.eventCh <- event
	return true, <-event.errCh
}
