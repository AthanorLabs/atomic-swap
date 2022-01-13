package dleq

import (
	"crypto/ed25519"
	"crypto/ecdsa"

	"filippo.io/edwards25519"
	//"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

type DLEQ interface {
	Prove() (*Proof, *ProofSecret, error)
	Verify(*Proof) error
}

type Secret [32]byte

type Secp256k1Signature [65]byte

type Ed25519Signature [ed25519.SignatureSize]byte

type PublicKeyTuple struct {
	ed25519Key *edwards25519.Point
	secp256k1Key *ecdsa.PublicKey
}

type SignatureTuple struct {
	ed25519Signature *Ed25519Signature
	secp256k1Key *Secp256k1Signature
}

type Proof struct {
	baseCommitments []*PublicKeyTuple
	firstChallenges [32]byte
	sValues [2]*PublicKeyTuple
	signatures *SignatureTuple
}

type ProofSecret struct {
	secret [32]byte
	ed25519Key *edwards25519.Scalar
	secp256k1Key *ecdsa.PrivateKey
}

var (
	dleqGenBinPath = "../farcaster-dleq/target/release/dleq-gen"
	dleqVerifyBinPath = "../farcaster-dleq/target/release/dleq-verify"

type FarcasterDLEq struct {}

func (d *FarcasterDLEq) Prove() (*Proof, *ProofSecret, error) {

	return nil, nil, nil
}

func (d *FarcasterDLEq) Verify(*Proof) error {
	return nil
}