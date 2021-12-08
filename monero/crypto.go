package monero

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/noot/atomic-swap/common"

	ed25519 "filippo.io/edwards25519"
	"github.com/ebfe/keccak"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

const (
	addressPrefixMainnet  byte = 18
	addressPrefixStagenet byte = 24
)

// PublicSpendOnSecp256k1 returns a public spend key on the secp256k1 curve
func PublicSpendOnSecp256k1(k []byte) (x, y *big.Int) {
	return secp256k1.S256().ScalarBaseMult(k)
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

// Keccak256 returns the keccak256 hash of the data.
func Keccak256(data ...[]byte) (result [32]byte) {
	h := keccak.New256()
	for _, b := range data {
		h.Write(b)
	}
	r := h.Sum(nil)
	copy(result[:], r)
	return
}

func getChecksum(data ...[]byte) (result [4]byte) {
	keccak256 := Keccak256(data...)
	copy(result[:], keccak256[:4])
	return
}

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

// AddressBytes returns the address as bytes for a PrivateKeyPair with the given environment (ie. mainnet or stagenet)
func (kp *PrivateKeyPair) AddressBytes(env common.Environment) []byte {
	psk := kp.sk.Public().key.Bytes()
	pvk := kp.vk.Public().key.Bytes()
	c := append(psk, pvk...)

	var prefix byte
	switch env {
	case common.Mainnet, common.Development:
		prefix = addressPrefixMainnet
	case common.Stagenet:
		prefix = addressPrefixStagenet
	}

	// address encoding is:
	// 0x12+(32-byte public spend key) + (32-byte-byte public view key)
	// + First_4_Bytes(Hash(0x12+(32-byte public spend key) + (32-byte public view key)))
	checksum := getChecksum(append([]byte{prefix}, c...))
	addr := append(append([]byte{prefix}, c...), checksum[:4]...)
	return addr
}

// SpendKeyBytes returns the canoncail byte encoding of the private spend key.
func (kp *PrivateKeyPair) SpendKeyBytes() []byte {
	return kp.sk.key.Bytes()
}

// AddressBytes returns the base58-encoded address for a PrivateKeyPair with the given environment
// (ie. mainnet or stagenet)
func (kp *PrivateKeyPair) Address(env common.Environment) Address {
	return Address(EncodeMoneroBase58(kp.AddressBytes(env)))
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

// Marshal JSON-marshals the private key pair, providing its PrivateSpendKey, PrivateViewKey, Address,
// and Environment. This is intended to be written to a file, which someone can use to regenerate the wallet.
func (kp *PrivateKeyPair) Marshal(env common.Environment) ([]byte, error) {
	m := make(map[string]string)
	m["PrivateSpendKey"] = kp.sk.Hex()
	m["PrivateViewKey"] = kp.vk.Hex()
	m["Address"] = string(kp.Address(env))
	m["Environment"] = env.String()
	return json.Marshal(m)
}

// PrivateSpendKey represents a monero private spend key
type PrivateSpendKey struct {
	key *ed25519.Scalar
}

// NewPrivateSpendKey returns a new PrivateSpendKey from the given canonically-encoded scalar.
func NewPrivateSpendKey(b []byte) (*PrivateSpendKey, error) {
	if len(b) != privateKeySize {
		return nil, errors.New("input is not 32 bytes")
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
	h := Keccak256(k.key.Bytes())
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
	return Keccak256(k.key.Bytes())
}

// HashString returns the keccak256 of the secret key bytes as a hex encoded string
func (k *PrivateSpendKey) HashString() string {
	h := Keccak256(k.key.Bytes())
	return hex.EncodeToString(h[:])
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

// NewPrivateViewKeyFromHash derives a private view key given a hash of a private spend key.
// The input is a hex-encoded string.
func NewPrivateViewKeyFromHash(hash string) (*PrivateViewKey, error) {
	h, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	vk, err := ed25519.NewScalar().SetBytesWithClamping(h)
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

// AddressBytes returns the address as bytes for a PublicKeyPair with the given environment (ie. mainnet or stagenet)
func (kp *PublicKeyPair) AddressBytes(env common.Environment) []byte {
	psk := kp.sk.key.Bytes()
	pvk := kp.vk.key.Bytes()
	c := append(psk, pvk...)

	var prefix byte
	switch env {
	case common.Mainnet, common.Development:
		prefix = addressPrefixMainnet
	case common.Stagenet:
		prefix = addressPrefixStagenet
	}

	// address encoding is:
	// 0x12+(32-byte public spend key) + (32-byte-byte public view key)
	// + First_4_Bytes(Hash(0x12+(32-byte public spend key) + (32-byte public view key)))
	checksum := getChecksum(append([]byte{prefix}, c...))
	addr := append(append([]byte{prefix}, c...), checksum[:4]...)
	return addr
}

// AddressBytes returns the base58-encoded address for a PublicKeyPair with the given environment
// (ie. mainnet or stagenet)
func (kp *PublicKeyPair) Address(env common.Environment) Address {
	return Address(EncodeMoneroBase58(kp.AddressBytes(env)))
}

// GenerateKeys returns a private spend key and view key
func GenerateKeys() (*PrivateKeyPair, error) {
	var seed [64]byte
	_, err := rand.Read(seed[:])
	if err != nil {
		return nil, err
	}

	s, err := ed25519.NewScalar().SetUniformBytes(seed[:])
	if err != nil {
		return nil, fmt.Errorf("failed to set bytes: %w", err)
	}

	sk := &PrivateSpendKey{
		key: s,
	}

	return sk.AsPrivateKeyPair()
}
