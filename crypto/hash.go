// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

// Package crypto is for cryptographic code used by both Monero and Ethereum.
// Chain specific crypto is in subpackages.
package crypto

import (
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// Keccak256 returns the keccak256 hash of the data.
func Keccak256(data ...[]byte) (result [32]byte) {
	copy(result[:], ethcrypto.Keccak256(data...))
	return
}
