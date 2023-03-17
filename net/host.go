// Package net adds swap-specific functionality to go-p2p-net/Host,
// in particular the swap messages for querying and initiation.
package net

import (
	"context"
	"sync"
	"time"

	p2pnet "github.com/athanorlabs/go-p2p-net"
	logging "github.com/ipfs/go-log"
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
	// ProtocolID is the base atomic swap network protocol ID prefix. The full ID
	// includes the chain ID at the end.
	ProtocolID          = "/atomic-swap/0.2"
	maxMessageSize      = 1 << 17
	maxRelayMessageSize = 2048
)

var log = logging.Logger("net")

// P2pHost contains libp2p functionality used by the Host.
type P2pHost interface {
	Start() error
	Stop() error

	Advertise()
	Discover(provides string, searchTime time.Duration) ([]peer.ID, error)

	SetStreamHandler(string, func(libp2pnetwork.Stream))
	SetAdvertisedNamespacesFunc(fn func() []string)

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

	makerHandler MakerHandler
	takerHandler TakerHandler

	// swap instance info
	swapMu sync.Mutex
	swaps  map[types.Hash]*swap
}

// Config holds the initialization parameters for the NewHost constructor.
type Config struct {
	Ctx        context.Context
	DataDir    string
	Port       uint16
	KeyFile    string
	Bootnodes  []string
	ProtocolID string
	ListenIP   string
	IsRelayer  bool
}

// NewHost returns a new Host.
// The host implemented in this package is swap-specific; ie. it supports swap-specific
// messages (initiate and query).
func NewHost(cfg *Config) (*Host, error) {
	h, err := p2pnet.NewHost(&p2pnet.Config{
		Ctx:        cfg.Ctx,
		DataDir:    cfg.DataDir,
		Port:       cfg.Port,
		KeyFile:    cfg.KeyFile,
		Bootnodes:  cfg.Bootnodes,
		ProtocolID: cfg.ProtocolID,
		ListenIP:   cfg.ListenIP,
	})
	if err != nil {
		return nil, err
	}

	return &Host{
		ctx:       cfg.Ctx,
		h:         h,
		isRelayer: cfg.IsRelayer,
		swaps:     make(map[types.Hash]*swap),
	}, nil
}

// P2pHost returns the underlying go-p2p-net host.
func (h *Host) P2pHost() P2pHost {
	return h.h
}

func (h *Host) advertisedNamespaces() []string {
	provides := []string{""}

	if len(h.makerHandler.GetOffers()) > 0 {
		provides = append(provides, string(coins.ProvidesXMR))
	}

	if h.isRelayer {
		provides = append(provides, RelayerProvidesStr)
	}

	return provides
}

// SetHandlers sets the maker and taker instances used by the host, and configures
// the stream handlers.
func (h *Host) SetHandlers(makerHandler MakerHandler, takerHandler TakerHandler) {
	h.makerHandler = makerHandler
	h.takerHandler = takerHandler
	h.h.SetAdvertisedNamespacesFunc(h.advertisedNamespaces)

	h.h.SetStreamHandler(queryProtocolID, h.handleQueryStream)
	if h.isRelayer {
		h.h.SetStreamHandler(relayProtocolID, h.handleRelayStream)
	}
	h.h.SetStreamHandler(swapID, h.handleProtocolStream)
}

// Start starts the bootstrap and discovery process.
func (h *Host) Start() error {
	if h.makerHandler == nil || h.takerHandler == nil {
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
	h.swapMu.Lock()
	defer h.swapMu.Unlock()

	swap, has := h.swaps[id]
	if !has {
		return errNoOngoingSwap
	}

	return p2pnet.WriteStreamMessage(swap.stream, msg, swap.stream.Conn().RemotePeer())
}

// CloseProtocolStream closes the current swap protocol stream.
func (h *Host) CloseProtocolStream(id types.Hash) {
	swap, has := h.swaps[id]
	if !has {
		return
	}

	log.Debugf("closing stream: peer=%s protocol=%s",
		swap.stream.Conn().RemotePeer(), swap.stream.Protocol(),
	)
	_ = swap.stream.Close()
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
