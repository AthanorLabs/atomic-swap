package secp256k1

import (
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/athanorlabs/atomic-swap/crypto"
)

const (
	//nolint:revive
	// https://github.com/bitcoin-core/secp256k1/blob/44c2452fd387f7ca604ab42d73746e7d3a44d8a2/include/secp256k1.h#L208
	Secp256k1TagPubkeyEven = byte(2)
	Secp256k1TagPubkeyOdd  = byte(3) //nolint:revive
)

var (
	errInvalidPubkeyLength = errors.New("encoded public key is not 64 bytes")
)

// PublicKey represents a secp256k1 public key
type PublicKey struct {
	x, y [32]byte // points stored in big-endian format
}

// NewPublicKey returns a new public key from the given (x, y) coordinates
func NewPublicKey(x, y [32]byte) *PublicKey {
	return &PublicKey{
		x: x,
		y: y,
	}
}

// NewPublicKeyFromBigInt returns a new public key from the given (x, y) coordinates
func NewPublicKeyFromBigInt(x, y *big.Int) *PublicKey {
	const ptSize = 32
	var xArray, yArray [ptSize]byte
	xSlice := x.Bytes()
	ySlice := y.Bytes()
	// Copying from a big-endian slice into a big-endian array, so we want padding bytes
	// on the left if the slice is shorter than the array.
	copy(xArray[ptSize-len(xSlice):], xSlice)
	copy(yArray[ptSize-len(ySlice):], ySlice)
	return NewPublicKey(xArray, yArray)
}

// NewPublicKeyFromHex returns a public key from a 64-byte hex encoded string
func NewPublicKeyFromHex(s string) (*PublicKey, error) {
	k, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	if len(k) != 64 {
		return nil, errInvalidPubkeyLength
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

// Compress returns the 33-byte compressed public key
func (k *PublicKey) Compress() [33]byte {
	cpk := [33]byte{}
	copy(cpk[1:33], k.x[:]) // pad x to the left if <32 bytes

	// check if y is odd
	// https://github.com/bitcoin-core/secp256k1/blob/1253a27756540d2ca526b2061d98d54868e9177c/src/field_10x26_impl.h#L315
	isOdd := k.y[31]&1 != 0
	if isOdd {
		cpk[0] = Secp256k1TagPubkeyOdd
	} else {
		cpk[0] = Secp256k1TagPubkeyEven
	}

	return cpk
}
