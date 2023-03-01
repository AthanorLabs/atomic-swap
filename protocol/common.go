// Package protocol has functions that are used by both the maker and taker during execution of the swap.
package protocol

import (
	"bytes"
	"fmt"

	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/crypto/secp256k1"
	"github.com/athanorlabs/atomic-swap/dleq"
)

// KeysAndProof contains a DLEq proof, a secp256k1 public key,
// and ed25519 public and private keypairs.
type KeysAndProof struct {
	DLEqProof          *dleq.Proof
	Secp256k1PublicKey *secp256k1.PublicKey
	PrivateKeyPair     *mcrypto.PrivateKeyPair
	PublicKeyPair      *mcrypto.PublicKeyPair
}

// GenerateKeysAndProof generates keys on the secp256k1 and ed25519 curves as well as
// a DLEq proof between the two.
func GenerateKeysAndProof() (*KeysAndProof, error) {
	d := &dleq.DefaultDLEq{}
	proof, err := d.Prove()
	if err != nil {
		return nil, err
	}

	res, err := d.Verify(proof)
	if err != nil {
		return nil, err
	}

	secret := proof.Secret()
	sk, err := mcrypto.NewPrivateSpendKey(common.Reverse(secret[:]))
	if err != nil {
		return nil, fmt.Errorf("failed to create private spend key: %w", err)
	}

	kp, err := sk.AsPrivateKeyPair()
	if err != nil {
		return nil, err
	}

	return &KeysAndProof{
		DLEqProof:          proof,
		Secp256k1PublicKey: res.Secp256k1PublicKey(),
		PrivateKeyPair:     kp,
		PublicKeyPair:      kp.PublicKeyPair(),
	}, nil
}

// VerifyResult is returned from verifying a DLEq proof.
type VerifyResult struct {
	Secp256k1PublicKey *secp256k1.PublicKey
	Ed25519PublicKey   *mcrypto.PublicKey
}

// VerifyKeysAndProof verifies the given DLEq proof and asserts that the resulting secp256k1 key corresponds
// to the given key.
func VerifyKeysAndProof(
	proofData []byte,
	secp256k1Pub *secp256k1.PublicKey,
	ed25519Pub *mcrypto.PublicKey,
) (*VerifyResult, error) {
	d := &dleq.DefaultDLEq{}
	proof := dleq.NewProofWithoutSecret(proofData)
	res, err := d.Verify(proof)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(res.Secp256k1PublicKey().Bytes(), secp256k1Pub.Bytes()) {
		return nil, errInvalidSecp256k1Key
	}

	if !bytes.Equal(res.Ed25519PublicKey().Bytes(), ed25519Pub.Bytes()) {
		return nil, errInvalidEd25519Key
	}

	return &VerifyResult{
		Secp256k1PublicKey: secp256k1Pub,
		Ed25519PublicKey:   ed25519Pub,
	}, nil
}
