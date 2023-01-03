package swapnet

import (
	"context"
	"sync"
	"time"

	logging "github.com/ipfs/go-log"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net"
)

const (
	protocolID     = "/atomic-swap/0.1"
	maxMessageSize = 1 << 17
)

var log = logging.Logger("host")
var _ Host = &host{}

// Host represents a peer-to-peer node (ie. a host)
type Host interface {
	Start() error
	Stop() error

	Advertise()
	Discover(provides types.ProvidesCoin, searchTime time.Duration) ([]peer.ID, error)

	AddrInfo() peer.AddrInfo
	Addresses() []string
	ConnectedPeers() []string
	PeerID() peer.ID

	Query(who peer.ID) (*QueryResponse, error)
	Initiate(who peer.AddrInfo, msg *SendKeysMessage, s common.SwapStateNet) error
	MessageSender
}

type swap struct {
	swapState SwapState
	stream    libp2pnetwork.Stream
}

type host struct {
	ctx     context.Context
	h       net.Host
	handler Handler

	// swap instance info
	swapMu sync.Mutex
	swaps  map[types.Hash]*swap
}

// NewHost returns a new Host.
// The host implemented in this package is swap-specific; ie. it supports swap-specific
// messages (initiate and query).
func NewHost(cfg *net.Config) (*host, error) {
	cfg.ProtocolID = protocolID

	h, err := net.NewHost(cfg)
	if err != nil {
		return nil, err
	}

	return &host{
		ctx:   cfg.Ctx,
		h:     h,
		swaps: make(map[types.Hash]*swap),
	}, nil
}

func (h *host) SetHandler(handler Handler) {
	fn := func() bool {
		return len(handler.GetOffers()) == 0
	}

	h.handler = handler
	h.h.SetShouldAdvertiseFunc(fn)
}

func (h *host) Start() error {
	if h.handler == nil {
		return errNilHandler
	}

	h.h.SetStreamHandler(queryID, h.handleQueryStream)
	h.h.SetStreamHandler(swapID, h.handleProtocolStream)
	return h.h.Start()
}

func (h *host) Stop() error {
	return h.h.Stop()
}

// SendSwapMessage sends a message to the peer who we're currently doing a swap with.
func (h *host) SendSwapMessage(msg Message, id types.Hash) error {
	h.swapMu.Lock()
	defer h.swapMu.Unlock()

	swap, has := h.swaps[id]
	if !has {
		return errNoOngoingSwap
	}

	return net.WriteStreamMessage(swap.stream, msg, swap.stream.Conn().RemotePeer())
}

func (h *host) Advertise() {
	h.h.Advertise()
}

func (h *host) Discover(provides types.ProvidesCoin, searchTime time.Duration) ([]peer.ID, error) {
	return h.h.Discover(string(provides), searchTime)
}

func (h *host) AddrInfo() peer.AddrInfo {
	return h.h.AddrInfo()
}

func (h *host) Addresses() []string {
	return h.h.Addresses()
}

func (h *host) ConnectedPeers() []string {
	return h.h.ConnectedPeers()
}

func (h *host) PeerID() peer.ID {
	return h.h.AddrInfo().ID
}
