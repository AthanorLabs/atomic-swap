package xmrmaker

import (
	"errors"
	"fmt"

	"github.com/cockroachdb/apd/v3"
)

var (
	// various instance and swap errors
	errUnexpectedMessageType         = errors.New("unexpected message type")
	errMissingKeys                   = errors.New("did not receive XMRTaker's public spend or view key")
	errMissingAddress                = errors.New("got empty contract address")
	errNilSwapState                  = errors.New("swap state is nil")
	errNilContractSwapID             = errors.New("expected swapID in NotifyETHLocked message")
	errCannotFindNewLog              = errors.New("cannot find New log")
	errUnexpectedSwapID              = errors.New("unexpected swap ID was emitted by New log")
	errSwapIDMismatch                = errors.New("hash of swap struct does not match swap ID")
	errLockTxReverted                = errors.New("other party failed to lock ETH asset (transaction reverted)")
	errInvalidETHLockedTransaction   = errors.New("eth locked tx was not to correct contract address")
	errInvalidT0                     = errors.New("invalid t0 value; asset was locked too far in the past")
	errInvalidT1                     = errors.New("invalid swap timeout set by counterparty")
	errRelayedTransactionTimeout     = errors.New("relayed transaction was not included within one minute")
	errClaimedLogInvalidContractAddr = errors.New("log was not emitted by correct contract")
	errClaimedLogWrongTopicLength    = errors.New("log did not have 3 topics")
	errClaimedLogWrongEvent          = errors.New("log did not have the Claimed event as its first topic")
	errClaimedLogWrongSwapID         = errors.New("log did not have the correct swap ID as its second topic")
	errClaimedLogWrongSecret         = errors.New("log did not have the correct secret as its third topic")
	errRelayingWithNonEthAsset       = errors.New("relayers with ERC20 token swaps are not currently supported")

	// protocol initiation errors
	errSwapDoesNotExist          = errors.New("contract swap ID does not exist")
	errProtocolAlreadyInProgress = errors.New("protocol already in progress")
	errOfferIDNotSet             = errors.New("offer ID was not set")
	errInvalidStageForRecovery   = errors.New("cannot create ongoing swap state if stage is not XMRLocked")
)

type errBalanceTooLow struct {
	unlockedBalance *apd.Decimal
	providedAmount  *apd.Decimal
}

func (e errBalanceTooLow) Error() string {
	return fmt.Sprintf("balance of %s XMR is below provided %s XMR",
		e.unlockedBalance.String(),
		e.providedAmount.String(),
	)
}

type errAmountProvidedTooLow struct {
	providedAmount *apd.Decimal
	minAmount      *apd.Decimal
}

func (e errAmountProvidedTooLow) Error() string {
	return fmt.Sprintf("%s ETH provided by taker is under offer minimum of %s XMR",
		e.providedAmount.String(),
		e.minAmount.String(),
	)
}

type errAmountProvidedTooHigh struct {
	providedAmount *apd.Decimal
	maxAmount      *apd.Decimal
}

func (e errAmountProvidedTooHigh) Error() string {
	return fmt.Sprintf("%s ETH provided by taker is over offer maximum of %s XMR",
		e.providedAmount.String(),
		e.maxAmount.String(),
	)
}

type errUnlockedBalanceTooLow struct {
	maxOfferAmount  *apd.Decimal
	unlockedBalance *apd.Decimal
}

func (e errUnlockedBalanceTooLow) Error() string {
	return fmt.Sprintf("balance %s XMR is too low for maximum offer amount of %s XMR",
		e.unlockedBalance.String(),
		e.maxOfferAmount.String(),
	)
}
