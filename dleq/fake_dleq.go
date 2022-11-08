//go:build !cgodleq

package dleq

import (
	"crypto/rand"

	"github.com/athanorlabs/atomic-swap/crypto/secp256k1"

	ed25519 "filippo.io/edwards25519"
	dsecp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// DefaultDLEq is FakeDLEq
type DefaultDLEq = FakeDLEq

// FakeDLEq generates a secret scalar that has a point on both curves,
// but doesn't actually prove it.
type FakeDLEq struct{}

// Prove returns a *Proof with a secret key, but no proof.
func (d *FakeDLEq) Prove() (*Proof, error) {
	const (
		ed25519BitSize   = 252
		secp256k1BitSize = 255
	)

	bits := min(ed25519BitSize, secp256k1BitSize)

	// generate secret
	s, err := generateRandomBits(bits)
	if err != nil {
		return nil, err
	}

	var secret [32]byte
	copy(secret[:], s)

	return &Proof{
		secret: secret,
	}, nil
}

// Verify returns the public keys corresponding to the secret key.
// It only fails if it's unable to generate the public keys.
func (d *FakeDLEq) Verify(proof *Proof) (*VerifyResult, error) {
	// generate secp256k1 public key
	scalar := &dsecp256k1.ModNScalar{}
	scalar.PutBytes(&proof.secret)
	sk := dsecp256k1.NewPrivateKey(scalar)
	pk := sk.PubKey()
	secp256k1Pub := secp256k1.NewPublicKeyFromBigInt(pk.X(), pk.Y())

	// generate ed25519 public key
	ed25519Sk, err := ed25519.NewScalar().SetCanonicalBytes(proof.secret[:])
	if err != nil {
		return nil, err
	}

	ed25519Pk := ed25519.NewIdentityPoint().ScalarBaseMult(ed25519Sk)
	var ed25519Pub [32]byte
	copy(ed25519Pub[:], ed25519Pk.Bytes())

	return &VerifyResult{
		secp256k1Pub: secp256k1Pub,
		ed25519Pub:   ed25519Pub,
	}, nil
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}

	return b
}

// generateRandomBits generates up to 256 random bits.
func generateRandomBits(bits uint64) ([]byte, error) {
	x := make([]byte, 32)
	_, err := rand.Read(x)
	if err != nil {
		return nil, err
	}

	toClear := 256 - bits
	x[31] &= 0xff >> toClear
	return x, nil
}
