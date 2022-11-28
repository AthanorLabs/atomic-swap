package dleq

import (
	csecp256k1 "github.com/athanorlabs/atomic-swap/crypto/secp256k1"

	dleq "github.com/noot/go-dleq"
	"github.com/noot/go-dleq/ed25519"
	"github.com/noot/go-dleq/secp256k1"

	dsecp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

type GoDLEq struct{}

var (
	curveA = secp256k1.NewCurve()
	curveB = ed25519.NewCurve()
)

func (d *GoDLEq) Prove() (*Proof, error) {
	x, err := dleq.GenerateSecretForCurves(curveA, curveB)
	if err != nil {
		panic(err)
	}

	proof, err := dleq.NewProof(curveA, curveB, x)
	if err != nil {
		panic(err)
	}

	err = proof.Verify(curveA, curveB)
	if err != nil {
		panic(err)
	}

	return &Proof{
		proof:  proof.Serialize(),
		secret: x,
	}, nil
}

func (d *GoDLEq) Verify(p *Proof) (*VerifyResult, error) {
	dleqProof := new(dleq.Proof)
	err := dleqProof.Deserialize(curveA, curveB, p.proof)
	if err != nil {
		return nil, err
	}

	err = dleqProof.Verify(curveA, curveB)
	if err != nil {
		return nil, err
	}

	secpPub, err := dsecp256k1.ParsePubKey(dleqProof.CommitmentA.Encode())
	if err != nil {
		return nil, err
	}

	secp256k1Pub := csecp256k1.NewPublicKeyFromBigInt(secpPub.X(), secpPub.Y())

	var ed25519Pub [32]byte
	copy(ed25519Pub[:], dleqProof.CommitmentB.Encode())

	return &VerifyResult{
		secp256k1Pub: secp256k1Pub,
		ed25519Pub:   ed25519Pub,
	}, nil
}
