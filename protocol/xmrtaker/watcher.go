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
		case logs := <-s.logClaimedCh:
			// contract was set to ready, send EventReady
			err := s.handleClaimedLogs(logs)
			if err != nil {
				log.Errorf("failed to handle ready logs: %s", err)
			}
		}
	}
}

func (s *swapState) handleClaimedLogs(logs []ethtypes.Log) error {
	if len(logs) == 0 {
		return errNoClaimedLogs
	}

	err := pcommon.CheckSwapID(logs[0], "Claimed", s.contractSwapID)
	if errors.Is(err, pcommon.ErrLogNotForUs) {
		return nil
	}
	if err != nil {
		return err
	}

	sk, err := contracts.GetSecretFromLog(&logs[0], "Claimed")
	if err != nil {
		return err
	}

	log.Debugf("got Claimed log: %v", logs[0])
	event := newEventETHClaimed(sk)
	s.eventCh <- event
	return <-event.errCh
}
