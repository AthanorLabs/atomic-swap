// Package protocol has no obvious theme and doesn't have much to do with a protocol. The
// EthereumAssetAmount seems like it belongs with the coins types in the common package.
// KeysAndProof seems like it belongs in some crypto package. ConvertContractSwapToMsg
// seems like it belongs in the net/message package.
package protocol

import (
	"encoding/hex"
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

// VerifyKeysAndProof verifies the given DLEq proof and asserts that the resulting secp256k1 key corresponds
// to the given key.
func VerifyKeysAndProof(proofStr, secp256k1PubString string) (*secp256k1.PublicKey, error) {
	pb, err := hex.DecodeString(proofStr)
	if err != nil {
		return nil, err
	}

	d := &dleq.DefaultDLEq{}
	proof := dleq.NewProofWithoutSecret(pb)
	res, err := d.Verify(proof)
	if err != nil {
		return nil, err
	}

	if res.Secp256k1PublicKey().String() != secp256k1PubString {
		return nil, errInvalidSecp256k1Key
	}

	secp256k1Pub, err := secp256k1.NewPublicKeyFromHex(secp256k1PubString)
	if err != nil {
		return nil, err
	}

	return secp256k1Pub, nil
}
