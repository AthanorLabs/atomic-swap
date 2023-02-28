package net

import (
	"context"
	"fmt"
	"time"

	p2pnet "github.com/athanorlabs/go-p2p-net"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/athanorlabs/atomic-swap/net/message"
)

const (
	queryID      = "/query/0"
	queryTimeout = time.Second * 5
)

func (h *Host) handleQueryStream(stream libp2pnetwork.Stream) {
	resp := &QueryResponse{
		Offers: h.handler.GetOffers(),
	}

	if err := p2pnet.WriteStreamMessage(stream, resp, stream.Conn().RemotePeer()); err != nil {
		log.Warnf("failed to send QueryResponse message to peer: err=%s", err)
	}

	_ = stream.Close()
}

// Query queries the given peer for its offers.
func (h *Host) Query(who peer.ID) (*QueryResponse, error) {
	ctx, cancel := context.WithTimeout(h.ctx, queryTimeout)
	defer cancel()

	if err := h.h.Connect(ctx, peer.AddrInfo{ID: who}); err != nil {
		return nil, err
	}

	stream, err := h.h.NewStream(ctx, who, protocol.ID(queryID))
	if err != nil {
		return nil, fmt.Errorf("failed to open stream with peer: err=%w", err)
	}

	log.Debugf("opened query stream: %s", stream.Conn())

	defer func() {
		_ = stream.Close()
	}()

	return receiveQueryResponse(stream)
}

func receiveQueryResponse(stream libp2pnetwork.Stream) (*QueryResponse, error) {
	msg, err := readStreamMessage(stream, maxMessageSize)
	if err != nil {
		return nil, fmt.Errorf("error reading QueryResponse: %w", err)
	}

	resp, ok := msg.(*QueryResponse)
	if !ok {
		return nil, fmt.Errorf("expected %s message but received %s",
			message.TypeToString(message.QueryResponseType),
			message.TypeToString(msg.Type()))
	}

	return resp, nil
}
