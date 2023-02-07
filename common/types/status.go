// Package types is for types that are shared by multiple packages
package types

// Status represents the stage that a swap is at.
type Status byte

// Status values
const (
	// ExpectingKeys is the status of the XMR Taker between taking an offer and receiving
	// a response with swap keys from the XMR Maker. It is also the Maker's status after creating
	// an offer up until receiving keys from a Taker accepting the offer.
	ExpectingKeys Status = iota
	// KeysExchanged is the status of the XMR Maker after an XMR Taker accepts his offer.
	KeysExchanged
	// ETHLocked is the Taker status after locking her ETH up until confirming that the Maker
	// locked his XMR.
	ETHLocked
	// XMRLocked is the XMR Maker's state after locking the XMR up until he confirms that the
	// the Taker has set the contract to ready.
	XMRLocked
	// ContractReady is the XMR Taker state after verifying the locked XMR and setting the contract
	// to ready.
	ContractReady
	// CompletedSuccess represents a successful swap.
	CompletedSuccess
	// CompletedRefund represents a swap that was refunded.
	CompletedRefund
	// CompletedAbort represents the case where the swap aborts before any funds are locked.
	CompletedAbort
	// UnknownStatus is a placeholder for unmatched status strings.
	// TODO: Why is it an indicator that a swap is ongoing?
	UnknownStatus
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

// String ...
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

// Info returns a description of the swap stage.
func (s Status) Info() string {
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
	case ExpectingKeys, KeysExchanged, ETHLocked, XMRLocked, ContractReady, UnknownStatus:
		return true
	default:
		return false
	}
}
