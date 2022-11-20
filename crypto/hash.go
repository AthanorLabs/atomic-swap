// Package crypto is for an ethereum hash (Keccak256) wrapper that returns a 32-byte array
// instead of 32-byte slice. TODO: Should this be combined with the secp256k1 package into
// an ecrypto package? (Just like we have mcrypto for Monero.)
package crypto

import (
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// Keccak256 returns the keccak256 hash of the data.
func Keccak256(data ...[]byte) (result [32]byte) {
	copy(result[:], ethcrypto.Keccak256(data...))
	return
}
