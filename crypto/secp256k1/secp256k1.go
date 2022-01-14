package secp256k1

import (
	mcrypto "github.com/noot/atomic-swap/monero/crypto"
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

// Keccak256 returns the heccak256 hash of the x and y coordinates concatenated
func (k *PublicKey) Keccak256() [32]byte {
	return mcrypto.Keccak256(append(k.x[:], k.y[:]...))
}

// X returns the x-coordinate of the key
func (k *PublicKey) X() [32]byte {
	return k.x
}

// Y returns the y-coordinate of the key
func (k *PublicKey) Y() [32]byte {
	return k.y
}
