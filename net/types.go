package net

import (
	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/net/message"
)

type SwapState = common.SwapStateNet //nolint:revive

//nolint:revive
type (
	MessageType     = message.Type
	Message         = message.Message
	QueryResponse   = message.QueryResponse
	SendKeysMessage = message.SendKeysMessage
)

// MessageSender is implemented by a Host
type MessageSender interface {
	SendSwapMessage(Message) error
}

// Handler handles swap initiation messages.
// It is implemented by *xmrmaker.xmrmaker
type Handler interface {
	GetOffers() []*types.Offer
	HandleInitiateMessage(msg *SendKeysMessage) (s SwapState, resp Message, err error)
}
