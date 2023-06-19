// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package net adds swap-specific functionality to go-p2p-net/Host,
// in particular the swap messages for querying and initiation.
package net

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	p2pnet "github.com/athanorlabs/go-p2p-net"
	logging "github.com/ipfs/go-log/v2"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net/message"
)

const (
	baseProtocolID = "/atomic-swap"

	// p2pAPIVersion needs to be incremented every time:
	// * types that the p2p APIs exchange, such as offers, change in a breaking way
	// * changes to the API itself, like adding/removing methods
	// * the swapCreator contract changes
	p2pAPIVersion = 2

	maxMessageSize      = 1 << 17
	maxRelayMessageSize = 2048
	connectionTimeout   = time.Second * 5
)

var log = logging.Logger("net")

// P2pHost contains libp2p functionality used by the Host.
type P2pHost interface {
	Start() error
	Stop() error

	Advertise()
	Discover(provides string, searchTime time.Duration) ([]peer.ID, error)

	SetStreamHandler(string, func(libp2pnetwork.Stream))

	Connectedness(peer.ID) libp2pnetwork.Connectedness
	Connect(context.Context, peer.AddrInfo) error
	NewStream(context.Context, peer.ID, protocol.ID) (libp2pnetwork.Stream, error)

	AddrInfo() peer.AddrInfo
	Addresses() []ma.Multiaddr
	PeerID() peer.ID
	ConnectedPeers() []string
}

// Host represents a p2p node that implements the atomic swap protocol.
type Host struct {
	ctx       context.Context
	h         P2pHost
	isRelayer bool

	// set to true if the node is a bootnode-only node
	isBootnode bool

	makerHandler MakerHandler
	relayHandler RelayHandler

	// swap instance info
	swapMu sync.RWMutex
	swaps  map[types.Hash]*swap
}

// Config holds the initialization parameters for the NewHost constructor.
type Config struct {
	Ctx       context.Context
	Env       common.Environment
	DataDir   string
	Port      uint16
	KeyFile   string
	Bootnodes []string
	ListenIP  string
	IsRelayer bool
}

// ChainProtocolID returns the versioned p2p protocol ID that includes the
// Ethereum chain name being used. The streams that are opened between peers use
// this prefix. All provided values advertised in the DHT also use this prefix.
// Note that dedicated bootnodes don't have a chain name and don't open p2p
// streams, so they just use the word "bootnode" in place of a chain name.
func ChainProtocolID(env common.Environment) string {
	return fmt.Sprintf("%s/%s/%d", baseProtocolID, common.ChainNameFromEnv(env), p2pAPIVersion)
}

// NewHost returns a new Host.
// The host implemented in this package is swap-specific; ie. it supports swap-specific
// messages (initiate and query).
func NewHost(cfg *Config) (*Host, error) {
	isBootnode := cfg.Env == common.Bootnode
	if isBootnode && cfg.IsRelayer {
		return nil, errBootnodeCannotRelay
	}

	h := &Host{
		ctx:        cfg.Ctx,
		h:          nil, // set below
		isRelayer:  cfg.IsRelayer,
		isBootnode: isBootnode,
		swaps:      make(map[types.Hash]*swap),
	}

	baseProtocolID := ChainProtocolID(cfg.Env)
	log.Debugf("using base protocol %s", baseProtocolID)

	var err error
	h.h, err = p2pnet.NewHost(&p2pnet.Config{
		Ctx:                      cfg.Ctx,
		DataDir:                  cfg.DataDir,
		Port:                     cfg.Port,
		KeyFile:                  cfg.KeyFile,
		Bootnodes:                cfg.Bootnodes,
		ProtocolID:               baseProtocolID,
		ListenIP:                 cfg.ListenIP,
		AdvertisedNamespacesFunc: h.advertisedNamespaces,
	})
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *Host) advertisedNamespaces() []string {
	provides := []string{""}

	if !h.isBootnode && len(h.makerHandler.GetOffers()) > 0 {
		provides = append(provides, string(coins.ProvidesXMR))
	}

	if !h.isBootnode && h.isRelayer {
		provides = append(provides, RelayerProvidesStr)
	}

	return provides
}

