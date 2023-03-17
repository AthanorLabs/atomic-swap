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
	MessageType        = byte
	Message            = common.Message
	QueryResponse      = message.QueryResponse
	SendKeysMessage    = message.SendKeysMessage
	RelayClaimRequest  = message.RelayClaimRequest
	RelayClaimResponse = message.RelayClaimResponse
)

// MakerHandler handles swap initiation messages and offer queries. It is
// implemented by *xmrmaker.Instance.
type MakerHandler interface {
	GetOffers() []*types.Offer
	HandleInitiateMessage(msg *SendKeysMessage) (SwapState, Message, error)
}

// TakerHandler handles relay claim requests. It is implemented by
// *xmrtaker.xmrtaker.
type TakerHandler interface {
	HandleRelayClaimRequest(msg *RelayClaimRequest) (*RelayClaimResponse, error)
}

type swap struct {
	swapState SwapState
	stream    libp2pnetwork.Stream
}
