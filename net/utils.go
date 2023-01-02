package net

import (
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	mrand "math/rand"
	"net"
	"os"
	"path/filepath"

	"github.com/athanorlabs/atomic-swap/net/message"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

func WriteStreamMessage(s io.Writer, msg message.Message, peerID peer.ID) error {
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
func ReadStreamMessage(s io.Reader, maxMessageSize uint32) (message.Message, error) {
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

// stringsToAddrInfos converts a string of peers in multiaddress format to a
// minimal set of multiaddr addresses.
func stringsToAddrInfos(peers []string) ([]peer.AddrInfo, error) {
	madders := make([]ma.Multiaddr, len(peers))
	for i, p := range peers {
		ma, err := ma.NewMultiaddr(p)
		if err != nil {
			return nil, err
		}
		madders[i] = ma
	}
	return peer.AddrInfosFromP2pAddrs(madders...)
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
	keyData, err := os.ReadFile(filepath.Clean(fp))
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
