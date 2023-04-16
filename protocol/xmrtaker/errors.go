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
	errBalanceTooLow             = errors.New("eth balance lower than amount to be provided")
	errInvalidStageForRecovery   = errors.New("cannot create ongoing swap state if stage is not ETHLocked or ContractReady") //nolint:lll
)

func errContractAddrMismatch(addr string) error {
	//nolint:lll
	return fmt.Errorf("cannot recover from swap where contract address is not the one loaded at start-up; please restart with --contract-address=%s", addr)
}

type errAmountProvidedTooLow struct {
	providedAmount *apd.Decimal
	minAmount      *apd.Decimal
}

func (e errAmountProvidedTooLow) Error() string {
	return fmt.Sprintf("%s ETH provided is under offer minimum of %s XMR",
		e.providedAmount.String(),
		e.minAmount.String(),
	)
}

type errAmountProvidedTooHigh struct {
	providedAmount *apd.Decimal
	maxAmount      *apd.Decimal
}

func (e errAmountProvidedTooHigh) Error() string {
	return fmt.Sprintf("%s ETH provided is over offer maximum of %s XMR",
		e.providedAmount.String(),
		e.maxAmount.String(),
	)
}
