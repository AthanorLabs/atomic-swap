package secp256k1

import (
	mcrypto "github.com/noot/atomic-swap/monero/crypto"
)

type PublicKey struct {
	x, y [32]byte
}

func NewPublicKey(x, y [32]byte) *PublicKey {
	return &PublicKey{
		x: x,
		y: y,
	}
}

func (k *PublicKey) Keccak256() [32]byte {
	return mcrypto.Keccak256(append(k.x[:], k.y[:]...))
}

func (k *PublicKey) X() [32]byte {
	return k.x
}

func (k *PublicKey) Y() [32]byte {
	return k.y
}
