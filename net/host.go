// Package net provides libraries for direct communication between swapd nodes using libp2p.
package net

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"sync"
	"time"

	badger "github.com/ipfs/go-ds-badger2"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	libp2phost "github.com/libp2p/go-libp2p/core/host"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	libp2pdiscovery "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoreds"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net/message"
)

const (
	protocolID = "/atomic-swap"
)

var log = logging.Logger("net")
var _ Host = &host{}

// Host represents a peer-to-peer node (ie. a host)
type Host interface {
	Start() error
	Stop() error

	Advertise()
	Discover(provides types.ProvidesCoin, searchTime time.Duration) ([]peer.ID, error)
	Query(who peer.ID) (*QueryResponse, error)
	Initiate(who peer.AddrInfo, msg *SendKeysMessage, s common.SwapStateNet) error
	MessageSender
}

type swap struct {
	swapState SwapState
	stream    libp2pnetwork.Stream
}

type host struct {
	ctx        context.Context
	cancel     context.CancelFunc
	protocolID string

	h         libp2phost.Host
	bootnodes []peer.AddrInfo
	discovery *discovery
	handler   Handler
	ds        *badger.Datastore

	// swap instance info
	swapMu sync.Mutex
	swaps  map[types.Hash]*swap

	queryMu sync.Mutex
}

// Config is used to configure the network Host.
type Config struct {
	Ctx         context.Context
	Environment common.Environment
	DataDir     string
	EthChainID  int64
	Port        uint16
	KeyFile     string
	Bootnodes   []string
}

// QUIC will have better performance in high-bandwidth protocols if you increase a socket
// receive buffer (sysctl -w net.core.rmem_max=2500000). We have a low-bandwidth protocol,
// so setting this variable keeps a warning out of our logs. See this for more information:
// https://github.com/lucas-clemente/quic-go/wiki/UDP-Receive-Buffer-Size
func init() {
	_ = os.Setenv("QUIC_GO_DISABLE_RECEIVE_BUFFER_WARNING", "true")
}

// NewHost returns a new host
func NewHost(cfg *Config) (*host, error) {
	if cfg.DataDir == "" || cfg.KeyFile == "" {
		panic("required parameters not set")
	}

	key, err := loadKey(cfg.KeyFile)
	if err != nil {
		log.Debugf("failed to load libp2p key, generating key %s...", cfg.KeyFile)
		key, err = generateKey(0, cfg.KeyFile)
		if err != nil {
			return nil, err
		}
	}

	listenIP := "0.0.0.0"
	if cfg.Environment == common.Development {
		listenIP = "127.0.0.1"
	}

	ds, err := badger.NewDatastore(path.Join(cfg.DataDir, "libp2p-datastore"), &badger.DefaultOptions)
	if err != nil {
		return nil, err
	}

	ps, err := pstoreds.NewPeerstore(cfg.Ctx, ds, pstoreds.DefaultOpts())
	if err != nil {
		return nil, err
	}

	// set libp2p host options
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/%s/udp/%d/quic", listenIP, cfg.Port),
		), libp2p.Identity(key),
		libp2p.NATPortMap(),
		libp2p.EnableRelayService(),
		libp2p.EnableNATService(),
		libp2p.EnableHolePunching(),
		libp2p.Peerstore(ps),
	}

	// format bootnodes
	bns, err := stringsToAddrInfos(cfg.Bootnodes)
	if err != nil {
		return nil, fmt.Errorf("failed to format bootnodes: %w", err)
	}

	if len(bns) > 0 {
		opts = append(opts, libp2p.EnableAutoRelay(autorelay.WithStaticRelays(bns)))
	}

	// create libp2p host instance
	basicHost, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	// There is libp2p bug when calling `dual.New` with a cancelled context creating a panic,
	// so we need the extra guard below:
	// Panic:  https://github.com/jbenet/goprocess/blob/v0.1.4/impl-mutex.go#L99
	// Caller: https://github.com/libp2p/go-libp2p-kad-dht/blob/v0.17.0/dht.go#L222
	if cfg.Ctx.Err() != nil {
		return nil, err
	}

	dht, err := dual.New(cfg.Ctx, basicHost,
		dual.DHTOption(kaddht.BootstrapPeers(bns...)),
	)
	if err != nil {
		return nil, err
	}

	routedHost := routedhost.Wrap(basicHost, dht)

	ourCtx, cancel := context.WithCancel(cfg.Ctx)
	hst := &host{
		ctx:        ourCtx,
		cancel:     cancel,
		protocolID: fmt.Sprintf("%s/%s/%d", protocolID, cfg.Environment, cfg.EthChainID), // TODO: need's version
		h:          routedHost,
		ds:         ds,
		bootnodes:  bns,
		swaps:      make(map[types.Hash]*swap),
		discovery: &discovery{
			ctx:         ourCtx,
			dht:         dht,
			h:           routedHost,
			rd:          libp2pdiscovery.NewRoutingDiscovery(dht),
			provides:    nil,
			advertiseCh: make(chan struct{}),
			offerAPI:    nil,
		},
	}

	return hst, nil
}

