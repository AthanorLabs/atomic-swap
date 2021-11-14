package net

import (
	crand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	mrand "math/rand"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

// StringToAddrInfos converts a single string peer id to AddrInfo
func StringToAddrInfo(s string) (peer.AddrInfo, error) {
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
		p, err := StringToAddrInfo(p)
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
