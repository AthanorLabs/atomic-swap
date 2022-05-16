package mcrypto

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/crypto"

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
// It does not validate if the view key corresponds to the spend key.
func NewPrivateKeyPair(sk *PrivateSpendKey, vk *PrivateViewKey) *PrivateKeyPair {
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

// SpendKeyBytes returns the canoncail byte encoding of the private spend key.
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

// SpendKey returns the keypair's spend key
func (kp *PrivateKeyPair) SpendKey() *PrivateSpendKey {
	return kp.sk
}

// ViewKey returns the keypair's view key
func (kp *PrivateKeyPair) ViewKey() *PrivateViewKey {
	return kp.vk
}

// PrivateKeyInfo ...
type PrivateKeyInfo struct {
	PrivateSpendKey string
	PrivateViewKey  string
	Address         string
	Environment     string
}

// Info return the private key pair as PrivateKeyInfo, providing its PrivateSpendKey, PrivateViewKey, Address,
// and Environment. This is intended to be written to a file, which someone can use to regenerate the wallet.
func (kp *PrivateKeyPair) Info(env common.Environment) *PrivateKeyInfo {
	return &PrivateKeyInfo{
		PrivateSpendKey: kp.sk.Hex(),
		PrivateViewKey:  kp.vk.Hex(),
		Address:         string(kp.Address(env)),
		Environment:     env.String(),
	}
}

// PrivateSpendKey represents a monero private spend key
type PrivateSpendKey struct {
	seed [32]byte
	key  *ed25519.Scalar
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

// Hex returns the hex-encoded canonical byte representation of the PrivateSpendKey.
func (k *PrivateSpendKey) Hex() string {
	return hex.EncodeToString(k.key.Bytes())
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
	vk, err := ed25519.NewScalar().SetBytesWithClamping(h[:])
	if err != nil {
		return nil, err
	}

	return &PrivateViewKey{
		key: vk,
	}, nil
}

// Hash returns the keccak256 of the secret key bytes
func (k *PrivateSpendKey) Hash() [32]byte {
	return crypto.Keccak256(k.key.Bytes())
}

// HashString returns the keccak256 of the secret key bytes as a hex encoded string
func (k *PrivateSpendKey) HashString() string {
	h := crypto.Keccak256(k.key.Bytes())
	return hex.EncodeToString(h[:])
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

// Hex returns the hex-encoded canonical byte representation of the PrivateViewKey.
func (k *PrivateViewKey) Hex() string {
	return hex.EncodeToString(k.key.Bytes())
}

// NewPrivateViewKeyFromHex returns a new PrivateViewKey from the given canonically- and hex-encoded scalar.
func NewPrivateViewKeyFromHex(vkHex string) (*PrivateViewKey, error) {
	vkBytes, err := hex.DecodeString(vkHex)
	if err != nil {
		return nil, err
	}

	vk, err := ed25519.NewScalar().SetCanonicalBytes(vkBytes)
	if err != nil {
		return nil, err
	}

	return &PrivateViewKey{
		key: vk,
	}, nil
}

// PublicKey represents a monero public spend or view key.
type PublicKey struct {
	key *ed25519.Point
}

// NewPublicKeyFromHex returns a new PublicKey from the given canonically- and hex-encoded point.
func NewPublicKeyFromHex(s string) (*PublicKey, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	k, err := ed25519.NewIdentityPoint().SetBytes(b)
	if err != nil {
		return nil, err
	}

	return &PublicKey{
		key: k,
	}, nil
}

// Hex returns the hex-encoded canonical byte representation of the PublicKey.
func (k *PublicKey) Hex() string {
	return hex.EncodeToString(k.key.Bytes())
}

// Bytes returns the canonical byte representation of the PublicKey.
func (k *PublicKey) Bytes() []byte {
	return k.key.Bytes()
}

// PublicKeyPair contains a public SpendKey and ViewKey
type PublicKeyPair struct {
	sk *PublicKey
	vk *PublicKey
}

// NewPublicKeyPairFromHex returns a new PublicKeyPair from the given canonically- and hex-encoded points.
func NewPublicKeyPairFromHex(skHex, vkHex string) (*PublicKeyPair, error) {
	skBytes, err := hex.DecodeString(skHex)
	if err != nil {
		return nil, err
	}

	vkBytes, err := hex.DecodeString(vkHex)
	if err != nil {
		return nil, err
	}

	sk, err := ed25519.NewIdentityPoint().SetBytes(skBytes)
	if err != nil {
		return nil, err
	}

	vk, err := ed25519.NewIdentityPoint().SetBytes(vkBytes)
	if err != nil {
		return nil, err
	}

	return &PublicKeyPair{
		sk: &PublicKey{key: sk},
		vk: &PublicKey{key: vk},
	}, nil
}

// NewPublicKeyPair returns a new PublicKeyPair from the given public spend and view keys.
func NewPublicKeyPair(sk, vk *PublicKey) *PublicKeyPair {
	return &PublicKeyPair{
		sk: sk,
		vk: vk,
	}
}

// SpendKey returns the keypair's spend key.
func (kp *PublicKeyPair) SpendKey() *PublicKey {
	return kp.sk
}

// ViewKey returns the keypair's view key.
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

	sk := &PrivateSpendKey{
		seed: seed,
		key:  s,
	}

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
