// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package mcrypto is for types and libraries that deal with Monero keys, addresses and
// signing.
package mcrypto

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/athanorlabs/atomic-swap/crypto"

	ed25519 "filippo.io/edwards25519"
)

const privateKeySize = 32

var (
	errInvalidInput = errors.New("input is not 32 bytes")
)

// PrivateKeyPair represents a monero private spend and view key.
type PrivateKeyPair struct {
	sk *PrivateSpendKey
	vk *PrivateViewKey
}

// NewPrivateKeyPair returns a new PrivateKeyPair from the given PrivateSpendKey and PrivateViewKey.
// Both values must be set, as no assumptions are made that the view key is derived from the spend
// key with this type.
func NewPrivateKeyPair(sk *PrivateSpendKey, vk *PrivateViewKey) *PrivateKeyPair {
	if sk == nil || vk == nil {
		panic("NewPrivateKeyPair requires a key pair")
	}
	return &PrivateKeyPair{
		sk: sk,
		vk: vk,
	}
}

// NewPrivateKeyPairFromBytes returns a new PrivateKeyPair given the canonical byte representation of
// a private spend and view key.
func NewPrivateKeyPairFromBytes(skBytes, vkBytes []byte) (*PrivateKeyPair, error) {
	if len(skBytes) != privateKeySize || len(vkBytes) != privateKeySize {
		return nil, errInvalidInput
	}

	sk, err := ed25519.NewScalar().SetCanonicalBytes(skBytes)
	if err != nil {
		return nil, err
	}

	vk, err := ed25519.NewScalar().SetCanonicalBytes(vkBytes)
	if err != nil {
		return nil, err
	}

	return &PrivateKeyPair{
		sk: &PrivateSpendKey{key: sk},
		vk: &PrivateViewKey{key: vk},
	}, nil
}

// SpendKeyBytes returns the canonical byte encoding of the private spend key.
func (kp *PrivateKeyPair) SpendKeyBytes() []byte {
	return kp.sk.key.Bytes()
}

// PublicKeyPair returns the PublicKeyPair corresponding to the PrivateKeyPair
func (kp *PrivateKeyPair) PublicKeyPair() *PublicKeyPair {
	return &PublicKeyPair{
		sk: kp.sk.Public(),
		vk: kp.vk.Public(),
	}
}

// SpendKey returns the key pair's spend key
func (kp *PrivateKeyPair) SpendKey() *PrivateSpendKey {
	return kp.sk
}

// ViewKey returns the key pair's view key
func (kp *PrivateKeyPair) ViewKey() *PrivateViewKey {
	return kp.vk
}

// PrivateSpendKey represents a monero private spend key
type PrivateSpendKey struct {
	key *ed25519.Scalar
}

// NewPrivateSpendKey returns a new PrivateSpendKey from the given canonically-encoded scalar.
func NewPrivateSpendKey(b []byte) (*PrivateSpendKey, error) {
	if len(b) != privateKeySize {
		return nil, errInvalidInput
	}

	sk, err := ed25519.NewScalar().SetCanonicalBytes(b)
	if err != nil {
		return nil, err
	}

	return &PrivateSpendKey{
		key: sk,
	}, nil
}

// Public returns the public key corresponding to the private key.
func (k *PrivateSpendKey) Public() *PublicKey {
	pk := ed25519.NewIdentityPoint().ScalarBaseMult(k.key)
	return &PublicKey{
		key: pk,
	}
}

// Hex formats the key as a hex string
func (k *PrivateSpendKey) Hex() string {
	return hex.EncodeToString(k.key.Bytes())
}

// String formats the key as a 0x-prefixed hex string
func (k *PrivateSpendKey) String() string {
	return "0x" + k.Hex()
}

// AsPrivateKeyPair returns the PrivateSpendKey as a PrivateKeyPair.
func (k *PrivateSpendKey) AsPrivateKeyPair() (*PrivateKeyPair, error) {
	vk, err := k.View()
	if err != nil {
		return nil, err
	}

	return &PrivateKeyPair{
		sk: k,
		vk: vk,
	}, nil
}

// View returns the private view key corresponding to the PrivateSpendKey.
func (k *PrivateSpendKey) View() (*PrivateViewKey, error) {
	h := crypto.Keccak256(k.key.Bytes())
	// We can't use SetBytesWithClamping below, which would do the sc_reduce32 computation
	// for us, because standard monero wallets do not modify the first and last byte when
	// calculating the view key.
	vkBytes := scReduce32(h)
	vk, err := ed25519.NewScalar().SetCanonicalBytes(vkBytes[:])
	if err != nil {
		return nil, err
	}

	return &PrivateViewKey{
		key: vk,
	}, nil
}

