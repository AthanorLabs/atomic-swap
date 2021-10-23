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

	log "github.com/ChainSafe/log15"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	libp2phost "github.com/libp2p/go-libp2p-core/host"
	libp2pnetwork "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	protocolID = "/eth-xmr-atomic-swap/0"
	maxReads   = 128
)

var logger = log.New("pkg", "net")
var _ Host = &host{}

type ReceivedMessage struct {
	Message Message
	Who     peer.ID
}

type Host interface {
	Start() error
	Stop() error
	SetOutgoingCh(<-chan *ReceivedMessage)
	ReceivedMessageCh() <-chan *ReceivedMessage
}

type host struct {
	ctx    context.Context
	cancel context.CancelFunc

	h           libp2phost.Host
	wantMessage *WantMessage
	//mdns *mdns
	bootnodes []peer.AddrInfo
	// messages received from the rest of the program, to be sent out
	outCh <-chan *ReceivedMessage

	// messages received from the network, to be sent to the rest of the program
	inCh chan *ReceivedMessage

	// next expected message from the network
	// empty, is just used for type matching
	nextExpectedMessage Message
}

func NewHost(port uint64, want, keyfile string, bootnodes []string) (*host, error) {
	ctx, cancel := context.WithCancel(context.Background())

	key, err := loadKey(keyfile)
	if err != nil {
		fmt.Println("failed to load libp2p key, generating key...", keyfile)
		key, err = generateKey(0, keyfile)
		if err != nil {
			return nil, err
		}
	}

	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
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
	bns, err := stringsToAddrInfos(bootnodes)
	if err != nil {
		return nil, err
	}

	// create libp2p host instance
	h, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	inCh := make(chan *ReceivedMessage)

	return &host{
		ctx:         ctx,
		cancel:      cancel,
		h:           h,
		wantMessage: &WantMessage{Want: want},
		//mdns: newMDNS(h),
		//discovery: discovery,
		bootnodes: bns,
		inCh:      inCh,
	}, nil
}

func (h *host) SetOutgoingCh(ch <-chan *ReceivedMessage) {
	h.outCh = ch
}

func (h *host) Start() error {
	h.h.SetStreamHandler(protocolID, h.handleStream)
	h.h.Network().SetConnHandler(h.handleConn)
	h.bootstrap()
	//h.mdns.start()

	return nil
}

// close closes host services and the libp2p host (host services first)
func (h *host) Stop() error {
	h.cancel()

	// close libp2p host
	if err := h.h.Close(); err != nil {
		logger.Error("Failed to close libp2p host", "error", err)
		return err
	}

	return nil
}

func (h *host) SendMessage(to peer.ID, msg Message) error {
	_, err := h.send(to, msg)
	return err
}

func (h *host) ReceivedMessageCh() <-chan *ReceivedMessage {
	return h.outCh
}

// send creates a new outbound stream with the given peer and writes the message. It also returns
// the newly created stream.
func (h *host) send(p peer.ID, msg Message) (libp2pnetwork.Stream, error) {
	// open outbound stream with host protocol id
	stream, err := h.h.NewStream(h.ctx, p, protocolID)
	if err != nil {
		logger.Trace("failed to open new stream with peer", "peer", p, "error", err)
		return nil, err
	}

	logger.Trace(
		"Opened stream",
		"peer", p,
	)

	err = h.writeToStream(stream, msg)
	if err != nil {
		return nil, err
	}

	logger.Trace(
		"Sent message to peer",
		"peer", p,
		"message", msg.String(),
	)

	return stream, nil
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

	return nil
}

func (h *host) handleConn(conn libp2pnetwork.Conn) {
	fmt.Printf("incoming connection, peer=%s\n", conn.RemotePeer())

	stream, err := h.send(conn.RemotePeer(), h.wantMessage)
	if err != nil {
		fmt.Printf("failed to send message, closing stream")
		_ = stream.Close()
		return
	}

	go h.handleStream(stream)
}

func (h *host) handleStream(stream libp2pnetwork.Stream) {
	fmt.Printf("incoming stream, peer=%s\n", stream.Conn().RemotePeer())
	msgBytes := make([]byte, 2048)

	for {
		tot, err := readStream(stream, msgBytes[:])
		if errors.Is(err, io.EOF) {
			return
		} else if err != nil {
			logger.Trace("failed to read from stream", "id", stream.ID(), "peer", stream.Conn().RemotePeer(), "protocol", stream.Protocol(), "error", err)
			_ = stream.Close()
			return
		}

		// decode message based on message type
		msg, err := h.decodeMessage(msgBytes[:tot])
		if err != nil {
			logger.Trace("failed to decode message from peer", "id", stream.ID(), "protocol", stream.Protocol(), "err", err)
			continue
		}

		logger.Trace(
			"received message from peer",
			"peer", stream.Conn().RemotePeer(),
			"msg", msg.String(),
		)

		h.handleMessage(stream, msg)
	}
}

func (h *host) decodeMessage(b []byte) (Message, error) {
	switch h.nextExpectedMessage.(type) {
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
	default:
		return nil, errors.New("not expecting any more messages")
	}
}

func (h *host) handleMessage(stream libp2pnetwork.Stream, m Message) {
	h.inCh <- &ReceivedMessage{
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
		logger.Warn("received message with size greater than allocated message buffer", "length", length, "buffer size", len(buf))
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
func (h *host) bootstrap() {
	failed := 0
	for _, addrInfo := range h.bootnodes {
		logger.Debug("bootstrapping to peer", "peer", addrInfo.ID)
		err := h.h.Connect(h.ctx, addrInfo)
		if err != nil {
			logger.Debug("failed to bootstrap to peer", "error", err)
			failed++
		}
	}
	if failed == len(h.bootnodes) && len(h.bootnodes) != 0 {
		logger.Error("failed to bootstrap to any bootnode")
	}
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
		if err = makeDir(fp); err != nil {
			return nil, err
		}
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

// makeDir makes directory if directory does not already exist
func makeDir(fp string) error {
	_, e := os.Stat(fp)
	if os.IsNotExist(e) {
		e = os.Mkdir(fp, os.ModePerm)
		if e != nil {
			return e
		}
	}
	return e
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
