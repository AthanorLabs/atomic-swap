package secp256k1

import (
	"encoding/hex"
	"errors"

	"github.com/noot/atomic-swap/crypto"
)

// PublicKey represents a secp256k1 public key
type PublicKey struct {
	x, y [32]byte
}

// NewPublicKey returns a new public key from the given (x, y) coordinates
func NewPublicKey(x, y [32]byte) *PublicKey {
	return &PublicKey{
		x: x,
		y: y,
	}
}

// NewPublicKeyFromHex returns a public key from a 64-byte hex encoded string
func NewPublicKeyFromHex(s string) (*PublicKey, error) {
	k, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	if len(k) != 64 {
		return nil, errors.New("encoded public key is not 64 bytes")
	}

	pk := &PublicKey{}
	copy(pk.x[:], k[:32])
	copy(pk.y[:], k[32:])
	return pk, nil
}

// Keccak256 returns the heccak256 hash of the x and y coordinates concatenated
func (k *PublicKey) Keccak256() [32]byte {
	return crypto.Keccak256(append(k.x[:], k.y[:]...))
}

// X returns the x-coordinate of the key
func (k *PublicKey) X() [32]byte {
	return k.x
}

// Y returns the y-coordinate of the key
func (k *PublicKey) Y() [32]byte {
	return k.y
}

// String returns the key as a 64-byte hex encoded string
func (k *PublicKey) String() string {
	return hex.EncodeToString(append(k.x[:], k.y[:]...))
}
