package net

import (
	crand "crypto/rand"
	"encoding/hex"
	"io"
	mrand "math/rand"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

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
