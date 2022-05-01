package dleq

import (
	"errors"

	"github.com/noot/atomic-swap/crypto/secp256k1"
	dleq "github.com/noot/cgo-dleq"

	ethsecp256k1 "github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// CGODLEq is a wrapper around the CGO bindings to dleq-rs
type CGODLEq struct{}

// Prove generates a new DLEq proof
func (d *CGODLEq) Prove() (*Proof, error) {
	proof, pk, err := dleq.Ed25519Secp256k1Prove()
	if err != nil {
		return nil, err
	}

	var secret [32]byte
	copy(secret[:], pk)

	return &Proof{
		secret: secret,
		proof:  []byte(proof),
	}, nil
}

// Verify verifies a DLEq proof
func (d *CGODLEq) Verify(proof *Proof) (*VerifyResult, error) {
	ed25519Pub, secp256k1Pub, err := dleq.Ed25519Secp256k1Verify(proof.proof)
	if err != nil {
		return nil, err
	}

	var edPub [32]byte
	copy(edPub[:], []byte(ed25519Pub))

	x, y := ethsecp256k1.DecompressPubkey([]byte(secp256k1Pub))
	if x == nil {
		return nil, errors.New("failed to decompress secp256k1 public key")
	}

	return &VerifyResult{
		ed25519Pub:   edPub,
		secp256k1Pub: secp256k1.NewPublicKeyFromBigInt(x, y),
	}, nil
}
