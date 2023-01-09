// Package swapnet adds swap-specific functionality to net/Host,
// in particular the swap messages for querying and initiation.
package swapnet

import (
	"context"
	"sync"
	"time"

	logging "github.com/ipfs/go-log"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net"
)

const (
	protocolID     = "/atomic-swap/1.0"
	maxMessageSize = 1 << 17
)

var log = logging.Logger("host")

// NetHost contains libp2p functionality used by the Host.
type NetHost interface {
	Start() error
	Stop() error

	Advertise()
	Discover(provides string, searchTime time.Duration) ([]peer.ID, error)

	SetStreamHandler(string, func(libp2pnetwork.Stream))
	SetShouldAdvertiseFunc(net.ShouldAdvertiseFunc)

	Connectedness(peer.ID) libp2pnetwork.Connectedness
	Connect(context.Context, peer.AddrInfo) error
	NewStream(context.Context, peer.ID, protocol.ID) (libp2pnetwork.Stream, error)

	AddrInfo() peer.AddrInfo
	Addresses() []string
	PeerID() peer.ID
	ConnectedPeers() []string
}

// Host represents a p2p node that implements the atomic swap protocol.
type Host struct {
	ctx     context.Context
	h       NetHost
	handler Handler

	// swap instance info
	swapMu sync.Mutex
	swaps  map[types.Hash]*swap
}

// NewHost returns a new Host.
// The host implemented in this package is swap-specific; ie. it supports swap-specific
// messages (initiate and query).
func NewHost(cfg *net.Config) (*Host, error) {
	cfg.ProtocolID = protocolID

	h, err := net.NewHost(cfg)
	if err != nil {
		return nil, err
	}

	return &Host{
		ctx:   cfg.Ctx,
		h:     h,
		swaps: make(map[types.Hash]*swap),
	}, nil
}

// SetHandler sets the Handler instance used by the host.
func (h *Host) SetHandler(handler Handler) {
	fn := func() bool {
		return len(handler.GetOffers()) == 0
	}

	h.handler = handler
	h.h.SetShouldAdvertiseFunc(fn)
}

// Start starts the bootstrap and discovery process.
func (h *Host) Start() error {
	if h.handler == nil {
		return errNilHandler
	}

	h.h.SetStreamHandler(queryID, h.handleQueryStream)
	h.h.SetStreamHandler(swapID, h.handleProtocolStream)
	return h.h.Start()
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

	return net.WriteStreamMessage(swap.stream, msg, swap.stream.Conn().RemotePeer())
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

// Advertise advertises in the DHT.
func (h *Host) Advertise() {
	h.h.Advertise()
}

// Discover searches the DHT for peers that advertise that they provide the given coin..
// It searches for up to `searchTime` duration of time.
func (h *Host) Discover(provides coins.ProvidesCoin, searchTime time.Duration) ([]peer.ID, error) {
	return h.h.Discover(string(provides), searchTime)
}

// AddrInfo returns the host's AddrInfo.
func (h *Host) AddrInfo() peer.AddrInfo {
	return h.h.AddrInfo()
}

// Addresses returns the list of multiaddress the host is listening on.
func (h *Host) Addresses() []string {
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
