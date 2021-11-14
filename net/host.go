package net

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	mrand "math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/noot/atomic-swap/common"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	libp2phost "github.com/libp2p/go-libp2p-core/host"
	libp2pnetwork "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
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

type MessageInfo struct {
	Message Message
	Who     peer.ID
}

type Host interface {
	Start() error
	Stop() error
	SetOutgoingCh(<-chan *MessageInfo)
	ReceivedMessageCh() <-chan *MessageInfo
	SetNextExpectedMessage(m Message)

	Discover(provides ProvidesCoin, searchTime time.Duration) ([]peer.AddrInfo, error)
}

type host struct {
	ctx    context.Context
	cancel context.CancelFunc

	h            libp2phost.Host
	helloMessage *HelloMessage
	discovery    *discovery

	bootnodes []peer.AddrInfo
	// messages received from the rest of the program, to be sent out
	outCh <-chan *MessageInfo

	// messages received from the network, to be sent to the rest of the program
	inCh chan *MessageInfo

	// next expected message from the network
	// empty, is just used for type matching
	nextExpectedMessage Message

	queryMu  sync.Mutex
	queryBuf []byte
}

// Config is used to configure the network Host.
type Config struct {
	Ctx           context.Context
	Port          uint64
	Provides      []ProvidesCoin
	MaximumAmount []uint64
	ExchangeRate  common.ExchangeRate
	KeyFile       string
	Bootnodes     []string
}

func NewHost(cfg *Config) (*host, error) {
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

	inCh := make(chan *MessageInfo)

	ourCtx, cancel := context.WithCancel(cfg.Ctx)
	hst := &host{
		ctx:    ourCtx,
		cancel: cancel,
		h:      h,
		helloMessage: &HelloMessage{
			Provides:      cfg.Provides,
			MaximumAmount: cfg.MaximumAmount,
			ExchangeRate:  cfg.ExchangeRate,
		},
		bootnodes: bns,
		inCh:      inCh,
		queryBuf:  make([]byte, 2048),
	}

	hst.discovery, err = newDiscovery(ourCtx, h, hst.getBootnodes, cfg.Provides...)
	if err != nil {
		return nil, err
	}

	return hst, nil
}

func (h *host) Discover(provides ProvidesCoin, searchTime time.Duration) ([]peer.AddrInfo, error) {
	return h.discovery.discover(provides, searchTime)
}

func (h *host) getBootnodes() []peer.AddrInfo {
	addrs := h.bootnodes
	for _, p := range h.h.Network().Peers() {
		addrs = append(addrs, h.h.Peerstore().PeerInfo(p))
	}
	return addrs
}

func (h *host) SetOutgoingCh(ch <-chan *MessageInfo) {
	h.outCh = ch
}

func (h *host) SetNextExpectedMessage(m Message) {
	h.nextExpectedMessage = m
}

