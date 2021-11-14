package net

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	libp2pnetwork "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	queryID      = "/query/0"
	queryTimeout = time.Second * 5
)

func (h *host) handleQueryStream(stream libp2pnetwork.Stream) {
	if err := h.writeToStream(stream, h.queryResponse); err != nil {
		log.Warnf("failed to send QueryResponse message to peer: err=%s", err)
	}

	_ = stream.Close()
}

func (h *host) Query(who peer.AddrInfo) (*QueryResponse, error) {
	ctx, cancel := context.WithTimeout(h.ctx, queryTimeout)
	defer cancel()

	if err := h.h.Connect(ctx, who); err != nil {
		return nil, err
	}

	stream, err := h.h.NewStream(ctx, who.ID, protocolID+queryID)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream with peer: err=%w", err)
	}

	log.Debug(
		"opened query stream, peer=", who.ID,
	)

	defer func() {
		_ = stream.Close()
	}()

	return h.receiveQueryResponse(stream)
}

func (h *host) receiveQueryResponse(stream libp2pnetwork.Stream) (*QueryResponse, error) {
	h.queryMu.Lock()
	defer h.queryMu.Unlock()

	buf := h.queryBuf

	n, err := readStream(stream, buf)
	if err != nil {
		return nil, fmt.Errorf("read stream error: %w", err)
	}

	if n == 0 {
		return nil, fmt.Errorf("received empty message")
	}

	var resp *QueryResponse
	if err := json.Unmarshal(buf[1:n], &resp); err != nil {
		return nil, err
	}

	return resp, nil
}