// SetHandlers sets the maker and taker instances used by the host, and configures
// the stream handlers.
func (h *Host) SetHandlers(makerHandler MakerHandler, relayHandler RelayHandler) {
	h.makerHandler = makerHandler
	h.relayHandler = relayHandler

	h.h.SetStreamHandler(queryProtocolID, h.handleQueryStream)
	h.h.SetStreamHandler(relayProtocolID, h.handleRelayStream)
	h.h.SetStreamHandler(relayerQueryProtocolID, h.handleRelayerQueryStream)
	h.h.SetStreamHandler(swapID, h.handleProtocolStream)
}

// Start starts the bootstrap and discovery process.
func (h *Host) Start() error {
	if (h.makerHandler == nil || h.relayHandler == nil) && !h.isBootnode {
		return errNilHandler
	}

	// Note: Start() is non-blocking
	if err := h.h.Start(); err != nil {
		return err
	}

	return nil
}

// Stop stops the host.
func (h *Host) Stop() error {
	return h.h.Stop()
}

// SendSwapMessage sends a message to the peer who we're currently doing a swap with.
func (h *Host) SendSwapMessage(msg Message, id types.Hash) error {
	h.swapMu.RLock()
	defer h.swapMu.RUnlock()

	swap, has := h.swaps[id]
	if !has {
		return errNoOngoingSwap
	}

	return p2pnet.WriteStreamMessage(swap.stream, msg, swap.stream.Conn().RemotePeer())
}

// CloseProtocolStream closes the current swap protocol stream.
func (h *Host) CloseProtocolStream(offerID types.Hash) {
	h.swapMu.Lock()
	defer h.swapMu.Unlock()
	swap, has := h.swaps[offerID]
	if !has || swap.streamClosed {
		return
	}

	swap.streamClosed = true
	log.Debugf("closing stream: peer=%s protocol=%s",
		swap.stream.Conn().RemotePeer(), swap.stream.Protocol(),
	)

	_ = swap.stream.Close()
}

// DeleteOngoingSwap deletes an ongoing swap from the network's state.
// Note: the caller of this function must ensure that `CloseProtocolStream`
// has also been called.
func (h *Host) DeleteOngoingSwap(offerID types.Hash) {
	h.swapMu.Lock()
	defer h.swapMu.Unlock()

	swap, has := h.swaps[offerID]
	if !has {
		return
	}

	if !swap.streamClosed {
		log.Errorf("deleting ongoing swap where stream isn't closed: peer=%s protocol=%s",
			swap.stream.Conn().RemotePeer(), swap.stream.Protocol(),
		)
	}

	delete(h.swaps, offerID)
}

// Advertise advertises the namespaces now instead of waiting for the next periodic
// update. We use it when a new advertised namespace is added.
func (h *Host) Advertise() {
	h.h.Advertise()
}

// Discover searches the DHT for peers that advertise that they provide the given coin..
// It searches for up to `searchTime` duration of time.
func (h *Host) Discover(provides string, searchTime time.Duration) ([]peer.ID, error) {
	return h.h.Discover(provides, searchTime)
}

// AddrInfo returns the host's AddrInfo.
func (h *Host) AddrInfo() peer.AddrInfo {
	return h.h.AddrInfo()
}

// Addresses returns the list of multiaddress the host is listening on.
func (h *Host) Addresses() []ma.Multiaddr {
	return h.h.Addresses()
}

// ConnectedPeers returns the multiaddresses of our currently connected peers.
func (h *Host) ConnectedPeers() []string {
	return h.h.ConnectedPeers()
}

// PeerID returns the host's peer ID.
func (h *Host) PeerID() peer.ID {
	return h.h.AddrInfo().ID
}

func readStreamMessage(stream libp2pnetwork.Stream, maxMessageSize uint32) (common.Message, error) {
	msgBytes, err := p2pnet.ReadStreamMessage(stream, maxMessageSize)
	if err != nil {
		return nil, err
	}

	return message.DecodeMessage(msgBytes)
}

// nextStreamMessage returns a channel that will receive the next message from the stream.
// if there is an error reading from the stream, the channel will be closed, thus
// the received value will be nil.
func nextStreamMessage(stream libp2pnetwork.Stream, maxMessageSize uint32) <-chan common.Message {
	ch := make(chan common.Message)
	go func() {
		for {
			msg, err := readStreamMessage(stream, maxMessageSize)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					log.Warnf("failed to read stream message: %s", err)
				}
				close(ch)
				return
			}

			ch <- msg
		}
	}()

	return ch
}
