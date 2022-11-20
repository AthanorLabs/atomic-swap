package common

import (
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net/message"
)

// SwapState is the interface used by other packages in *xmrtaker.swapState or *xmrmaker.swapState.
type SwapState interface {
	SwapStateNet
	SwapStateRPC
}

// SwapStateNet handles incoming protocol messages for an initiated protocol.
// It is implemented by *xmrtaker.swapState and *xmrmaker.swapState
type SwapStateNet interface {
	HandleProtocolMessage(msg message.Message) error
	ID() types.Hash
	Exit() error
}

// SwapStateRPC contains the methods used by the RPC server into the SwapState.
type SwapStateRPC interface {
	SendKeysMessage() *message.SendKeysMessage
	ID() types.Hash
	Exit() error
}
