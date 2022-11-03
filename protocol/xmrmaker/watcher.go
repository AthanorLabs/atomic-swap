package xmrmaker

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
			// there won't be any more events after this
			return
		}
	}
}

func (s *swapState) handleReadyLogs(logs []ethtypes.Log) error {
	if len(logs) == 0 {
		return errNoReadyLogs
	}

	err := s.checkSwapID(logs[0], "Ready")
	if err == errLogNotForUs {
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

	err := s.checkSwapID(logs[0], "Refunded")
	if err == errLogNotForUs {
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
