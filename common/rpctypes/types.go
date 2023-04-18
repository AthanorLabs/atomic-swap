// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

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
	OfferID types.Hash `json:"offerID" validate:"required"`
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
	PeerIDs []peer.ID `json:"peerIDs" validate:"dive,required"`
}

// QueryPeerRequest ...
type QueryPeerRequest struct {
	// Peer ID of peer to query
	PeerID peer.ID `json:"peerID" validate:"required"`
}

// QueryPeerResponse ...
type QueryPeerResponse struct {
	Offers []*types.Offer `json:"offers" validate:"dive,required"`
}

// PeerWithOffers ...
type PeerWithOffers struct {
	PeerID peer.ID        `json:"peerID" validate:"required"`
	Offers []*types.Offer `json:"offers" validate:"dive,required"`
}

// QueryAllRequest ...
type QueryAllRequest = DiscoverRequest

// QueryAllResponse ...
type QueryAllResponse struct {
	PeersWithOffers []*PeerWithOffers `json:"peersWithOffers" validate:"dive,required"`
}

// TakeOfferRequest ...
type TakeOfferRequest struct {
	PeerID         peer.ID      `json:"peerID" validate:"required"`
	OfferID        types.Hash   `json:"offerID" validate:"required"`
	ProvidesAmount *apd.Decimal `json:"providesAmount" validate:"required"` // eth asset amount
}

// MakeOfferRequest ...
type MakeOfferRequest struct {
	MinAmount    *apd.Decimal        `json:"minAmount" validate:"required"`
	MaxAmount    *apd.Decimal        `json:"maxAmount" validate:"required"`
	ExchangeRate *coins.ExchangeRate `json:"exchangeRate" validate:"required"`
	EthAsset     types.EthAsset      `json:"ethAsset,omitempty"`
	UseRelayer   bool                `json:"useRelayer,omitempty"`
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
	OfferID types.Hash        `json:"offerID" validate:"required"`
	To      ethcommon.Address `json:"to" validate:"required"`
	Data    []byte            `json:"data" validate:"required"`
	Value   *apd.Decimal      `json:"value" validate:"required"` // In ETH (or other ETH asset) not WEI
}

// SignerTxSigned is a response from the front-end saying the given tx has been submitted successfully
type SignerTxSigned struct {
	OfferID types.Hash     `json:"offerID" validate:"required"`
	TxHash  ethcommon.Hash `json:"txHash" validate:"required"`
}

// TokenInfoRequest is used to request lookup of the token's metadata.
type TokenInfoRequest struct {
	TokenAddr ethcommon.Address `json:"tokenAddr" validate:"required"`
}

// TokenInfoResponse contains the metadata for the requested token
type TokenInfoResponse = coins.ERC20TokenInfo

// BalancesRequest is used to request the combined Monero and Ethereum balances
// as well as the balances of any tokens included in the request.
type BalancesRequest struct {
	TokenAddrs []ethcommon.Address `json:"tokensAddrs" validate:"dive,required"`
}

// BalancesResponse holds the response for the combined Monero, Ethereum and
// optional token Balances request
type BalancesResponse struct {
	MoneroAddress           *mcrypto.Address          `json:"moneroAddress" validate:"required"`
	PiconeroBalance         *coins.PiconeroAmount     `json:"piconeroBalance" validate:"required"`
	PiconeroUnlockedBalance *coins.PiconeroAmount     `json:"piconeroUnlockedBalance" validate:"required"`
	BlocksToUnlock          uint64                    `json:"blocksToUnlock"`
	EthAddress              ethcommon.Address         `json:"ethAddress" validate:"required"`
	WeiBalance              *coins.WeiAmount          `json:"weiBalance" validate:"required"`
	TokenBalances           []*coins.ERC20TokenAmount `json:"tokenBalances" validate:"dive,required"`
}

// AddressesResponse ...
type AddressesResponse struct {
	Addrs []string `json:"addresses" validate:"dive,required"`
}

// PeersResponse ...
type PeersResponse struct {
	Addrs []string `json:"addresses" validate:"dive,required"`
}
