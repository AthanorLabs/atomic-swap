package net

import (
	"context"
	"fmt"
	"time"

	p2pnet "github.com/athanorlabs/go-p2p-net"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/net/message"
)

const (
	relayProtocolID    = "/relay/0"
	relayClaimTimeout  = time.Second * 30 // TODO: Vet this value
	relayerProvidesStr = "relayer"
)

// DiscoverRelayers returns the peer IDs of hosts that advertised their willingness to
// relay claim transactions.
func (h *Host) DiscoverRelayers() ([]peer.ID, error) {
	const defaultDiscoverTime = time.Second * 3
	return h.Discover(relayerProvidesStr, defaultDiscoverTime)
}

func (h *Host) handleRelayStream(stream libp2pnetwork.Stream) {
	defer func() { _ = stream.Close() }()

	// TODO: If the request is from a Maker/OfferID combo that we did a swap with, we
	//       should always be willing to relay.
	if !h.isRelayer {
		return
	}

	msg, err := readStreamMessage(stream, maxRelayMessageSize)
	if err != nil {
		log.Debugf("error reading RelayClaimRequest: %s", err)
		return
	}

	req, ok := msg.(*RelayClaimRequest)
	if !ok {
		log.Debugf("ignoring wrong message type=%s sent to relay stream", message.TypeToString(msg.Type()))
		return
	}

	resp, err := h.takerHandler.HandleRelayClaimRequest(req)
	if err != nil {
		log.Debugf("Did not handle relay request: %s", err)
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
	ctx, cancel := context.WithTimeout(h.ctx, relayClaimTimeout)
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
	msg, err := readStreamMessage(stream, maxMessageSize)
	if err != nil {
		return nil, fmt.Errorf("error reading relay SubmitTransactionResponse: %w", err)
	}

	resp, ok := msg.(*RelayClaimResponse)
	if !ok {
		return nil, fmt.Errorf("expected %s message but received %s",
			message.TypeToString(message.RelayClaimResponseType),
			message.TypeToString(msg.Type()))
	}

	return resp, nil
}