func (h *host) Start() error {
	h.nextExpectedMessage = &HelloMessage{}
	h.h.SetStreamHandler(protocolID, h.handleStream)
	h.h.SetStreamHandler(protocolID+queryID, h.handleQueryStream)

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

// close closes host services and the libp2p host (host services first)
func (h *host) Stop() error {
	h.cancel()

	// close libp2p host
	if err := h.h.Close(); err != nil {
		log.Error("Failed to close libp2p host", "error", err)
		return err
	}

	return nil
}

func (h *host) SendMessage(to peer.ID, msg Message) error {
	_, err := h.send(to, msg)
	return err
}

func (h *host) ReceivedMessageCh() <-chan *MessageInfo {
	return h.inCh
}

// send creates a new outbound stream with the given peer and writes the message. It also returns
// the newly created stream.
func (h *host) send(p peer.ID, msg Message) (libp2pnetwork.Stream, error) {
	// open outbound stream with host protocol id
	stream, err := h.h.NewStream(h.ctx, p, protocolID)
	if err != nil {
		log.Debug("failed to open new stream with peer", "peer", p, "error", err)
		return nil, err
	}

	log.Debug(
		"Opened stream, peer=", p,
	)

	err = h.writeToStream(stream, msg)
	if err != nil {
		return nil, err
	}

	return stream, nil
}

func (h *host) writeToStream(s libp2pnetwork.Stream, msg Message) error {
	//defer s.Close()

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

	stream, err := h.send(conn.RemotePeer(), h.helloMessage)
	if err != nil {
		log.Info("failed to send message, closing stream")
		_ = stream.Close()
		return
	}

	go h.handleStream(stream)
}

func (h *host) handleStream(stream libp2pnetwork.Stream) {
	msgBytes := make([]byte, 2048)

	for {
		tot, err := readStream(stream, msgBytes[:])
		if errors.Is(err, io.EOF) {
			return
		} else if err != nil {
			//log.Debug("failed to read from stream", "id", stream.ID(), "peer", stream.Conn().RemotePeer(), "protocol", stream.Protocol(), "error", err)
			_ = stream.Close()
			return
		}

		// decode message based on message type
		msg, err := h.decodeMessage(msgBytes[:tot])
		if err != nil {
			log.Debug("failed to decode message from peer, id=", stream.ID(), " protocol=", stream.Protocol(), " err=", err)
			continue
		}

		log.Debug(
			"received message from peer, peer=", stream.Conn().RemotePeer(), " msg=", msg.String(),
		)

		h.handleMessage(stream, msg)
	}
}

func (h *host) decodeMessage(b []byte) (Message, error) {
	switch h.nextExpectedMessage.(type) {
	case *HelloMessage:
		var m *HelloMessage
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		return m, nil
	case *SendKeysMessage:
		var m *SendKeysMessage
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		return m, nil
	case *NotifyContractDeployed:
		var m *NotifyContractDeployed
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		return m, nil
	case *NotifyXMRLock:
		var m *NotifyXMRLock
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		return m, nil
	case *NotifyClaimed:
		var m *NotifyClaimed
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		return m, nil
	default:
		return nil, errors.New("not expecting any more messages")
	}
}

func (h *host) handleMessage(stream libp2pnetwork.Stream, m Message) {
	h.inCh <- &MessageInfo{
		Message: m,
		Who:     stream.Conn().RemotePeer(),
	}

	next := <-h.outCh
	if next == nil {
		fmt.Println("no more outgoing messages")
		return
	}

	if next.Who != stream.Conn().RemotePeer() {
		fmt.Println("peer ID mismatch")
	}

	if err := h.writeToStream(stream, next.Message); err != nil {
		fmt.Println("failed to write to stream")
		return
	}
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
		return 0, nil // msg length of 0 is allowed, for example transactions handshake
	}

	if length > uint64(len(buf)) {
		log.Warn("received message with size greater than allocated message buffer", "length", length, "buffer size", len(buf))
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

// stringToAddrInfos converts a single string peer id to AddrInfo
func stringToAddrInfo(s string) (peer.AddrInfo, error) {
	maddr, err := ma.NewMultiaddr(s)
	if err != nil {
		return peer.AddrInfo{}, err
	}
	p, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return peer.AddrInfo{}, err
	}
	return *p, err
}

// stringsToAddrInfos converts a string of peer ids to AddrInfo
func stringsToAddrInfos(peers []string) ([]peer.AddrInfo, error) {
	pinfos := make([]peer.AddrInfo, len(peers))
	for i, p := range peers {
		p, err := stringToAddrInfo(p)
		if err != nil {
			return nil, err
		}
		pinfos[i] = p
	}
	return pinfos, nil
}

// generateKey generates an ed25519 private key and writes it to the data directory
// If the seed is zero, we use real cryptographic randomness. Otherwise, we use a
// deterministic randomness source to make keys the same across multiple runs.
func generateKey(seed int64, fp string) (crypto.PrivKey, error) {
	var r io.Reader
	if seed == 0 {
		r = crand.Reader
	} else {
		r = mrand.New(mrand.NewSource(seed)) //nolint
	}
	key, _, err := crypto.GenerateEd25519Key(r)
	if err != nil {
		return nil, err
	}
	if seed == 0 {
		if err = saveKey(key, fp); err != nil {
			return nil, err
		}
	}
	return key, nil
}

// loadKey attempts to load a private key from the provided filepath
func loadKey(fp string) (crypto.PrivKey, error) {
	keyData, err := ioutil.ReadFile(filepath.Clean(fp))
	if err != nil {
		return nil, err
	}
	dec := make([]byte, hex.DecodedLen(len(keyData)))
	_, err = hex.Decode(dec, keyData)
	if err != nil {
		return nil, err
	}
	return crypto.UnmarshalEd25519PrivateKey(dec)
}

// saveKey attempts to save a private key to the provided filepath
func saveKey(priv crypto.PrivKey, fp string) (err error) {
	f, err := os.Create(filepath.Clean(fp))
	if err != nil {
		return err
	}
	raw, err := priv.Raw()
	if err != nil {
		return err
	}
	enc := make([]byte, hex.EncodedLen(len(raw)))
	hex.Encode(enc, raw)
	if _, err = f.Write(enc); err != nil {
		return err
	}
	return f.Close()
}

func uint64ToLEB128(in uint64) []byte {
	var out []byte
	for {
		b := uint8(in & 0x7f)
		in >>= 7
		if in != 0 {
			b |= 0x80
		}
		out = append(out, b)
		if in == 0 {
			break
		}
	}
	return out
}

func readLEB128ToUint64(r io.Reader, buf []byte) (uint64, int, error) {
	if len(buf) == 0 {
		return 0, 0, errors.New("buffer has length 0")
	}

	var out uint64
	var shift uint

	maxSize := 10 // Max bytes in LEB128 encoding of uint64 is 10.
	bytesRead := 0

	for {
		n, err := r.Read(buf[:1])
		if err != nil {
			return 0, bytesRead, err
		}

		bytesRead += n

		b := buf[0]
		out |= uint64(0x7F&b) << shift
		if b&0x80 == 0 {
			break
		}

		maxSize--
		if maxSize == 0 {
			return 0, bytesRead, fmt.Errorf("invalid LEB128 encoded data")
		}

		shift += 7
	}
	return out, bytesRead, nil
}
