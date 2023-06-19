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
	queryProtocolID = "/query"
)

func (h *Host) handleQueryStream(stream libp2pnetwork.Stream) {
	defer func() { _ = stream.Close() }()

	resp := &QueryResponse{
		Offers: h.makerHandler.GetOffers(),
	}

	if err := p2pnet.WriteStreamMessage(stream, resp, stream.Conn().RemotePeer()); err != nil {
		log.Warnf("failed to send QueryResponse message to peer: err=%s", err)
	}
}

// Query queries the given peer for its offers.
func (h *Host) Query(who peer.ID) (*QueryResponse, error) {
	ctx, cancel := context.WithTimeout(h.ctx, connectionTimeout)
	defer cancel()

	if err := h.h.Connect(ctx, peer.AddrInfo{ID: who}); err != nil {
		return nil, err
	}

	stream, err := h.h.NewStream(ctx, who, queryProtocolID)
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
	const queryResponseTimeout = time.Second * 15

	select {
	case msg := <-nextStreamMessage(stream, maxMessageSize):
		if msg == nil {
			return nil, errors.New("failed to read QueryResponse")
		}

		resp, ok := msg.(*QueryResponse)
		if !ok {
			return nil, fmt.Errorf("expected %s message but received %s",
				message.TypeToString(message.QueryResponseType),
				message.TypeToString(msg.Type()))
		}

		return resp, nil
	case <-time.After(queryResponseTimeout):
		return nil, errors.New("timed out waiting for QueryResponse")
	}
}
