package rpctypes

import (
	"github.com/noot/atomic-swap/common/types"
)

// SubscribeSwapStatusRequest ...
type SubscribeSwapStatusRequest struct {
	ID uint64 `json:"id"`
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

// TakeOfferRequest ...
type TakeOfferRequest struct {
	Multiaddr      string  `json:"multiaddr"`
	OfferID        string  `json:"offerID"`
	ProvidesAmount float64 `json:"providesAmount"`
}

// TakeOfferResponse ...
type TakeOfferResponse struct {
	ID       uint64 `json:"id"`
	InfoFile string `json:"infoFile"`
}

// MakeOfferRequest ...
type MakeOfferRequest struct {
	MinimumAmount float64            `json:"minimumAmount"`
	MaximumAmount float64            `json:"maximumAmount"`
	ExchangeRate  types.ExchangeRate `json:"exchangeRate"`
}

// MakeOfferResponse ...
type MakeOfferResponse struct {
	ID       string `json:"offerID"`
	InfoFile string `json:"infoFile"`
}

type SignerRequest struct {
	OfferID    string `json:"offerID"`
	EthAddress string `json:"ethAddress"`
	XMRAddress string `json:"xmrAddress"`
}

type SignerResponse struct {
	OfferID string `json:"offerID"`
	To      string `json:"to"`
	Data    string `json:"data"`
	Value   string `json:"value"`
}

type SignerTxSigned struct {
	OfferID string `json:"offerID"`
	TxHash  string `json:"txHash"`
}
