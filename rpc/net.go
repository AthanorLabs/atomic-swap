// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package rpc

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net/message"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
)

const defaultSearchTime = time.Second * 12

// Net contains the network-related functions required by the rpc service.
type Net interface {
	PeerID() peer.ID
	ConnectedPeers() []string
	Addresses() []ma.Multiaddr
	Discover(provides string, searchTime time.Duration) ([]peer.ID, error)
	Query(who peer.ID) (*message.QueryResponse, error)
	Initiate(who peer.AddrInfo, sendKeysMessage common.Message, s common.SwapStateNet) error
	CloseProtocolStream(types.Hash)
}

// NetService is the RPC service prefixed by net_.
type NetService struct {
	net        Net
	xmrtaker   XMRTaker
	xmrmaker   XMRMaker
	sm         swap.Manager
	isBootnode bool
}

// NewNetService ...
func NewNetService(net Net, xmrtaker XMRTaker, xmrmaker XMRMaker, sm swap.Manager, isBootnode bool) *NetService {
	return &NetService{
		net:        net,
		xmrtaker:   xmrtaker,
		xmrmaker:   xmrmaker,
		sm:         sm,
		isBootnode: isBootnode,
	}
}

// Addresses returns the local listening multi-addresses. Note that local listening
// addresses do not correspond to what remote peers connect to unless your host has a
// public IP directly attached to a local interface.
func (s *NetService) Addresses(_ *http.Request, _ *interface{}, resp *rpctypes.AddressesResponse) error {
	// Multiaddr is an interface that you can serialize, but you need a concrete
	// type to deserialize, so we just use strings in the AddressesResponse.
	addresses := s.net.Addresses()
	resp.Addrs = make([]string, 0, len(addresses))
	for _, a := range addresses {
		resp.Addrs = append(resp.Addrs, a.String())
	}
	return nil
}

// Peers returns the peers that this node is currently connected to.
func (s *NetService) Peers(_ *http.Request, _ *interface{}, resp *rpctypes.PeersResponse) error {
	resp.Addrs = s.net.ConnectedPeers()
	return nil
}

// QueryAll discovers peers who provide a certain coin and queries all of them for their current offers.
func (s *NetService) QueryAll(_ *http.Request, req *rpctypes.QueryAllRequest, resp *rpctypes.QueryAllResponse) error {
	if s.isBootnode {
		return errUnsupportedForBootnode
	}

	peerIDs, err := s.discover(req)
	if err != nil {
		return err
	}

	resp.PeersWithOffers = make([]*rpctypes.PeerWithOffers, 0, len(peerIDs))
	for _, p := range peerIDs {
		msg, err := s.net.Query(p)
		if err != nil {
			log.Debugf("Failed to query peer ID %s", p)
			continue
		}
		if len(msg.Offers) > 0 {
			resp.PeersWithOffers = append(resp.PeersWithOffers, &rpctypes.PeerWithOffers{
				PeerID: p,
				Offers: msg.Offers,
			})
		}
	}

	return nil
}

func (s *NetService) discover(req *rpctypes.DiscoverRequest) ([]peer.ID, error) {
	searchTime, err := time.ParseDuration(fmt.Sprintf("%ds", req.SearchTime))
	if err != nil {
		return nil, err
	}

	if searchTime == 0 {
		searchTime = defaultSearchTime
	}

	return s.net.Discover(req.Provides, searchTime)
}

// Discover discovers peers over the network that provide a certain coin up for `SearchTime` duration of time.
func (s *NetService) Discover(_ *http.Request, req *rpctypes.DiscoverRequest, resp *rpctypes.DiscoverResponse) error {
	searchTime, err := time.ParseDuration(fmt.Sprintf("%ds", req.SearchTime))
	if err != nil {
		return err
	}

	if searchTime == 0 {
		searchTime = defaultSearchTime
	}

	resp.PeerIDs, err = s.net.Discover(req.Provides, searchTime)
	if err != nil {
		return err
	}

	return nil
}

// QueryPeer queries a peer for the coins they provide, their maximum amounts, and desired exchange rate.
func (s *NetService) QueryPeer(
	_ *http.Request,
	req *rpctypes.QueryPeerRequest,
	resp *rpctypes.QueryPeerResponse,
) error {
	if s.isBootnode {
		return errUnsupportedForBootnode
	}

	msg, err := s.net.Query(req.PeerID)
	if err != nil {
		return err
	}

	resp.Offers = msg.Offers
	return nil
}

// TakeOffer initiates a swap with the given peer by taking an offer they've made.
func (s *NetService) TakeOffer(
	_ *http.Request,
	req *rpctypes.TakeOfferRequest,
	_ *interface{},
) error {
	if s.isBootnode {
		return errUnsupportedForBootnode
	}

	err := s.takeOffer(req.PeerID, req.OfferID, req.ProvidesAmount)
	if err != nil {
		return err
	}

	return nil
}

func (s *NetService) takeOffer(makerPeerID peer.ID, offerID types.Hash, providesAmount *apd.Decimal) error {
	queryResp, err := s.net.Query(makerPeerID)
	if err != nil {
		return err
	}

	var offer *types.Offer
	for _, maybeOffer := range queryResp.Offers {
		if offerID == maybeOffer.ID {
			offer = maybeOffer
			break
		}
	}
	if offer == nil {
		return errNoOfferWithID
	}

	swapState, err := s.xmrtaker.InitiateProtocol(makerPeerID, providesAmount, offer)
	if err != nil {
		return err
	}

	skm := swapState.SendKeysMessage().(*message.SendKeysMessage)
	skm.OfferID = offerID
	skm.ProvidedAmount = providesAmount

	if err = s.net.Initiate(peer.AddrInfo{ID: makerPeerID}, skm, swapState); err != nil {
		if err = swapState.Exit(); err != nil {
			log.Warnf("Swap exit failure: %s", err)
		}
		return err
	}

	return nil
}

// MakeOffer creates and advertises a new swap offer.
func (s *NetService) MakeOffer(
	_ *http.Request,
	req *rpctypes.MakeOfferRequest,
	resp *rpctypes.MakeOfferResponse,
) error {
	if s.isBootnode {
		return errUnsupportedForBootnode
	}

	offerResp, err := s.makeOffer(req)
	if err != nil {
		return err
	}
	*resp = *offerResp
	return nil
}

func (s *NetService) makeOffer(req *rpctypes.MakeOfferRequest) (*rpctypes.MakeOfferResponse, error) {
	offer := types.NewOffer(
		coins.ProvidesXMR,
		req.MinAmount,
		req.MaxAmount,
		req.ExchangeRate,
		req.EthAsset,
	)

	_, err := s.xmrmaker.MakeOffer(offer, req.UseRelayer)
	if err != nil {
		return nil, err
	}

	return &rpctypes.MakeOfferResponse{
		PeerID:  s.net.PeerID(),
		OfferID: offer.ID,
	}, nil
}
