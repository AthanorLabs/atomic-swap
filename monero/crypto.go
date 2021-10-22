package monero

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	ed25519 "filippo.io/edwards25519"
	"github.com/ebfe/keccak"
)

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

func NewPrivateKeyPair(skBytes, vkBytes []byte) (*PrivateKeyPair, error) {
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
	checksum := getChecksum(append([]byte{0x18}, c...))
	addr := append(append([]byte{0x18}, c...), checksum[:4]...)
	return addr
}

func (kp *PrivateKeyPair) Address() Address {
	return Address(EncodeMoneroBase58(kp.AddressBytes()))
}

type PrivateSpendKey struct {
	key *ed25519.Scalar
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

type PublicKey struct {
	key *ed25519.Point
}

// PublicKeyPair contains a public SpendKey and ViewKey
type PublicKeyPair struct {
	sk *PublicKey
	vk *PublicKey
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
