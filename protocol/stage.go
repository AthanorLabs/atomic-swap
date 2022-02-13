package protocol

import (
	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/net/message"
)

// GetStage returns the stage corresponding to a next expected message type.
func GetStage(t message.Type) common.Stage {
	switch t {
	case message.SendKeysType:
		return common.ExpectingKeysStage
	case message.NotifyContractDeployedType:
		return common.KeysExchangedStage
	case message.NotifyXMRLockType:
		return common.ContractDeployedStage
	case message.NotifyReadyType:
		return common.XMRLockedStage
	case message.NotifyClaimedType:
		return common.ContractReadyStage
	case message.NilType:
		return common.ClaimOrRefundStage
	default:
		return common.UnknownStage
	}
}