func (h *host) SetHandler(handler Handler) {
	h.handler = handler
	h.discovery.setOfferAPI(handler)
}

func (h *host) Start() error {
	if h.handler == nil {
		return errNilHandler
	}

	h.h.SetStreamHandler(protocol.ID(h.protocolID+queryID), h.handleQueryStream)
	h.h.SetStreamHandler(protocol.ID(h.protocolID+swapID), h.handleProtocolStream)
	log.Debugf("supporting protocols %s and %s",
		protocol.ID(h.protocolID+queryID),
		protocol.ID(h.protocolID+swapID),
	)

	for _, addr := range h.h.Addrs() {
		log.Info("Started listening: address=", addr)
	}

	// ignore error - node should still be able to run without connecting to
	// bootstrap nodes (for now)
	_ = h.bootstrap() // TODO: Is this needed? Was it already done when initializing the DHT?

	go h.logPeers()

	return h.discovery.start()
}

func (h *host) logPeers() {
	logPeersInterval := time.Minute * 5

	for {
		log.Debugf("peer count: %d", len(h.h.Network().Peers()))
		err := common.SleepWithContext(h.ctx, logPeersInterval)
		if err != nil {
			// context was cancelled, return
			return
		}
	}
}

// Stop closes host services and the libp2p host (host services first)
func (h *host) Stop() error {
	h.cancel()

	if err := h.discovery.stop(); err != nil {
		return err
	}

	if err := h.h.Close(); err != nil {
		return fmt.Errorf("failed to close libp2p host: %w", err)
	}

	err := h.h.Peerstore().Close()
	if err != nil {
		return fmt.Errorf("failed to close peerstore: %w", err)
	}

	err = h.ds.Close()
	if err != nil {
		return fmt.Errorf("failed to close libp2p datastore: %w", err)
	}

	return nil
}

func (h *host) Advertise() {
	h.discovery.advertiseCh <- struct{}{}
}

func (h *host) Addresses() []string {
	var addrs []string
	for _, ma := range h.multiaddrs() {
		addrs = append(addrs, ma.String())
	}
	return addrs
}

func (h *host) PeerID() peer.ID {
	return h.h.ID()
}

func (h *host) ConnectedPeers() []string {
	var peers []string
	for _, c := range h.h.Network().Conns() {
		// the remote multi addr returned is just the transport
		p := fmt.Sprintf("%s/p2p/%s", c.RemoteMultiaddr(), c.RemotePeer())
		peers = append(peers, p)
	}
	return peers
}

// Discover searches the DHT for peers that advertise that they provide the given coin.
// It searches for up to `searchTime` duration of time.
func (h *host) Discover(provides types.ProvidesCoin, searchTime time.Duration) ([]peer.ID, error) {
	return h.discovery.discover(provides, searchTime)
}

