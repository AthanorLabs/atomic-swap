package crypto

import (
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// Keccak256 returns the keccak256 hash of the data.
func Keccak256(data ...[]byte) (result [32]byte) {
	copy(result[:], ethcrypto.Keccak256(data...))
	return
}
