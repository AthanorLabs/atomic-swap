// Package net provides libraries for direct communication between swapd nodes using libp2p.
package net

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"path"
	"sync"
	"time"

	"github.com/chyeh/pubip"
	badger "github.com/ipfs/go-ds-badger2"
	"github.com/libp2p/go-libp2p"
	libp2phost "github.com/libp2p/go-libp2p/core/host"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoreds"
	ma "github.com/multiformats/go-multiaddr"

	//"github.com/chyeh/pubip"
	logging "github.com/ipfs/go-log"

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

	queryMu sync.Mutex
}

// Config is used to configure the network Host.
type Config struct {
	Ctx           context.Context
	Environment   common.Environment
	DataDir       string
	EthChainID    int64
	Port          uint16
	KeyFile       string
	Bootnodes     []string
	StaticNATPort bool
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

	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", cfg.Port))
	if err != nil {
		return nil, err
	}

	ds, err := badger.NewDatastore(path.Join(cfg.DataDir, "libp2p-datastore"), &badger.DefaultOptions)
	if err != nil {
		return nil, err
	}

	ps, err := pstoreds.NewPeerstore(cfg.Ctx, ds, pstoreds.DefaultOpts())
	if err != nil {
		return nil, err
	}

	// format bootnodes
	bns, err := stringsToAddrInfos(cfg.Bootnodes)
	if err != nil {
		return nil, fmt.Errorf("failed to format bootnodes: %w", err)
	}

	// set libp2p host options
	opts := []libp2p.Option{
		libp2p.ListenAddrs(addr),
		libp2p.Identity(key),
		libp2p.NATPortMap(),
		libp2p.EnableRelayService(),
		libp2p.EnableNATService(),
		libp2p.EnableHolePunching(),
		libp2p.Peerstore(ps),
	}

	if len(bns) > 0 {
		opts = append(opts, libp2p.EnableAutoRelay(autorelay.WithStaticRelays(bns)))
	}
	if cfg.StaticNATPort {
		var externalAddr ma.Multiaddr
		ip, err := pubip.Get() //nolint:govet
		if err != nil {
			return nil, fmt.Errorf("failed to get public IP error: %s", err)
		}
		log.Infof("Public IP for static NAT port is %s", ip)
		externalAddr, err = ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/udp/%d/quic", ip, cfg.Port))
		if err != nil {
			return nil, err
		}
		opts = append(opts, libp2p.AddrsFactory(func(as []ma.Multiaddr) []ma.Multiaddr {
			var addrs []ma.Multiaddr
			for _, addr := range as {
				if !privateIPs.AddrBlocked(addr) {
					addrs = append(addrs, addr)
				}
			}
			return append(addrs, externalAddr)
		}))
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
		ds:         ds,
		bootnodes:  bns,
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

	return writeStreamMessage(swap.stream, msg, swap.stream.Conn().RemotePeer())
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
// message body bytes.
func readStreamMessage(s io.Reader) (Message, error) {
	if s == nil {
		return nil, errNilStream
	}

	var msgLen uint32
	err := binary.Read(s, binary.LittleEndian, &msgLen)
	if err != nil {
		return nil, fmt.Errorf("failed to read message length: %w", err)
	}

	if msgLen > maxMessageSize {
		log.Warnf("Received message longer than max allowed size: msg size=%d, max=%d",
			msgLen, maxMessageSize)
		return nil, fmt.Errorf("message size %d too large", msgLen)
	}

	buf := make([]byte, msgLen)
	_, err = io.ReadFull(s, buf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}

	return message.DecodeMessage(buf)
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
