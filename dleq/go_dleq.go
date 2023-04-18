// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package dleq

import (
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	csecp256k1 "github.com/athanorlabs/atomic-swap/crypto/secp256k1"

	dleq "github.com/athanorlabs/go-dleq"
	"github.com/athanorlabs/go-dleq/ed25519"
	"github.com/athanorlabs/go-dleq/secp256k1"

	dsecp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// DefaultDLEq is the default DLEq prover.
// Currently, the only implementation is GoDLEq.
type DefaultDLEq = GoDLEq

// GoDLEq is a wrapper around the go-dleq library prover and verifier.
type GoDLEq struct{}

var (
	curveEthereum = secp256k1.NewCurve()
	curveMonero   = ed25519.NewCurve()
)

// Prove generates a secret scalar and a proof that it has a corresponding
// public key on the secp256k1 and ed25519 curves.
func (d *GoDLEq) Prove() (*Proof, error) {
	x, err := dleq.GenerateSecretForCurves(curveEthereum, curveMonero)
	if err != nil {
		return nil, err
	}

	proof, err := dleq.NewProof(curveEthereum, curveMonero, x)
	if err != nil {
		return nil, err
	}

	err = proof.Verify(curveEthereum, curveMonero)
	if err != nil {
		return nil, err
	}

	return &Proof{
		proof:  proof.Serialize(),
		secret: x,
	}, nil
}

// Verify verifies the given proof. It returns the secp256k1
// and ed25519 public keys corresponding to the secret value.
func (d *GoDLEq) Verify(p *Proof) (*VerifyResult, error) {
	dleqProof := new(dleq.Proof)
	err := dleqProof.Deserialize(curveEthereum, curveMonero, p.proof)
	if err != nil {
		return nil, err
	}

	err = dleqProof.Verify(curveEthereum, curveMonero)
	if err != nil {
		return nil, err
	}

	secpPub, err := dsecp256k1.ParsePubKey(dleqProof.CommitmentA.Encode())
	if err != nil {
		return nil, err
	}

	secp256k1Pub := csecp256k1.NewPublicKeyFromBigInt(secpPub.X(), secpPub.Y())

	ed25519Pub, err := mcrypto.NewPublicKeyFromBytes(dleqProof.CommitmentB.Encode())
	if err != nil {
		return nil, err
	}

	return &VerifyResult{
		secp256k1Pub: secp256k1Pub,
		ed25519Pub:   ed25519Pub,
	}, nil
}
