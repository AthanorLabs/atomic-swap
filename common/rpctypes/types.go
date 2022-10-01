package rpctypes

import (
	"math/big"

	"github.com/athanorlabs/atomic-swap/common"
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
	Peers [][]string `json:"peers"`
}

// QueryPeerRequest ...
type QueryPeerRequest struct {
	// Multiaddr of peer to query
	Multiaddr string `json:"multiaddr"`
}

// QueryPeerResponse ...
type QueryPeerResponse struct {
	Offers []*types.Offer `json:"offers"`
}

// PeerWithOffers ...
type PeerWithOffers struct {
	Peer   []string       `json:"peer"`
	Offers []*types.Offer `json:"offers"`
}

// QueryAllResponse ...
type QueryAllResponse struct {
	PeersWithOffers []*PeerWithOffers `json:"peersWithOffers"`
}

// TakeOfferRequest ...
type TakeOfferRequest struct {
	Multiaddr      string  `json:"multiaddr"`
	OfferID        string  `json:"offerID"`
	ProvidesAmount float64 `json:"providesAmount"`
}

// TakeOfferResponse ...
type TakeOfferResponse struct {
	InfoFile string `json:"infoFile"`
}

// MakeOfferRequest ...
type MakeOfferRequest struct {
	MinimumAmount float64            `json:"minimumAmount"`
	MaximumAmount float64            `json:"maximumAmount"`
	ExchangeRate  types.ExchangeRate `json:"exchangeRate"`
	EthAsset      string             `json:"ethAsset,omitempty"`
}

// MakeOfferResponse ...
type MakeOfferResponse struct {
	ID       string `json:"offerID"`
	InfoFile string `json:"infoFile"`
}

// SignerRequest initiates the signer_subscribe handler from the front-end
type SignerRequest struct {
	OfferID    string `json:"offerID"`
	EthAddress string `json:"ethAddress"`
	XMRAddress string `json:"xmrAddress"`
}

// SignerResponse sends a tx to be signed to the front-end
type SignerResponse struct {
	OfferID string `json:"offerID"`
	To      string `json:"to"`
	Data    string `json:"data"`
	Value   string `json:"value"`
}

// SignerTxSigned is a response from the front-end saying the given tx has been submitted successfully
type SignerTxSigned struct {
	OfferID string `json:"offerID"`
	TxHash  string `json:"txHash"`
}

// BalancesResponse holds the response for the combined Monero and Ethereum Balances request
type BalancesResponse struct {
	MoneroAddress           string              `json:"monero_address"`
	PiconeroBalance         common.MoneroAmount `json:"piconero_balance"`
	PiconeroUnlockedBalance common.MoneroAmount `json:"piconero_unlocked_balance"`
	BlocksToUnlock          uint64              `json:"blocks_to_unlock"`
	EthAddress              string              `json:"eth_address"`
	EthBalance              *big.Int            `json:"eth_balance"`
}