// SendSwapMessage sends a message to the peer who we're currently doing a swap with.
func (h *host) SendSwapMessage(msg Message, id types.Hash) error {
	h.swapMu.Lock()
	defer h.swapMu.Unlock()

	swap, has := h.swaps[id]
	if !has {
		return errNoOngoingSwap
	}

	return writeStreamMessage(swap.stream, msg, swap.stream.Conn().RemotePeer())
}

// multiaddrs returns the local multiaddresses that we are listening on
func (h *host) multiaddrs() []ma.Multiaddr {
	addr := h.addrInfo()
	multiaddrs, err := peer.AddrInfoToP2pAddrs(&addr)
	if err != nil {
		// This shouldn't ever happen, but don't want to panic
		log.Errorf("Failed to convert AddrInfo=%q to Multiaddr: %s", addr, err)
	}
	return multiaddrs
}

func (h *host) addrInfo() peer.AddrInfo {
	return peer.AddrInfo{
		ID:    h.h.ID(),
		Addrs: h.h.Addrs(),
	}
}

func writeStreamMessage(s io.Writer, msg Message, peerID peer.ID) error {
	encMsg, err := msg.Encode()
	if err != nil {
		return err
	}

	err = binary.Write(s, binary.LittleEndian, uint32(len(encMsg)))
	if err != nil {
		return err
	}

	_, err = s.Write(encMsg)
	if err != nil {
		return err
	}

	log.Debugf("Sent message to peer=%s type=%s", peerID, msg.Type())

	return nil
}

// readStreamMessage reads the 4-byte LE size header and message body returning the
// message body bytes. io.EOF is returned if the stream is closed before any bytes
// are received. If a partial message is received before the stream closes,
// io.ErrUnexpectedEOF is returned.
func readStreamMessage(s io.Reader) (Message, error) {
	if s == nil {
		return nil, errNilStream
	}

	lenBuf := make([]byte, 4) // uint32 size
	n, err := io.ReadFull(s, lenBuf)
	if err != nil {
		if isEOF(err) {
			if n > 0 {
				err = io.ErrUnexpectedEOF
			} else {
				err = io.EOF
			}
		}
		return nil, err
	}
	msgLen := binary.LittleEndian.Uint32(lenBuf)

	if msgLen > maxMessageSize {
		log.Warnf("Received message longer than max allowed size: msg size=%d, max=%d",
			msgLen, maxMessageSize)
		return nil, fmt.Errorf("message size %d too large", msgLen)
	}

	msgBuf := make([]byte, msgLen)
	_, err = io.ReadFull(s, msgBuf)
	if err != nil {
		if isEOF(err) {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}

	return message.DecodeMessage(msgBuf)
}

func isEOF(err error) bool {
	switch {
	case
		errors.Is(err, net.ErrClosed), // what libp2p with QUIC usually generates
		errors.Is(err, io.EOF),
		errors.Is(err, io.ErrUnexpectedEOF),
		errors.Is(err, io.ErrClosedPipe):
		return true
	default:
		return false
	}
}

// bootstrap connects the host to the configured bootnodes
func (h *host) bootstrap() error {
	failed := 0
	for _, addrInfo := range h.bootnodes {
		log.Debugf("bootstrapping to peer: %s (%s)", addrInfo, h.h.Network().Connectedness(addrInfo.ID))
		err := h.h.Connect(h.ctx, addrInfo)
		if err != nil {
			log.Debugf("failed to bootstrap to peer: err=%s", err)
			failed++
		}
		log.Debugf("Bootstrapped connections to: %s", h.h.Network().ConnsToPeer(addrInfo.ID))
	}

	if failed == len(h.bootnodes) && len(h.bootnodes) != 0 {
		return errFailedToBootstrap
	}

	return nil
}
