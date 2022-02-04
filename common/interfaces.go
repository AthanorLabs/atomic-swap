package common

import (
	"github.com/noot/atomic-swap/net/message"
)

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

type SwapStateRPC interface {
	SendKeysMessage() (*message.SendKeysMessage, error)
	ID() uint64
}
