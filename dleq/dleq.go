package dleq

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Interface ...
type Interface interface {
	Prove() (*Proof, error)
	Verify(*Proof) error
}

// Proof represents a DLEq proof
type Proof struct {
	secret [32]byte
	proof  []byte
}

var (
	dleqGenBinPath    = "../farcaster-dleq/target/release/dleq-gen"
	dleqVerifyBinPath = "../farcaster-dleq/target/release/dleq-verify"
	defaultProofPath  = "../dleq_proof"
)

// FarcasterDLEq is a wrapper around the binaries in farcaster-dleq
type FarcasterDLEq struct{}

// Prove generates a new DLEq proof
func (d *FarcasterDLEq) Prove() (*Proof, error) {
	t := time.Now().Format("2006-Jan-2-15:04:05")
	path := fmt.Sprintf("%s-%s", defaultProofPath, t)

	cmd := exec.Command(dleqGenBinPath, path)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	secret, err := ioutil.ReadFile(filepath.Clean(path + ".key"))
	if err != nil {
		return nil, err
	}

	var sc [32]byte
	copy(sc[:], secret)

	proof, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	return &Proof{
		secret: sc,
		proof:  proof,
	}, nil
}

// Verify verifies a DLEq proof
func (d *FarcasterDLEq) Verify(p *Proof) error {
	t := time.Now().Format("2006-Jan-2-15:04:05")
	path := fmt.Sprintf("%s-verify-%s", defaultProofPath, t)

	if err := ioutil.WriteFile(path, p.proof, os.ModePerm); err != nil {
		return err
	}

	cmd := exec.Command(dleqVerifyBinPath, path)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
