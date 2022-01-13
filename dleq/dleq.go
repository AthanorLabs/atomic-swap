package dleq

import (
	"io/ioutil"
	"os"
	"os/exec"
)

type DLEQ interface {
	Prove() (*Proof, error)
	Verify(*Proof) error
}

type Proof struct {
	secret [32]byte
	proof  []byte
}

var (
	dleqGenBinPath    = "../farcaster-dleq/target/release/dleq-gen"
	dleqVerifyBinPath = "../farcaster-dleq/target/release/dleq-verify"
	defaultProofPath  = "../dleq_proof"
)

type FarcasterDLEq struct{}

func (d *FarcasterDLEq) Prove() (*Proof, error) {
	cmd := exec.Command(dleqGenBinPath, defaultProofPath)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	secret, err := ioutil.ReadFile(defaultProofPath + ".key")
	if err != nil {
		return nil, err
	}

	var sc [32]byte
	copy(sc[:], secret)

	proof, err := ioutil.ReadFile(defaultProofPath)
	if err != nil {
		return nil, err
	}

	return &Proof{
		secret: sc,
		proof:  proof,
	}, nil
}

func (d *FarcasterDLEq) Verify(p *Proof) error {
	if err := ioutil.WriteFile(defaultProofPath, p.proof, os.ModePerm); err != nil {
		return err
	}

	cmd := exec.Command(dleqVerifyBinPath, defaultProofPath)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
