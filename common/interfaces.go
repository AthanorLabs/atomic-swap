package common

import (
	"github.com/noot/atomic-swap/net/message"
)

// SwapState is the interface used by other packages in *alice.swapState or *bob.swapState.
type SwapState interface {
	SwapStateNet
	SwapStateRPC
}

// SwapStateNet handles incoming protocol messages for an initiated protocol.
// It is implemented by *alice.swapState and *bob.swapState
type SwapStateNet interface {
	HandleProtocolMessage(msg message.Message) (resp message.Message, done bool, err error)
	ProtocolExited() error
}

// SwapStateRPC contains the methods used by the RPC server into the SwapState.
type SwapStateRPC interface {
	SendKeysMessage() (*message.SendKeysMessage, error)
	ID() uint64
	Stage() Stage
}

// Stage represents the stage that a swap is at.
type Stage byte

const (
	ExpectingKeysStage Stage = iota //nolint:revive
	KeysExchangedStage
	ContractDeployedStage
	XMRLockedStage
	ContractReadyStage
	ClaimOrRefundStage
	UnknownStage
)

const unknownString string = "unknown"

// String ...
func (s Stage) String() string {
	switch s {
	case ExpectingKeysStage:
		return "ExpectingKeys"
	case KeysExchangedStage:
		return "KeysExchanged"
	case ContractDeployedStage:
		return "ContractDeployed"
	case XMRLockedStage:
		return "XMRLocked"
	case ContractReadyStage:
		return "ContractReady"
	case ClaimOrRefundStage:
		return "ClaimOrRefund"
	default:
		return unknownString
	}
}

// Info returns a description of the swap stage.
func (s Stage) Info() string {
	switch s {
	case ExpectingKeysStage:
		return "keys have not yet been exchanged"
	case KeysExchangedStage:
		return "keys have been exchanged, but no value has been locked"
	case ContractDeployedStage:
		return "the ETH provider has locked their ether, but no XMR has been locked"
	case XMRLockedStage:
		return "both the XMR and ETH providers have locked their funds"
	case ContractReadyStage:
		return "the locked ether is ready to be claimed"
	case ClaimOrRefundStage:
		return "the locked funds have been claimed or refunded"
	default:
		return unknownString
	}
}

// ExitStatus represents the exit status of a swap.
// It is "Ongoing" if the swap is still ongoing.
type ExitStatus byte

const (
	// Ongoing represents an ongoing swap.
	Ongoing ExitStatus = iota
	// Success represents a successful swap.
	Success
	// Refunded represents a swap that was refunded.
	Refunded
	// Aborted represents the case where the swap aborts before any funds are locked.
	Aborted
)

// String ...
func (s ExitStatus) String() string {
	switch s {
	case Ongoing:
		return "ongoing"
	case Success:
		return "success"
	case Refunded:
		return "refunded"
	case Aborted:
		return "aborted"
	default:
		return "unknown"
	}
}

type StageOrExitStatus struct {
	Stage      *Stage
	ExitStatus *ExitStatus
}

func (s *StageOrExitStatus) String() string {
	if s.Stage != nil {
		return s.Stage.String()
	}

	if s.ExitStatus != nil {
		return s.ExitStatus.String()
	}

	return unknownString
}

func (s *StageOrExitStatus) SetStage(stage Stage) {
	s.Stage = &stage
	s.ExitStatus = nil
}

func (s *StageOrExitStatus) SetExitStatus(es ExitStatus) {
	s.Stage = nil
	s.ExitStatus = &es
}
