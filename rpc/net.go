// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package rpc

import (
	"context"
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

	ethcommon "github.com/ethereum/go-ethereum/common"
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
	ctx        context.Context
	net        Net
	xmrtaker   XMRTaker
	xmrmaker   XMRMaker
	pb         ProtocolBackend
	sm         swap.Manager
	isBootnode bool
}

// NewNetService ...
func NewNetService(
	ctx context.Context,
	net Net,
	xmrtaker XMRTaker,
	xmrmaker XMRMaker,
	pb ProtocolBackend,
	sm swap.Manager,
	isBootnode bool,
) *NetService {
	return &NetService{
		ctx:        ctx,
		net:        net,
		xmrtaker:   xmrtaker,
		xmrmaker:   xmrmaker,
		pb:         pb,
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

// Pairs returns all currently available pairs from offers of all peers
func (s *NetService) Pairs(_ *http.Request, req *rpctypes.PairsRequest, resp *rpctypes.PairsResponse) error {
	if s.isBootnode {
		return errUnsupportedForBootnode
	}

	peerIDs, err := s.discover(&rpctypes.DiscoverRequest{
		Provides:   "",
		SearchTime: req.SearchTime,
	})
	if err != nil {
		return err
	}

	pairs := make(map[ethcommon.Address]*types.Pair)

	for _, p := range peerIDs {
		msg, err := s.net.Query(p)
		if err != nil {
			log.Debugf("Failed to query peer ID %s", p)
			continue
		}

		if len(msg.Offers) == 0 {
			continue
		}

		for _, o := range msg.Offers {
			address := o.EthAsset.Address()
			pair, exists := pairs[address]

			if !exists {
				pair = types.NewPair(o.EthAsset)
				if pair.EthAsset.IsToken() {
					tokenInfo, tokenInfoErr := s.pb.ETHClient().ERC20Info(s.ctx, address)
					if tokenInfoErr != nil {
						log.Debugf("Error while reading token info: %s", tokenInfoErr)
						continue
					}
					pair.Token = *tokenInfo
				} else {
					pair.Token.Name = "Ether"
					pair.Token.Symbol = "ETH"
					pair.Token.NumDecimals = 18
					pair.Verified = true
				}
				pairs[address] = pair
			}

			err = pair.AddOffer(o)
			if err != nil {
				return err
			}
		}
	}

	pairsArray := make([]*types.Pair, 0, len(pairs))
	for _, pair := range pairs {
		if pair.EthAsset.IsETH() {
			pairsArray = append([]*types.Pair{pair}, pairsArray...)
		} else {
			pairsArray = append(pairsArray, pair)
		}
	}

	resp.Pairs = pairsArray
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
