package net

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/noot/atomic-swap/common"

	"github.com/libp2p/go-libp2p"
	libp2phost "github.com/libp2p/go-libp2p-core/host"
	libp2pnetwork "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	ma "github.com/multiformats/go-multiaddr"

	logging "github.com/ipfs/go-log"
)

const (
	protocolID     = "/atomic-swap"
	maxReads       = 128
	defaultKeyFile = "net.key"
)

var log = logging.Logger("net")
var _ Host = &host{}

// Host represents a peer-to-peer node (ie. a host)
type Host interface {
	Start() error
	Stop() error

	Discover(provides common.ProvidesCoin, searchTime time.Duration) ([]peer.AddrInfo, error)
	Query(who peer.AddrInfo) (*QueryResponse, error)
	Initiate(who peer.AddrInfo, msg *SendKeysMessage, s SwapState) error
	MessageSender
}

// MessageSender is implemented by a Host
type MessageSender interface {
	SendSwapMessage(Message) error
}

type host struct {
	ctx        context.Context
	cancel     context.CancelFunc
	protocolID string

	h         libp2phost.Host
	bootnodes []peer.AddrInfo
	discovery *discovery
	handler   Handler

	// swap instance info
	swapMu     sync.Mutex
	swapState  SwapState
	swapStream libp2pnetwork.Stream

	queryMu  sync.Mutex
	queryBuf []byte
}

// Config is used to configure the network Host.
type Config struct {
	Ctx         context.Context
	Environment common.Environment
	ChainID     int64
	Port        uint16
	KeyFile     string
	Bootnodes   []string
	Handler     Handler
}

// NewHost returns a new host
func NewHost(cfg *Config) (*host, error) { //nolint:revive
	if cfg.KeyFile == "" {
		cfg.KeyFile = defaultKeyFile
	}

	key, err := loadKey(cfg.KeyFile)
	if err != nil {
		fmt.Println("failed to load libp2p key, generating key...", cfg.KeyFile)
		key, err = generateKey(0, cfg.KeyFile)
		if err != nil {
			return nil, err
		}
	}

	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cfg.Port))
	if err != nil {
		return nil, err
	}

	// set libp2p host options
	opts := []libp2p.Option{
		libp2p.ListenAddrs(addr),
		libp2p.DisableRelay(),
		libp2p.Identity(key),
		libp2p.NATPortMap(),
	}

	// format bootnodes
	bns, err := stringsToAddrInfos(cfg.Bootnodes)
	if err != nil {
		return nil, fmt.Errorf("failed to format bootnodes: %w", err)
	}

	// create libp2p host instance
	h, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	ourCtx, cancel := context.WithCancel(cfg.Ctx)
	hst := &host{
		ctx:        ourCtx,
		cancel:     cancel,
		protocolID: fmt.Sprintf("%s/%s/%d", protocolID, cfg.Environment, cfg.ChainID),
		h:          h,
		handler:    cfg.Handler,
		bootnodes:  bns,
		queryBuf:   make([]byte, 2048),
	}

	hst.discovery, err = newDiscovery(ourCtx, h, hst.getBootnodes)
	if err != nil {
		return nil, err
	}

	return hst, nil
}

func (h *host) Start() error {
	h.h.SetStreamHandler(protocol.ID(h.protocolID+queryID), h.handleQueryStream)
	h.h.SetStreamHandler(protocol.ID(h.protocolID+swapID), h.handleProtocolStream)

	h.h.Network().SetConnHandler(h.handleConn)
	for _, addr := range h.multiaddrs() {
		log.Info("Started listening: address=", addr)
	}

	if err := h.bootstrap(); err != nil {
		return err
	}

	if err := h.discovery.start(); err != nil {
		return err
	}

	return nil
}

// close closes host services and the libp2p host (host services first)
func (h *host) Stop() error {
	h.cancel()

	if err := h.discovery.stop(); err != nil {
		return err
	}

	// close libp2p host
	if err := h.h.Close(); err != nil {
		log.Error("Failed to close libp2p host", "error", err)
		return err
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
func (h *host) Discover(provides common.ProvidesCoin, searchTime time.Duration) ([]peer.AddrInfo, error) {
	return h.discovery.discover(provides, searchTime)
}

// SendSwapMessage sends a message to the peer who we're currently doing a swap with.
func (h *host) SendSwapMessage(msg Message) error {
	h.swapMu.Lock()
	defer h.swapMu.Unlock()

	if h.swapStream == nil {
		return errors.New("no swap currently happening")
	}

	return h.writeToStream(h.swapStream, msg)
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
		"Sent message to peer=", s.Conn().RemotePeer(), " message=", msg.String(),
	)

	return nil
}

func (h *host) handleConn(conn libp2pnetwork.Conn) {
	log.Debug("incoming connection, peer=", conn.RemotePeer())
}

// readStream reads from the stream into the given buffer, returning the number of bytes read
func readStream(stream libp2pnetwork.Stream, buf []byte) (int, error) {
	if stream == nil {
		return 0, errors.New("stream is nil")
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
		return errors.New("failed to bootstrap to any bootnode")
	}

	return nil
}
