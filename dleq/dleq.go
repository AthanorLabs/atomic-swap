// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

// Package dleq provides a sub-api built on top of the go-dleq package for our atomic
// swaps. The API allows you to verify that a Monero public spend key on the ed25519 curve
// have the same discrete logarithm (same shared secret) as a public key on the secp256k1
// curve. A ZK DLEq proof is used to prove equivalence of the secret key corresponding to
// public keys on both curves.
package dleq

import (
	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/crypto/secp256k1"
)

// Interface ...
type Interface interface {
	Prove() (*Proof, error)
	Verify(*Proof) (*VerifyResult, error)
}

// Proof represents a DLEq proof
type Proof struct {
	secret [32]byte
	proof  []byte
}

// NewProofWithoutSecret returns a new Proof without a secret from the given proof slice
func NewProofWithoutSecret(p []byte) *Proof {
	return &Proof{
		proof: p,
	}
}

// Secret returns the proof's 32-byte secret
func (p *Proof) Secret() [32]byte {
	var s [32]byte
	copy(s[:], common.Reverse(p.secret[:]))
	return s
}

// Proof returns the encoded DLEq proof
func (p *Proof) Proof() []byte {
	return p.proof
}

// VerifyResult contains the public keys resulting from verifying a DLEq proof
type VerifyResult struct {
	ed25519Pub   *mcrypto.PublicKey
	secp256k1Pub *secp256k1.PublicKey
}

// Secp256k1PublicKey returns the secp256k1 public key associated with the DLEq verification
func (r *VerifyResult) Secp256k1PublicKey() *secp256k1.PublicKey {
	return r.secp256k1Pub
}

// Ed25519PublicKey returns the ed25519 public key associated with the DLEq verification
func (r *VerifyResult) Ed25519PublicKey() *mcrypto.PublicKey {
	return r.ed25519Pub
}
