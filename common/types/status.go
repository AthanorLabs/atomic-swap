package types

// Status represents the stage that a swap is at.
type Status byte

const (
	ExpectingKeys Status = iota //nolint:revive
	KeysExchanged
	ContractDeployed
	XMRLocked
	ContractReady
	// CompletedSuccess represents a successful swap.
	CompletedSuccess
	// CompletedRefund represents a swap that was refunded.
	CompletedRefund
	// CompletedAbort represents the case where the swap aborts before any funds are locked.
	CompletedAbort
	UnknownStage
)

const unknownString string = "unknown"

// String ...
func (s Status) String() string {
	switch s {
	case ExpectingKeys:
		return "ExpectingKeys"
	case KeysExchanged:
		return "KeysExchanged"
	case ContractDeployed:
		return "ContractDeployed"
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
	case ContractDeployed:
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

func (s Status) IsOngoing() bool {
	switch s {
	case ExpectingKeys, KeysExchanged, ContractDeployed, XMRLocked, ContractReady, UnknownStage:
		return true
	default:
		return false
	}
}

// // ExitStatus represents the exit status of a swap.
// // It is "Ongoing" if the swap is still ongoing.
// type ExitStatus byte

// const (
// 	// Ongoing represents an ongoing swap.
// 	Ongoing ExitStatus = iota
// 	// Success represents a successful swap.
// 	Success
// 	// Refunded represents a swap that was refunded.
// 	Refunded
// 	// Aborted represents the case where the swap aborts before any funds are locked.
// 	Aborted
// )

// // String ...
// func (s ExitStatus) String() string {
// 	switch s {
// 	case Ongoing:
// 		return "ongoing"
// 	case Success:
// 		return "success"
// 	case Refunded:
// 		return "refunded"
// 	case Aborted:
// 		return "aborted"
// 	default:
// 		return "unknown"
// 	}
// }

// // StageOrExitStatus ...
// type StageOrExitStatus struct {
// 	Stage      *Stage
// 	ExitStatus *ExitStatus
// }

// // String ...
// func (s *StageOrExitStatus) String() string {
// 	if s.Stage != nil {
// 		return s.Stage.String()
// 	}

// 	if s.ExitStatus != nil {
// 		return s.ExitStatus.String()
// 	}

// 	return unknownString
// }

// // SetStage ...
// func (s *StageOrExitStatus) SetStage(stage Stage) {
// 	s.Stage = &stage
// 	s.ExitStatus = nil
// }

// // SetExitStatus ...
// func (s *StageOrExitStatus) SetExitStatus(es ExitStatus) {
// 	s.Stage = nil
// 	s.ExitStatus = &es
// }
