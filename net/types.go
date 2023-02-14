package net

import (
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net/message"

	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
)

type SwapState = common.SwapStateNet //nolint:revive

//nolint:revive
type (
	MessageType     = byte
	Message         = common.Message
	QueryResponse   = message.QueryResponse
	SendKeysMessage = message.SendKeysMessage
)

// Handler handles swap initiation messages.
// It is implemented by *xmrmaker.xmrmaker
type Handler interface {
	GetOffers() []*types.Offer
	HandleInitiateMessage(msg *SendKeysMessage) (s SwapState, resp Message, err error)
}

type swap struct {
	swapState SwapState
	stream    libp2pnetwork.Stream
}
