package protocol

import (
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net/message"
)

// GetStatus returns the status corresponding to a next expected message type.
func GetStatus(t message.Type) types.Status {
	switch t {
	case message.SendKeysType:
		return types.ExpectingKeys
	case message.NotifyETHLockedType:
		return types.KeysExchanged
	case message.NotifyXMRLockType:
		return types.ETHLocked
	// case message.NotifyReadyType:
	// 	return types.XMRLocked
	case message.NotifyClaimedType:
		return types.ContractReady
	default:
		return types.UnknownStatus
	}
}
