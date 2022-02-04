package net

import (
	"github.com/noot/atomic-swap/common/types"
)

// MessageSender is implemented by a Host
type MessageSender interface {
	SendSwapMessage(Message) error
}

// Handler handles swap initiation messages.
// It is implemented by *bob.bob
type Handler interface {
	GetOffers() []*types.Offer
	HandleInitiateMessage(msg *SendKeysMessage) (s SwapState, resp Message, err error)
}
