package mcrypto

import (
	"encoding/hex"
	"errors"

	ed25519 "filippo.io/edwards25519"
)

// create a local types that we can marshal
type mScalar ed25519.Scalar
type mPoint ed25519.Point

// MarshalText returns the 64-symbol hex representation of the 32-byte k in little endian.
func (s *mScalar) MarshalText() ([]byte, error) {
	if s == nil {
		return nil, errors.New("can not marshal uninitialized scalar")
	}
	sBytes := (*ed25519.Scalar)(s).Bytes()
	return []byte(hex.EncodeToString(sBytes)), nil
}

// UnmarshalText assigns the scalar from hex input in little endian that is exactly 32
// bytes (64 hex symbols). The input is an ed25519 scalar and must already be reduced or
// we return an error.
func (s *mScalar) UnmarshalText(hexStr []byte) error {
	sBytes, err := hex.DecodeString(string(hexStr))
	if err != nil {
		return err
	}
	// SetCanonicalBytes will verify that we passed exactly 32 bytes
	sNew, err := (*ed25519.Scalar)(s).SetCanonicalBytes(sBytes)
	if err != nil {
		return err
	}
	*(*ed25519.Scalar)(s) = *sNew
	return err
}

// MarshalText returns the 64-symbol hex representation of the 32-byte k in little endian.
func (p *mPoint) MarshalText() ([]byte, error) {
	pBytes := (*ed25519.Point)(p).Bytes()
	return []byte(hex.EncodeToString(pBytes)), nil
}

// UnmarshalText assigns the scalar from hex input in little endian that is exactly 32
// bytes (64 hex symbols). The input is an ed25519 scalar and must already be reduced or
// we return an error.
func (p *mPoint) UnmarshalText(hexStr []byte) error {
	pointBytes, err := hex.DecodeString(string(hexStr))
	if err != nil {
		return err
	}
	_, err = (*ed25519.Point)(p).SetBytes(pointBytes)
	return err
}

// MarshalText returns the 64-symbol LE hex representation of k
func (k *PrivateSpendKey) MarshalText() ([]byte, error) {
	return (*mScalar)(k.key).MarshalText()
}

// UnmarshalText assigns k from LE hex input (64 symbols, 32 bytes).
func (k *PrivateSpendKey) UnmarshalText(input []byte) error {
	k.key = ed25519.NewScalar()
	return (*mScalar)(k.key).UnmarshalText(input)
}

// MarshalText returns the 64-symbol LE hex representation of k
func (k *PrivateViewKey) MarshalText() ([]byte, error) {
	return (*mScalar)(k.key).MarshalText()
}

// UnmarshalText assigns k from LE hex input (64 symbols, 32 bytes).
func (k *PrivateViewKey) UnmarshalText(input []byte) error {
	return (*mScalar)(k.key).UnmarshalText(input)
}

// MarshalText returns the 64-symbol LE hex representation of k
func (k *PublicKey) MarshalText() ([]byte, error) {
	return (*mPoint)(k.key).MarshalText()
}

// UnmarshalText assigns k from LE hex input (64 symbols, 32 bytes).
func (k *PublicKey) UnmarshalText(input []byte) error {
	return (*mPoint)(k.key).UnmarshalText(input)
}
