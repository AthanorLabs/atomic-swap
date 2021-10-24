package monero

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	ed25519 "filippo.io/edwards25519"
	"github.com/ebfe/keccak"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

const (
	addressPrefixMainnet byte = 0x12
)

func PublicSpendOnSecp256k1(k []byte) (a, b *big.Int) {
	return secp256k1.S256().ScalarBaseMult(k)
}

func SumSpendAndViewKeys(a, b *PublicKeyPair) *PublicKeyPair {
	return &PublicKeyPair{
		sk: SumPublicKeys(a.sk, b.sk),
		vk: SumPublicKeys(a.vk, b.vk),
	}
}

// Sum sums two public keys (points)
func SumPublicKeys(a, b *PublicKey) *PublicKey {
	s := ed25519.NewIdentityPoint().Add(a.key, b.key)
	return &PublicKey{
		key: s,
	}
}

// Sum sums two private spend keys (scalars)
func SumPrivateSpendKeys(a, b *PrivateSpendKey) *PrivateSpendKey {
	s := ed25519.NewScalar().Add(a.key, b.key)
	return &PrivateSpendKey{
		key: s,
	}
}

// Sum sums two private view keys (scalars)
func SumPrivateViewKeys(a, b *PrivateViewKey) *PrivateViewKey {
	s := ed25519.NewScalar().Add(a.key, b.key)
	return &PrivateViewKey{
		key: s,
	}
}

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

type PrivateKeyPair struct {
	sk *PrivateSpendKey
	vk *PrivateViewKey
}

func NewPrivateKeyPair(sk *PrivateSpendKey, vk *PrivateViewKey) *PrivateKeyPair {
	return &PrivateKeyPair{
		sk: sk,
		vk: vk,
	}
}

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

func (kp *PrivateKeyPair) AddressBytes() []byte {
	psk := kp.sk.Public().key.Bytes()
	pvk := kp.vk.Public().key.Bytes()
	c := append(psk, pvk...)

	// address encoding is:
	// 0x12+(32-byte public spend key) + (32-byte-byte public view key)
	// + First_4_Bytes(Hash(0x12+(32-byte public spend key) + (32-byte public view key)))
	checksum := getChecksum(append([]byte{addressPrefixMainnet}, c...))
	addr := append(append([]byte{addressPrefixMainnet}, c...), checksum[:4]...)
	return addr
}

func (kp *PrivateKeyPair) SpendKeyBytes() []byte {
	return kp.sk.key.Bytes()
}

func (kp *PrivateKeyPair) Address() Address {
	return Address(EncodeMoneroBase58(kp.AddressBytes()))
}

func (kp *PrivateKeyPair) PublicKeyPair() *PublicKeyPair {
	return &PublicKeyPair{
		sk: kp.sk.Public(),
		vk: kp.vk.Public(),
	}
}

func (kp *PrivateKeyPair) SpendKey() *PrivateSpendKey {
	return kp.sk
}

func (kp *PrivateKeyPair) ViewKey() *PrivateViewKey {
	return kp.vk
}

type PrivateSpendKey struct {
	key *ed25519.Scalar
}

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

func (k *PrivateSpendKey) Public() *PublicKey {
	pk := ed25519.NewIdentityPoint().ScalarBaseMult(k.key)
	return &PublicKey{
		key: pk,
	}
}

func (k *PrivateSpendKey) Hex() string {
	return hex.EncodeToString(k.key.Bytes())
}

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

type PrivateViewKey struct {
	key *ed25519.Scalar
}

func (k *PrivateViewKey) Public() *PublicKey {
	pk := ed25519.NewIdentityPoint().ScalarBaseMult(k.key)
	return &PublicKey{
		key: pk,
	}
}

func (k *PrivateViewKey) Hex() string {
	return hex.EncodeToString(k.key.Bytes())
}

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

type PublicKey struct {
	key *ed25519.Point
}

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

func (k *PublicKey) Hex() string {
	return hex.EncodeToString(k.key.Bytes())
}

func (k *PublicKey) Bytes() []byte {
	return k.key.Bytes()
}

// PublicKeyPair contains a public SpendKey and ViewKey
type PublicKeyPair struct {
	sk *PublicKey
	vk *PublicKey
}

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

func NewPublicKeyPair(sk, vk *PublicKey) *PublicKeyPair {
	return &PublicKeyPair{
		sk: sk,
		vk: vk,
	}
}

func (kp *PublicKeyPair) SpendKey() *PublicKey {
	return kp.sk
}

func (kp *PublicKeyPair) ViewKey() *PublicKey {
	return kp.vk
}

func (kp *PublicKeyPair) AddressBytes() []byte {
	psk := kp.sk.key.Bytes()
	pvk := kp.vk.key.Bytes()
	c := append(psk, pvk...)

	// address encoding is:
	// 0x12+(32-byte public spend key) + (32-byte-byte public view key)
	// + First_4_Bytes(Hash(0x12+(32-byte public spend key) + (32-byte public view key)))
	checksum := getChecksum(append([]byte{addressPrefixMainnet}, c...))
	addr := append(append([]byte{addressPrefixMainnet}, c...), checksum[:4]...)
	return addr
}

func (kp *PublicKeyPair) Address() Address {
	return Address(EncodeMoneroBase58(kp.AddressBytes()))
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

// GenerateKeys returns a private spend key and view key
func GenerateKeysTruncated() (*PrivateKeyPair, error) {
	var seed [64]byte
	_, err := rand.Read(seed[:])
	if err != nil {
		return nil, err
	}

	s, err := ed25519.NewScalar().SetUniformBytes(seed[:])
	if err != nil {
		return nil, fmt.Errorf("failed to set bytes: %w", err)
	}

	sBytes := s.Bytes()
	fmt.Println("sBytes pre-zeroization:  ", sBytes)
	sBytes[31] &= 0x0f
	fmt.Println("sBytes post-zeroization: ", sBytes)
	s, err = ed25519.NewScalar().SetCanonicalBytes(sBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to set bytes: %w", err)
	}

	sk := &PrivateSpendKey{
		key: s,
	}

	return sk.AsPrivateKeyPair()
}
