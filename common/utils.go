package common

import (
	"crypto/ecdsa"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// Reverse returns a copy of the slice with the bytes in reverse order
func Reverse(s []byte) []byte {
	l := len(s)
	rs := make([]byte, l)
	for i := 0; i < l; i++ {
		rs[i] = s[l-i-1]
	}
	return rs
}

// EthereumPrivateKeyToAddress returns the address associated with a private key
func EthereumPrivateKeyToAddress(privkey *ecdsa.PrivateKey) ethcommon.Address {
	pub := privkey.Public().(*ecdsa.PublicKey)
	return ethcrypto.PubkeyToAddress(*pub)
}

// GetTopic returns the Ethereum topic (ie. keccak256 hash) of the given event or function signature string.
func GetTopic(sig string) ethcommon.Hash {
	h := ethcrypto.Keccak256([]byte(sig))
	var b [32]byte
	copy(b[:], h)
	return b
}
