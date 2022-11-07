package xmrtaker

import (
	"errors"
	"fmt"
)

var (
	// various instance and swap errors
	errNoOngoingSwap           = errors.New("no ongoing swap with given offer ID")
	errSenderIsNotExternal     = errors.New("swap is not using an external transaction sender")
	errUnexpectedMessageType   = errors.New("unexpected message type")
	errUnexpectedEventType     = errors.New("unexpected event type")
	errMissingKeys             = errors.New("did not receive XMRMaker's public spend or private view key")
	errMissingAddress          = errors.New("did not receive XMRMaker's address")
	errNoClaimLogsFound        = errors.New("no Claimed logs found")
	errCannotRefund            = errors.New("swap is not at a stage where it can refund")
	errRefundInvalid           = errors.New("can not refund, swap does not exist")
	errRefundSwapCompleted     = fmt.Errorf("can not refund, %w", errSwapCompleted)
	errNoLockedXMRAddress      = errors.New("got empty address for locked XMR")
	errNoPublicKeysSet         = errors.New("our public keys aren't set")
	errCounterpartyKeysNotSet  = errors.New("counterparty's keys aren't set")
	errSwapInstantiationNoLogs = errors.New("expected 1 log, got 0")
	errSwapCompleted           = errors.New("swap is already completed")
	errNoClaimedLogs           = errors.New("got Claimed event from watcher with no logs")

	// inititation errors
	errProtocolAlreadyInProgress = errors.New("protocol already in progress")
	errBalanceTooLow             = errors.New("eth balance lower than amount to be provided")
	errNoSwapContractSet         = errors.New("no swap contract found")
	errMustProvideWalletAddress  = errors.New("must provide wallet address if transfer back is set")
)
