// Copyright 2023 The AthanorLabs/atomic-swap Authors
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

type errTokenBalanceTooLow struct {
	providedAmount *apd.Decimal // standard units
	tokenBalance   *apd.Decimal // standard units
	symbol         string
}

func (e errTokenBalanceTooLow) Error() string {
	return fmt.Sprintf("balance of %s %s is below provided %s %s",
		e.tokenBalance.Text('f'), e.symbol,
		e.providedAmount.Text('f'), e.symbol,
	)
}

func errContractAddrMismatch(addr string) error {
	//nolint:lll
	return fmt.Errorf("cannot recover from swap where contract address is not the one loaded at start-up; please restart with --contract-address=%s", addr)
}

type errAmountProvidedTooLow struct {
	providedAmtETH *apd.Decimal
	offerMinAmtETH *apd.Decimal
}

func (e errAmountProvidedTooLow) Error() string {
	return fmt.Sprintf("%s ETH provided is under offer minimum of %s ETH",
		e.providedAmtETH.Text('f'),
		e.offerMinAmtETH.Text('f'),
	)
}

type errAmountProvidedTooHigh struct {
	providedAmtETH *apd.Decimal
	offerMaxETH    *apd.Decimal
}

func (e errAmountProvidedTooHigh) Error() string {
	return fmt.Sprintf("%s ETH provided is over offer maximum of %s ETH",
		e.providedAmtETH.Text('f'),
		e.offerMaxETH.Text('f'),
	)
}

type errETHBalanceTooLow struct {
	currentBalanceETH  *apd.Decimal
	requiredBalanceETH *apd.Decimal
}

func (e errETHBalanceTooLow) Error() string {
	return fmt.Sprintf("balance of %s ETH is under required amount of %s ETH",
		e.currentBalanceETH.Text('f'),
		e.requiredBalanceETH.Text('f'),
	)
}