// Bytes returns the PrivateSpendKey as canonical bytes
func (k *PrivateSpendKey) Bytes() []byte {
	return k.key.Bytes()
}

// PrivateViewKey represents a monero private view key.
type PrivateViewKey struct {
	key *ed25519.Scalar
}

// Public returns the PublicKey corresponding to this PrivateViewKey.
func (k *PrivateViewKey) Public() *PublicKey {
	pk := ed25519.NewIdentityPoint().ScalarBaseMult(k.key)
	return &PublicKey{
		key: pk,
	}
}

// Bytes returns the canonical 32-byte little-endian encoding of PrivateViewKey.
func (k *PrivateViewKey) Bytes() []byte {
	return k.key.Bytes()
}

// Hex formats the key as a hex string
func (k *PrivateViewKey) Hex() string {
	return hex.EncodeToString(k.key.Bytes())
}

// String formats the key as a 0x-prefixed hex string
func (k *PrivateViewKey) String() string {
	return "0x" + k.Hex()
}

// PublicKey represents a monero public spend or view key.
type PublicKey struct {
	key *ed25519.Point
}

// NewPublicKeyFromBytes returns a new PublicKey from the given canonically-encoded point.
func NewPublicKeyFromBytes(b []byte) (*PublicKey, error) {
	k, err := ed25519.NewIdentityPoint().SetBytes(b)
	if err != nil {
		return nil, err
	}

	return &PublicKey{
		key: k,
	}, nil
}

// Bytes returns the canonical 32-byte, little-endian encoding of PublicKey.
func (k *PublicKey) Bytes() []byte {
	return k.key.Bytes()
}

// Hex formats the key as a hex string
func (k *PublicKey) Hex() string {
	return hex.EncodeToString(k.key.Bytes())
}

// String formats the key as a 0x-prefixed hex string
func (k *PublicKey) String() string {
	return "0x" + k.Hex()
}

// PublicKeyPair contains a public SpendKey and ViewKey
type PublicKeyPair struct {
	sk *PublicKey
	vk *PublicKey
}

// NewPublicKeyPair returns a new PublicKeyPair from the given public spend and view keys.
func NewPublicKeyPair(sk, vk *PublicKey) *PublicKeyPair {
	return &PublicKeyPair{
		sk: sk,
		vk: vk,
	}
}

// SpendKey returns the key pair's spend key.
func (kp *PublicKeyPair) SpendKey() *PublicKey {
	return kp.sk
}

// ViewKey returns the key pair's view key.
func (kp *PublicKeyPair) ViewKey() *PublicKey {
	return kp.vk
}

// GenerateKeys generates a private spend key and view key
func GenerateKeys() (*PrivateKeyPair, error) {
	var seed [32]byte
	_, err := rand.Read(seed[:])
	if err != nil {
		return nil, err
	}

	// we hash the seed for compatibility w/ the ed25519 stdlib
	h := sha512.Sum512(seed[:])

	s, err := ed25519.NewScalar().SetBytesWithClamping(h[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to set bytes: %w", err)
	}

	sk := &PrivateSpendKey{key: s}

	return sk.AsPrivateKeyPair()
}

// SumSpendAndViewKeys sums two PublicKeyPairs, returning another PublicKeyPair.
func SumSpendAndViewKeys(a, b *PublicKeyPair) *PublicKeyPair {
	return &PublicKeyPair{
		sk: SumPublicKeys(a.sk, b.sk),
		vk: SumPublicKeys(a.vk, b.vk),
	}
}

// SumPublicKeys sums two public keys (points)
func SumPublicKeys(a, b *PublicKey) *PublicKey {
	s := ed25519.NewIdentityPoint().Add(a.key, b.key)
	return &PublicKey{
		key: s,
	}
}

// SumPrivateSpendKeys sums two private spend keys (scalars)
func SumPrivateSpendKeys(a, b *PrivateSpendKey) *PrivateSpendKey {
	s := ed25519.NewScalar().Add(a.key, b.key)
	return &PrivateSpendKey{
		key: s,
	}
}

// SumPrivateViewKeys sums two private view keys (scalars)
func SumPrivateViewKeys(a, b *PrivateViewKey) *PrivateViewKey {
	s := ed25519.NewScalar().Add(a.key, b.key)
	return &PrivateViewKey{
		key: s,
	}
}
