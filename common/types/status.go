// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package types is for types that are shared by multiple packages
package types

import (
	"fmt"
)

// Status represents the stage that a swap is at.
type Status byte

// Status values
const (
	// UnknownStatus is a placeholder for unmatched status strings and
	// uninitialized variables
	UnknownStatus Status = iota
	// ExpectingKeys is the status of the taker between taking an offer and
	// receiving a response with swap keys from the maker. It is also the
	// maker's status after creating an offer up until receiving keys from a
	// taker accepting the offer.
	ExpectingKeys
	// KeysExchanged is the status of the maker after a taker accepts his offer.
	KeysExchanged
	// ETHLocked is the taker status after locking her ETH up until confirming
	// that the maker locked his XMR.
	ETHLocked
	// XMRLocked is the maker's state after locking the XMR up until he confirms
	// that the the taker has set the contract to ready.
	XMRLocked
	// ContractReady is the taker's state after verifying the locked XMR and
	// setting the contract to ready.
	ContractReady
	// SweepingXMR is the taker's state after claiming the XMR and sweeping it
	// back into their primary wallet.
	// It can also be the maker's state if the maker refunds and is sweeping
	// the XMR back into his primary wallet.
	// Note: if sweeping is disabled, this stage does not occur.
	// Also note that the swap protocol is technically "done" at this stage;
	// however, this stage is required so that the node is aware that a sweep
	// is occurring in case of a daemon restart.
	SweepingXMR
	// CompletedSuccess represents a successful swap.
	CompletedSuccess
	// CompletedRefund represents a swap that was refunded.
	CompletedRefund
	// CompletedAbort represents the case where the swap aborts before any funds
	// are locked.
	CompletedAbort
)

const unknownString string = "unknown"

// NewStatus returns a Status from the given string.
// If there is no match, it returns UnknownStatus
func NewStatus(str string) Status {
	switch str {
	case "ExpectingKeys":
		return ExpectingKeys
	case "KeysExchanged":
		return KeysExchanged
	case "ETHLocked":
		return ETHLocked
	case "XMRLocked":
		return XMRLocked
	case "ContractReady":
		return ContractReady
	case "SweepingXMR":
		return SweepingXMR
	case "Success":
		return CompletedSuccess
	case "Refunded":
		return CompletedRefund
	case "Aborted":
		return CompletedAbort
	default:
		return UnknownStatus
	}
}

// String returns the status as a text string.
func (s Status) String() string {
	switch s {
	case ExpectingKeys:
		return "ExpectingKeys"
	case KeysExchanged:
		return "KeysExchanged"
	case ETHLocked:
		return "ETHLocked"
	case XMRLocked:
		return "XMRLocked"
	case ContractReady:
		return "ContractReady"
	case SweepingXMR:
		return "SweepingXMR"
	case CompletedSuccess:
		return "Success"
	case CompletedRefund:
		return "Refunded"
	case CompletedAbort:
		return "Aborted"
	default:
		return unknownString
	}
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (s *Status) UnmarshalText(data []byte) error {
	newStatus := NewStatus(string(data))
	if newStatus == UnknownStatus {
		return fmt.Errorf("unknown status %q", string(data))
	}
	*s = newStatus
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (s Status) MarshalText() ([]byte, error) {
	textStr := s.String()
	if textStr == unknownString {
		return nil, fmt.Errorf("unknown status %d", s)
	}
	return []byte(textStr), nil
}

// Description returns a description of the swap stage.
func (s Status) Description() string {
	switch s {
	case ExpectingKeys:
		return "keys have not yet been exchanged"
	case KeysExchanged:
		return "keys have been exchanged, but no value has been locked"
	case ETHLocked:
		return "the ETH provider has locked their ether, but no XMR has been locked"
	case XMRLocked:
		return "both the XMR and ETH providers have locked their funds"
	case ContractReady:
		return "the locked ether is ready to be claimed"
	case SweepingXMR:
		return "the XMR is being swept back into the primary wallet"
	case CompletedSuccess:
		return "the locked funds have been claimed and the swap has completed successfully"
	case CompletedRefund:
		return "the locked funds have been refunded and the swap has completed"
	case CompletedAbort:
		return "the swap was aborted before any funds were locked"
	default:
		return unknownString
	}
}

// IsOngoing returns true if the status means the swap has not completed
func (s Status) IsOngoing() bool {
	switch s {
	case ExpectingKeys, KeysExchanged, ETHLocked, XMRLocked, ContractReady, SweepingXMR:
		return true
	case UnknownStatus:
		panic("swap should not have UnknownStatus")
	default:
		return false
	}
}
