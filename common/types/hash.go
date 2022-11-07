package types

import (
	"encoding/hex"
	"fmt"
	"strings"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

// Hash represents a 32-byte hash
type Hash = ethcommon.Hash

// EmptyHash is an empty Hash
var EmptyHash = Hash{}

// IsHashZero returns true if the hash is all zeros, otherwise false
func IsHashZero(h Hash) bool {
	return h == EmptyHash
}

// HexToHash decodes a hex-encoded string into a hash
func HexToHash(s string) (Hash, error) {
	h, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil {
		return Hash{}, err
	}
	if len(h) != len(Hash{}) {
		return Hash{}, fmt.Errorf("invalid len=%d hash", len(h))
	}

	var hash Hash
	copy(hash[:], h)
	return hash, nil
}
