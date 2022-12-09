// Package rpctypes provides the serialized types for queries and responses shared by
// swapd's JSON-RPC server and client-side libraries.
package rpctypes

import (
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/common/types"
)

const (
	NetDiscover         = "net_discover" //nolint:revive
	NetQueryPeer        = "net_queryPeer"
	SubscribeNewPeer    = "net_subscribeNewPeer"
	SubscribeMakeOffer  = "net_makeOfferAndSubscribe"
	SubscribeTakeOffer  = "net_takeOfferAndSubscribe"
	SubscribeSwapStatus = "swap_subscribeStatus"
	SubscribeSigner     = "signer_subscribe"
)

// SubscribeSwapStatusRequest ...
type SubscribeSwapStatusRequest struct {
	ID types.Hash `json:"id"`
}

// SubscribeSwapStatusResponse ...
type SubscribeSwapStatusResponse struct {
	Status string `json:"status"`
}

// DiscoverRequest ...
type DiscoverRequest struct {
	Provides   types.ProvidesCoin `json:"provides"`
	SearchTime uint64             `json:"searchTime"` // in seconds
}

// DiscoverResponse ...
type DiscoverResponse struct {
	PeerIDs []peer.ID `json:"peerIDs"`
}

// QueryPeerRequest ...
type QueryPeerRequest struct {
	// Peer ID of peer to query
	PeerID peer.ID `json:"peerID"`
}

// QueryPeerResponse ...
type QueryPeerResponse struct {
	Offers []*types.Offer `json:"offers"`
}

// PeerWithOffers ...
type PeerWithOffers struct {
	PeerID peer.ID        `json:"peer"`
	Offers []*types.Offer `json:"offers"`
}

// QueryAllResponse ...
type QueryAllResponse struct {
	PeersWithOffers []*PeerWithOffers `json:"peersWithOffers"`
}

// TakeOfferRequest ...
type TakeOfferRequest struct {
	PeerID         peer.ID    `json:"peerID"`
	OfferID        types.Hash `json:"offerID"`
	ProvidesAmount float64    `json:"providesAmount"`
}

// MakeOfferRequest ...
type MakeOfferRequest struct {
	MinimumAmount     float64            `json:"minimumAmount"`
	MaximumAmount     float64            `json:"maximumAmount"`
	ExchangeRate      types.ExchangeRate `json:"exchangeRate"`
	EthAsset          string             `json:"ethAsset,omitempty"`
	RelayerEndpoint   string             `json:"relayerEndpoint,omitempty"`
	RelayerCommission float64            `json:"relayerCommission,omitempty"`
}

// MakeOfferResponse ...
type MakeOfferResponse struct {
	PeerID  peer.ID    `json:"peerID"`
	OfferID types.Hash `json:"offerID"`
}

// SignerRequest initiates the signer_subscribe handler from the front-end
type SignerRequest struct {
	OfferID    types.Hash `json:"offerID"`
	EthAddress string     `json:"ethAddress"`
	XMRAddress string     `json:"xmrAddress"`
}

// SignerResponse sends a tx to be signed to the front-end
type SignerResponse struct {
	OfferID types.Hash `json:"offerID"`
	To      string     `json:"to"`
	Data    string     `json:"data"`
	Value   string     `json:"value"`
}

// SignerTxSigned is a response from the front-end saying the given tx has been submitted successfully
type SignerTxSigned struct {
	OfferID types.Hash     `json:"offerID"`
	TxHash  ethcommon.Hash `json:"txHash"`
}

// BalancesResponse holds the response for the combined Monero and Ethereum Balances request
type BalancesResponse struct {
	MoneroAddress           string   `json:"moneroAddress"`
	PiconeroBalance         uint64   `json:"piconeroBalance"`
	PiconeroUnlockedBalance uint64   `json:"piconeroUnlockedBalance"`
	BlocksToUnlock          uint64   `json:"blocksToUnlock"`
	EthAddress              string   `json:"ethAddress"`
	WeiBalance              *big.Int `json:"weiBalance"`
}
