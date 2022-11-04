package xmrtaker

import (
	"errors"
	"strings"

	contracts "github.com/athanorlabs/atomic-swap/ethereum"

	"github.com/ethereum/go-ethereum/accounts/abi"
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

	err := s.checkSwapID(logs[0], "Claimed")
	if err == errLogNotForUs {
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

// TODO this is copypasta, move to protocol/
func (s *swapState) checkSwapID(log ethtypes.Log, eventName string) error {
	abiSF, err := abi.JSON(strings.NewReader(contracts.SwapFactoryMetaData.ABI))
	if err != nil {
		return err
	}

	data := log.Data
	res, err := abiSF.Unpack(eventName, data)
	if err != nil {
		return err
	}

	if len(res) < 1 {
		return errors.New("log had not enough parameters")
	}

	swapID := res[0].([32]byte)
	if swapID != s.contractSwapID {
		return errLogNotForUs
	}

	return nil
}
