// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package xmrtaker

import (
	"errors"
	"fmt"

	"github.com/cockroachdb/apd/v3"
)

var (
	// various instance and swap errors
	errNoOngoingSwap           = errors.New("no ongoing swap with given offer ID")
	errSenderIsNotExternal     = errors.New("swap is not using an external transaction sender")
	errUnexpectedMessageType   = errors.New("unexpected message type")
	errUnexpectedEventType     = errors.New("unexpected event type")
	errMissingKeys             = errors.New("did not receive XMRMaker's public spend or private view key")
	errMissingProvidedAmount   = errors.New("did not receive provided amount")
	errMissingAddress          = errors.New("did not receive XMRMaker's address")
	errNoClaimLogsFound        = errors.New("no Claimed logs found")
	errRefundInvalid           = errors.New("cannot refund, swap does not exist")
	errRefundSwapCompleted     = fmt.Errorf("cannot refund, %w", errSwapCompleted)
	errCounterpartyKeysNotSet  = errors.New("counterparty's keys aren't set")
	errSwapInstantiationNoLogs = errors.New("expected 1 log, got 0")
	errSwapCompleted           = errors.New("swap is already completed")

	// initiation errors
	errProtocolAlreadyInProgress = errors.New("protocol already in progress")
	errInvalidStageForRecovery   = errors.New("cannot create ongoing swap state if stage is not ETHLocked or ContractReady") //nolint:lll
)

type errAssetBalanceTooLow struct {
	providedAmount *apd.Decimal
	balance        *apd.Decimal
	symbol         string
}

func (e errAssetBalanceTooLow) Error() string {
	return fmt.Sprintf("balance of %s %s is below provided %s %s",
		e.balance.Text('f'), e.symbol,
		e.providedAmount.Text('f'), e.symbol,
	)
}

func errContractAddrMismatch(addr string) error {
	//nolint:lll
	return fmt.Errorf("cannot recover from swap where contract address is not the one loaded at start-up; please restart with --contract-address=%s", addr)
}
