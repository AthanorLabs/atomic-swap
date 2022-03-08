package protocol

import (
	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/net/message"
)

// GetStatus returns the status corresponding to a next expected message type.
func GetStatus(t message.Type) types.Status {
	switch t {
	case message.SendKeysType:
		return types.ExpectingKeys
	case message.NotifyContractDeployedType:
		return types.KeysExchanged
	case message.NotifyXMRLockType:
		return types.ContractDeployed
	case message.NotifyReadyType:
		return types.XMRLocked
	case message.NotifyClaimedType:
		return types.ContractReady
	default:
		return types.UnknownStatus
	}
}
