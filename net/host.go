package net

import (
	"context"
	"fmt"
	"path"
	"sync"
	"time"

	badger "github.com/ipfs/go-ds-badger2"
	"github.com/libp2p/go-libp2p"
	libp2phost "github.com/libp2p/go-libp2p/core/host"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoreds"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/chyeh/pubip"
	logging "github.com/ipfs/go-log"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
)

const (
	protocolID = "/atomic-swap"
	maxReads   = 128
)

var log = logging.Logger("net")
var _ Host = &host{}

// Host represents a peer-to-peer node (ie. a host)
type Host interface {
	Start() error
	Stop() error

	Advertise()
	Discover(provides types.ProvidesCoin, searchTime time.Duration) ([]peer.AddrInfo, error)
	Query(who peer.AddrInfo) (*QueryResponse, error)
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

	queryMu  sync.Mutex
	queryBuf []byte
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
	Handler     Handler
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

	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cfg.Port))
	if err != nil {
		return nil, err
	}

	var externalAddr ma.Multiaddr
	ip, err := pubip.Get()
	if err != nil {
		log.Warnf("failed to get public IP error: %v", err)
	} else {
		log.Debugf("got public IP address %s", ip)
		externalAddr, err = ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", ip, cfg.Port))
		if err != nil {
			return nil, err
		}
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
		libp2p.ListenAddrs(addr),
		libp2p.Identity(key),
		libp2p.NATPortMap(),
		libp2p.EnableAutoRelay(), // TODO: pass our bootnodes as static relays to this call?
		libp2p.EnableNATService(),
		libp2p.EnableHolePunching(),
		libp2p.Peerstore(ps),
		libp2p.AddrsFactory(func(as []ma.Multiaddr) []ma.Multiaddr {
			if cfg.Environment == common.Development {
				return as
			}

			// only advertize non-local addrs (if not in dev mode)
			addrs := []ma.Multiaddr{}
			for _, addr := range as {
				if !privateIPs.AddrBlocked(addr) {
					addrs = append(addrs, addr)
				}
			}

			if externalAddr == nil {
				return addrs
			}

			return append(addrs, externalAddr)
		}),
	}

	// format bootnodes
	bns, err := stringsToAddrInfos(cfg.Bootnodes)
	if err != nil {
		return nil, fmt.Errorf("failed to format bootnodes: %w", err)
	}

	// create libp2p host instance
	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	ourCtx, cancel := context.WithCancel(cfg.Ctx)
	hst := &host{
		ctx:        ourCtx,
		cancel:     cancel,
		protocolID: fmt.Sprintf("%s/%s/%d", protocolID, cfg.Environment, cfg.EthChainID),
		h:          h,
		handler:    cfg.Handler,
		ds:         ds,
		bootnodes:  bns,
		queryBuf:   make([]byte, 1024*5),
		swaps:      make(map[types.Hash]*swap),
	}

	hst.discovery, err = newDiscovery(ourCtx, h, hst.getBootnodes)
	if err != nil {
		return nil, err
	}

	return hst, nil
}

func (h *host) SetHandler(handler Handler) {
	h.handler = handler
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

	for _, addr := range h.multiaddrs() {
		log.Info("Started listening: address=", addr)
	}

	// ignore error - node should still be able to run without connecting to
	// bootstrap nodes (for now)
	_ = h.bootstrap()

	go h.logPeers()

	return h.discovery.start()
}

func (h *host) logPeers() {
	for {
		if h.ctx.Err() != nil {
			return
		}

		log.Debugf("peer count: %d", len(h.h.Network().Peers()))
		time.Sleep(time.Minute)
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

// Discover searches the DHT for peers that advertise that they provide the given coin.
// It searches for up to `searchTime` duration of time.
func (h *host) Discover(provides types.ProvidesCoin, searchTime time.Duration) ([]peer.AddrInfo, error) {
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

	return h.writeToStream(swap.stream, msg)
}

func (h *host) getBootnodes() []peer.AddrInfo {
	addrs := h.bootnodes
	for _, p := range h.h.Network().Peers() {
		addrs = append(addrs, h.h.Peerstore().PeerInfo(p))
	}
	return addrs
}

// multiaddrs returns the multiaddresses of the host
func (h *host) multiaddrs() (multiaddrs []ma.Multiaddr) {
	addrs := h.h.Addrs()
	for _, addr := range addrs {
		multiaddr, err := ma.NewMultiaddr(fmt.Sprintf("%s/p2p/%s", addr, h.h.ID()))
		if err != nil {
			continue
		}
		multiaddrs = append(multiaddrs, multiaddr)
	}
	return multiaddrs
}

func (h *host) addrInfo() peer.AddrInfo {
	return peer.AddrInfo{
		ID:    h.h.ID(),
		Addrs: h.h.Addrs(),
	}
}

func (h *host) writeToStream(s libp2pnetwork.Stream, msg Message) error {
	encMsg, err := msg.Encode()
	if err != nil {
		return err
	}

	msgLen := uint64(len(encMsg))
	lenBytes := uint64ToLEB128(msgLen)
	encMsg = append(lenBytes, encMsg...)

	_, err = s.Write(encMsg)
	if err != nil {
		return err
	}

	log.Debug(
		"Sent message to peer=", s.Conn().RemotePeer(), " type=", msg.Type(),
	)

	return nil
}

// readStream reads from the stream into the given buffer, returning the number of bytes read
func readStream(stream libp2pnetwork.Stream, buf []byte) (int, error) {
	if stream == nil {
		return 0, errNilStream
	}

	var (
		tot int
	)

	length, bytesRead, err := readLEB128ToUint64(stream, buf[:1])
	if err != nil {
		return bytesRead, fmt.Errorf("failed to read length: %w", err)
	}

	if length == 0 {
		return 0, nil
	}

	if length > uint64(len(buf)) {
		log.Warnf("received message with size greater than allocated message buffer: msg size=%d, buffer size=%d",
			length, len(buf))
		return 0, fmt.Errorf("message size greater than allocated message buffer: got %d", length)
	}

	tot = 0
	for i := 0; i < maxReads; i++ {
		n, err := stream.Read(buf[tot:])
		if err != nil {
			return n + tot, err
		}

		tot += n
		if tot == int(length) {
			break
		}
	}

	if tot != int(length) {
		return tot, fmt.Errorf("failed to read entire message: expected %d bytes, received %d bytes", length, tot)
	}

	return tot, nil
}

// bootstrap connects the host to the configured bootnodes
func (h *host) bootstrap() error {
	failed := 0
	for _, addrInfo := range h.bootnodes {
		log.Debugf("bootstrapping to peer: peer=%s", addrInfo.ID)
		err := h.h.Connect(h.ctx, addrInfo)
		if err != nil {
			log.Debugf("failed to bootstrap to peer: err=%s", err)
			failed++
		}
	}

	if failed == len(h.bootnodes) && len(h.bootnodes) != 0 {
		return errFailedToBootstrap
	}

	return nil
}
