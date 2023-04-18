// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package net

import (
	"context"
	"errors"
	"fmt"
	"time"

	p2pnet "github.com/athanorlabs/go-p2p-net"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/net/message"
)

const (
	relayProtocolID = "/relay/0"

	// RelayerProvidesStr is the DHT namespace advertised by nodes willing to relay
	// claims for arbitrary XMR makers.
	RelayerProvidesStr = "relayer"
)

// DiscoverRelayers returns the peer IDs of hosts that advertised their willingness to
// relay claim transactions.
func (h *Host) DiscoverRelayers() ([]peer.ID, error) {
	const defaultDiscoverTime = time.Second * 3
	return h.Discover(RelayerProvidesStr, defaultDiscoverTime)
}

func (h *Host) handleRelayStream(stream libp2pnetwork.Stream) {
	defer func() { _ = stream.Close() }()

	msg, err := readStreamMessage(stream, maxRelayMessageSize)
	if err != nil {
		log.Debugf("error reading RelayClaimRequest: %s", err)
		return
	}

	curPeer := stream.Conn().RemotePeer()

	req, ok := msg.(*RelayClaimRequest)
	if !ok {
		log.Debugf("ignoring wrong message type=%s sent to relay stream from %s",
			message.TypeToString(msg.Type()), curPeer)
		return
	}

	// Handle case where we are not a relayer, and the request didn't set the offerID
	// to indicate that it make from a running swap partner.

	// While HandleRelayClaimRequest(...) will do lower level validation on the
	// claim request, there are 2 validations best handled here:
	// (1) If the network layer is not advertising that we are a relayer to the
	//     DHT, we should not be getting claim requests targeted for open
	//     relayers (i.e. requests that do not have the OfferID set).
	// (2) If the request is purportedly from a maker to a taker of a current
	//     swap, then:
	//     (a) The swap should exist in our swaps map
	//     (b) The peerID who sent us the request much match the peerID with
	//         whom we are performing the swap.
	if req.OfferID == nil && !h.isRelayer {
		return
	} else if req.OfferID != nil {
		h.swapMu.RLock()
		swap, ok := h.swaps[*req.OfferID]
		h.swapMu.RUnlock()

		found := ok && swap.isTaker
		if !found || curPeer != swap.stream.Conn().RemotePeer() {
			log.Debugf("received invalid taker-specific claim request from peer=%s offerID=%s swap-found=%t",
				curPeer, req.OfferID, found)
			return
		}
	}

	resp, err := h.relayHandler.HandleRelayClaimRequest(req)
	if err != nil {
		log.Debugf("did not handle relay request: %s", err)
		return
	}

	log.Debugf("Relayed claim for %s with tx=%s", req.Swap.Claimer, resp.TxHash)

	if err := p2pnet.WriteStreamMessage(stream, resp, stream.Conn().RemotePeer()); err != nil {
		log.Warnf("failed to send RelayClaimResponse message to peer: %s", err)
		return
	}
}

// SubmitClaimToRelayer sends a request to relay a swap claim to a peer.
func (h *Host) SubmitClaimToRelayer(relayerID peer.ID, request *RelayClaimRequest) (*RelayClaimResponse, error) {
	ctx, cancel := context.WithTimeout(h.ctx, connectionTimeout)
	defer cancel()

	if err := h.h.Connect(ctx, peer.AddrInfo{ID: relayerID}); err != nil {
		return nil, err
	}

	stream, err := h.h.NewStream(ctx, relayerID, relayProtocolID)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream with peer: err=%w", err)
	}

	defer func() { _ = stream.Close() }()
	log.Debugf("opened relay stream: %s", stream.Conn())

	if err := p2pnet.WriteStreamMessage(stream, request, relayerID); err != nil {
		log.Warnf("failed to send RelayClaimRequest to peer: err=%s", err)
		return nil, err
	}

	return receiveRelayClaimResponse(stream)
}

func receiveRelayClaimResponse(stream libp2pnetwork.Stream) (*RelayClaimResponse, error) {
	// The timeout should be short enough, that the Maker can try multiple relayers
	// before T1 expires even if the receiving node accepts the relay request and
	// just sits on it without doing anything.
	const relayResponseTimeout = time.Second * 45

	select {
	case msg := <-nextStreamMessage(stream, maxMessageSize):
		if msg == nil {
			return nil, errors.New("failed to read RelayClaimResponse")
		}

		resp, ok := msg.(*RelayClaimResponse)
		if !ok {
			return nil, fmt.Errorf("expected %s message but received %s",
				message.TypeToString(message.RelayClaimResponseType),
				message.TypeToString(msg.Type()))
		}

		return resp, nil
	case <-time.After(relayResponseTimeout):
		return nil, errors.New("timed out waiting for QueryResponse")
	}
}
