package xmrmaker

import (
	"errors"
	"fmt"
)

var (
	// various instance and swap errors
	errUnexpectedMessageType = errors.New("unexpected message type")
	errMissingKeys           = errors.New("did not receive XMRTaker's public spend or view key")
	errMissingAddress        = errors.New("got empty contract address")
	errNoRefundLogsFound     = errors.New("no refund logs found")
	errClaimPastTime         = errors.New("past t1, can no longer claim")
	errClaimInvalid          = errors.New("can not claim, swap does not exist")
	errClaimSwapComplete     = fmt.Errorf("can not claim, %w", errSwapCompleted)
	errNilSwapState          = errors.New("swap state is nil")
	errNilMessage            = errors.New("message is nil")
	errIncorrectMessageType  = errors.New("received unexpected message")
	errNilContractSwapID     = errors.New("expected swapID in NotifyETHLocked message")
	errClaimTxHasNoLogs      = errors.New("claim transaction has no logs")
	errCannotFindNewLog      = errors.New("cannot find New log")
	errUnexpectedSwapID      = errors.New("unexpected swap ID was emitted by New log")
	errInvalidSwapContract   = errors.New("given contract address does not contain correct code")
	errSwapIDMismatch        = errors.New("hash of swap struct does not match swap ID")

	// protocol initiation errors
	errProtocolAlreadyInProgress = errors.New("protocol already in progress")
	errBalanceTooLow             = errors.New("balance lower than amount to be provided")
	errNoOfferWithID             = errors.New("failed to find offer with given ID")
	errAmountProvidedTooLow      = errors.New("amount provided by taker is too low for offer")
	errAmountProvidedTooHigh     = errors.New("amount provided by taker is too high for offer")
	errUnlockedBalanceTooLow     = errors.New("unlocked balance is less than maximum offer amount")
	errSwapCompleted             = errors.New("swap is already completed")
)
