package xmrmaker

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/athanorlabs/atomic-swap/common"
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
	// errInvalidSwapContract         = errors.New("given contract address does not contain correct code")
	errSwapIDMismatch              = errors.New("hash of swap struct does not match swap ID")
	errLockTxReverted              = errors.New("other party failed to lock ETH asset (transaction reverted)")
	errInvalidETHLockedTransaction = errors.New("eth locked tx was not to correct contract address")
	errRelayerCommissionTooHigh    = errors.New("relayer commission must be less than 0.1 (10%)")

	// protocol initiation errors
	errProtocolAlreadyInProgress = errors.New("protocol already in progress")
	errOfferIDNotSet             = errors.New("offer ID was not set")
	errSwapCompleted             = errors.New("swap is already completed")
)

type errBalanceTooLow struct {
	unlockedBalance float64
	providedAmount  float64
}

func (e errBalanceTooLow) Error() string {
	return fmt.Sprintf("balance of %s XMR is below provided %s XMR",
		common.FmtFloat(e.unlockedBalance),
		common.FmtFloat(e.providedAmount),
	)
}

type errAmountProvidedTooLow struct {
	providedAmount float64
	minAmount      float64
}

func (e errAmountProvidedTooLow) Error() string {
	return fmt.Sprintf("%s XMR provided by taker is under offer minimum of %s XMR",
		common.FmtFloat(e.providedAmount),
		common.FmtFloat(e.minAmount),
	)
}

type errAmountProvidedTooHigh struct {
	providedAmount float64
	maxAmount      float64
}

func (e errAmountProvidedTooHigh) Error() string {
	return fmt.Sprintf("%s XMR provided by taker is over offer maximum of %s XMR",
		common.FmtFloat(e.providedAmount),
		common.FmtFloat(e.maxAmount),
	)
}

type errUnlockedBalanceTooLow struct {
	minAmount       float64
	unlockedBalance float64
}

func (e errUnlockedBalanceTooLow) Error() string {
	return fmt.Sprintf("balance %s XMR is too low for maximum offer amount of %s XMR",
		strconv.FormatFloat(e.minAmount, 'f', -1, 64),
		strconv.FormatFloat(e.unlockedBalance, 'f', -1, 64),
	)
}
