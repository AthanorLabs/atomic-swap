package net

import (
	"github.com/libp2p/go-libp2p/core/peer"

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
	HandleInitiateMessage(peerID peer.ID, msg *SendKeysMessage) (SwapState, Message, error)
}

// RelayHandler handles relay claim requests. It is implemented by
// *backend.backend.
type RelayHandler interface {
	HandleRelayClaimRequest(msg *RelayClaimRequest) (*RelayClaimResponse, error)
}

type swap struct {
	swapState SwapState
	stream    libp2pnetwork.Stream
	// isTaker is true if we initiate the swap (created the outbound stream). If
	// the value is false, it is not safe to assume that we are the maker, as
	// the swap entry will get created before validation of the first received
	// message.
	isTaker bool
}
