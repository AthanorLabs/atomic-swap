// Package rpctypes provides the serialized types for queries and responses shared by
// swapd's JSON-RPC server and client-side libraries.
package rpctypes

import (
	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
)

// JSON RPC method names that we serve on the localhost server
const (
	NetDiscover         = "net_discover"
	NetQueryPeer        = "net_queryPeer"
	SubscribeNewPeer    = "net_subscribeNewPeer"
	SubscribeMakeOffer  = "net_makeOfferAndSubscribe"
	SubscribeTakeOffer  = "net_takeOfferAndSubscribe"
	SubscribeSwapStatus = "swap_subscribeStatus"
	SubscribeSigner     = "signer_subscribe"
)

// SubscribeSwapStatusRequest ...
type SubscribeSwapStatusRequest struct {
	OfferID types.Hash `json:"offerID"`
}

// SubscribeSwapStatusResponse ...
type SubscribeSwapStatusResponse struct {
	Status types.Status `json:"status" validate:"required"`
}

// DiscoverRequest ...
type DiscoverRequest struct {
	Provides   string `json:"provides"`
	SearchTime uint64 `json:"searchTime"` // in seconds
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
	PeerID peer.ID        `json:"peerID"`
	Offers []*types.Offer `json:"offers"`
}

// QueryAllRequest ...
type QueryAllRequest = DiscoverRequest

// QueryAllResponse ...
type QueryAllResponse struct {
	PeersWithOffers []*PeerWithOffers `json:"peersWithOffers"`
}

// TakeOfferRequest ...
type TakeOfferRequest struct {
	PeerID         peer.ID      `json:"peerID"`
	OfferID        types.Hash   `json:"offerID"`
	ProvidesAmount *apd.Decimal `json:"providesAmount"` // ether amount
}

// MakeOfferRequest ...
type MakeOfferRequest struct {
	MinAmount         *apd.Decimal        `json:"minAmount"`
	MaxAmount         *apd.Decimal        `json:"maxAmount"`
	ExchangeRate      *coins.ExchangeRate `json:"exchangeRate"`
	EthAsset          string              `json:"ethAsset,omitempty"`
	RelayerEndpoint   string              `json:"relayerEndpoint,omitempty"`
	RelayerCommission *apd.Decimal        `json:"relayerCommission,omitempty"`
}

// MakeOfferResponse ...
type MakeOfferResponse struct {
	PeerID  peer.ID    `json:"peerID" validate:"required"`
	OfferID types.Hash `json:"offerID" validate:"required"`
}

// SignerRequest initiates the signer_subscribe handler from the front-end
type SignerRequest struct {
	OfferID    types.Hash        `json:"offerID" validate:"required"`
	EthAddress ethcommon.Address `json:"ethAddress" validate:"required"`
	XMRAddress *mcrypto.Address  `json:"xmrAddress" validate:"required"`
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
	MoneroAddress           *mcrypto.Address      `json:"moneroAddress" validate:"required"`
	PiconeroBalance         *coins.PiconeroAmount `json:"piconeroBalance" validate:"required"`
	PiconeroUnlockedBalance *coins.PiconeroAmount `json:"piconeroUnlockedBalance" validate:"required"`
	BlocksToUnlock          uint64                `json:"blocksToUnlock"`
	EthAddress              ethcommon.Address     `json:"ethAddress" validate:"required"`
	WeiBalance              *coins.WeiAmount      `json:"weiBalance" validate:"required"`
}

// AddressesResponse ...
type AddressesResponse struct {
	Addrs []string `json:"addresses"`
}

// PeersResponse ...
type PeersResponse struct {
	Addrs []string `json:"addresses"`
}
