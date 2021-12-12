package crypto

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
)

// Signature represents an ed25519 signature
type Signature struct {
	s []byte
}

// NewSignatureFromHex returns a new Signature from the given hex-encoded string.
// The string must be 64 bytes.
func NewSignatureFromHex(s string) (*Signature, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	if len(b) != ed25519.SignatureSize {
		return nil, errors.New("invalid length for signature")
	}

	return &Signature{
		s: b,
	}, nil
}

// Hex returns the signature as a hex-encoded string.
func (s *Signature) Hex() string {
	return hex.EncodeToString(s.s)
}

// Sign signs the given message with the private key.
// The private key must have been created with GenerateKeys().
func (k *PrivateSpendKey) Sign(msg []byte) (*Signature, error) {
	if k.seed == [32]byte{} {
		return nil, errors.New("private key does not have seed, key must be created with GenerateKeys")
	}

	pub := k.Public().key.Bytes()
	pk := ed25519.PrivateKey(append(k.seed[:], pub...))
	return &Signature{
		s: ed25519.Sign(pk, msg),
	}, nil
}

// Verify verifies that the message was signed with the given signature and key.
func (k *PublicKey) Verify(msg []byte, sig *Signature) bool {
	pk := ed25519.PublicKey(k.key.Bytes())
	return ed25519.Verify(pk, msg, sig.s)
}
