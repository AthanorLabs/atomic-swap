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
	sm         SwapManager
	isBootnode bool
}

// NewNetService ...
func NewNetService(net Net, xmrtaker XMRTaker, xmrmaker XMRMaker, sm SwapManager, isBootnode bool) *NetService {
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
	peerIDs, err := s.discover(req)
	if err != nil {
		return err
	}

	resp.PeersWithOffers = make([]*rpctypes.PeerWithOffers, len(peerIDs))
	for i, p := range peerIDs {
		resp.PeersWithOffers[i] = &rpctypes.PeerWithOffers{
			PeerID: p,
		}
		msg, err := s.net.Query(p)
		if err != nil {
			log.Debugf("Failed to query peer ID %s", p)
			continue
		}
		resp.PeersWithOffers[i].Offers = msg.Offers
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
func (s *NetService) QueryPeer(_ *http.Request, req *rpctypes.QueryPeerRequest,
	resp *rpctypes.QueryPeerResponse) error {

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
	_, err := s.takeOffer(req.PeerID, req.OfferID, req.ProvidesAmount)
	if err != nil {
		return err
	}

	return nil
}

func (s *NetService) takeOffer(makerPeerID peer.ID, offerID types.Hash, providesAmount *apd.Decimal) (
	<-chan types.Status,
	error,
) {
	queryResp, err := s.net.Query(makerPeerID)
	if err != nil {
		return nil, err
	}

	var offer *types.Offer
	for _, maybeOffer := range queryResp.Offers {
		if offerID == maybeOffer.ID {
			offer = maybeOffer
			break
		}
	}
	if offer == nil {
		return nil, errNoOfferWithID
	}

	swapState, err := s.xmrtaker.InitiateProtocol(makerPeerID, providesAmount, offer)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate protocol: %w", err)
	}

	skm := swapState.SendKeysMessage().(*message.SendKeysMessage)
	skm.OfferID = offerID
	skm.ProvidedAmount = providesAmount

	if err = s.net.Initiate(peer.AddrInfo{ID: makerPeerID}, skm, swapState); err != nil {
		if err = swapState.Exit(); err != nil {
			log.Warnf("Swap exit failure: %s", err)
		}
		return nil, err
	}

	info, err := s.sm.GetOngoingSwap(offerID)
	if err != nil {
		return nil, err
	}

	return info.StatusCh(), nil
}

// TakeOfferSyncResponse ...
type TakeOfferSyncResponse struct {
	Status types.Status `json:"status" validate:"required"`
}

// TakeOfferSync initiates a swap with the given peer by taking an offer they've made.
// It synchronously waits until the swap is completed before returning its status.
func (s *NetService) TakeOfferSync(
	_ *http.Request,
	req *rpctypes.TakeOfferRequest,
	resp *TakeOfferSyncResponse,
) error {

	if _, err := s.takeOffer(req.PeerID, req.OfferID, req.ProvidesAmount); err != nil {
		return err
	}

	const checkSwapSleepDuration = time.Millisecond * 100

	for {
		time.Sleep(checkSwapSleepDuration)

		info, err := s.sm.GetPastSwap(req.OfferID)
		if err != nil {
			return err
		}

		if info == nil {
			continue
		}

		resp.Status = info.Status
		break
	}

	return nil
}

// MakeOffer creates and advertises a new swap offer.
func (s *NetService) MakeOffer(
	_ *http.Request,
	req *rpctypes.MakeOfferRequest,
	resp *rpctypes.MakeOfferResponse,
) error {
	offerResp, _, err := s.makeOffer(req)
	if err != nil {
		return err
	}
	*resp = *offerResp
	return nil
}

func (s *NetService) makeOffer(req *rpctypes.MakeOfferRequest) (*rpctypes.MakeOfferResponse, *types.OfferExtra, error) {
	offer := types.NewOffer(
		coins.ProvidesXMR,
		req.MinAmount,
		req.MaxAmount,
		req.ExchangeRate,
		req.EthAsset,
	)

	offerExtra, err := s.xmrmaker.MakeOffer(offer, req.UseRelayer)
	if err != nil {
		return nil, nil, err
	}

	return &rpctypes.MakeOfferResponse{
		PeerID:  s.net.PeerID(),
		OfferID: offer.ID,
	}, offerExtra, nil
}
