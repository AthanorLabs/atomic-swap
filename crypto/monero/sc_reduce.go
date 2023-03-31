// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package mcrypto

import (
	"math/big"

	"github.com/athanorlabs/atomic-swap/common"
)

// curveOrder, often called "l", is the prime used by ed25519
var curveOrder *big.Int

func init() {
	// python3 -c 'print((2**252 + 27742317777372353535851937790883648493).to_bytes(32, "big").hex())'
	const lHex = "1000000000000000000000000000000014def9dea2f79cd65812631a5cf5d3ed"
	var ok bool
	curveOrder, ok = new(big.Int).SetString(lHex, 16)
	if !ok {
		panic("invalid hex constant")
	}
}

// scReduce32 reduces the 32-byte little endian input s by computing and returning
// s mod l, where l is ed25519 curve order prime.
func scReduce32(s [32]byte) [32]byte {
	scalar := new(big.Int).SetBytes(common.Reverse(s[:]))
	reduced := common.Reverse(new(big.Int).Mod(scalar, curveOrder).Bytes())
	var reduced32 [32]byte
	copy(reduced32[:], reduced) // little endian, so high order byte padding is automatic
	return reduced32
}
