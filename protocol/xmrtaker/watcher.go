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
			err := s.handleClaimedLogs(&l)
			if err != nil {
				log.Errorf("failed to handle ready logs: %s", err)
			}
		}
	}
}

func (s *swapState) handleClaimedLogs(l *ethtypes.Log) error {
	err := pcommon.CheckSwapID(l, "Claimed", s.contractSwapID)
	if errors.Is(err, pcommon.ErrLogNotForUs) {
		return nil
	}
	if err != nil {
		return err
	}

	sk, err := contracts.GetSecretFromLog(l, "Claimed")
	if err != nil {
		return err
	}

	// contract was set to ready, send EventReady
	event := newEventETHClaimed(sk)
	s.eventCh <- event
	return <-event.errCh
}
