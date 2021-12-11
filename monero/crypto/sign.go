package crypto

import (
	"crypto/ed25519"
	"errors"
)

type Signature struct {
	s []byte
}

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

func (k *PublicKey) Verify(msg []byte, sig *Signature) bool {
	pk := ed25519.PublicKey(k.key.Bytes())
	return ed25519.Verify(pk, msg, sig.s)
}
