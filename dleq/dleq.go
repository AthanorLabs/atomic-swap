package dleq

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/crypto/secp256k1"
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

// NewProofWithSecret returns a new Proof with the given secret.
// Note that the returned proof actually lacks the `proof` field.
func NewProofWithSecret(s [32]byte) *Proof {
	return &Proof{
		secret: s,
	}
}

// Secret returns the proof's 32-byte secret
func (p *Proof) Secret() [32]byte {
	return p.secret
}

// Proof returns the encoded DLEq proof
func (p *Proof) Proof() []byte {
	return p.proof
}

// VerifyResult contains the public keys resulting from verifying a DLEq proof
type VerifyResult struct {
	ed25519Pub   [32]byte
	secp256k1Pub *secp256k1.PublicKey
}

// Secp256k1PublicKey returns the secp256k1 public key associated with the DLEq verification
func (r *VerifyResult) Secp256k1PublicKey() *secp256k1.PublicKey {
	return r.secp256k1Pub
}

var (
	dleqGenBinPath    = getFarcasterDLEqBinaryPath() + "dleq-gen"
	dleqVerifyBinPath = getFarcasterDLEqBinaryPath() + "dleq-verify"
	defaultProofPath  = "/tmp/dleq_proof"
)

// TODO: this is kinda sus, make it actually find the bin better. maybe env vars?
func getFarcasterDLEqBinaryPath() string {
	bin := "./farcaster-dleq/target/release/dleq-gen"
	_, err := os.Stat(bin)
	if !errors.Is(err, os.ErrNotExist) {
		return "./farcaster-dleq/target/release/"
	}

	bin = "../farcaster-dleq/target/release/dleq-gen"
	_, err = os.Stat(bin)
	if !errors.Is(err, os.ErrNotExist) {
		return "../farcaster-dleq/target/release/"
	}

	return "../../farcaster-dleq/target/release/"
}

// FarcasterDLEq is a wrapper around the binaries in farcaster-dleq
type FarcasterDLEq struct{}

// Prove generates a new DLEq proof
func (d *FarcasterDLEq) Prove() (*Proof, error) {
	t := time.Now().Format(common.TimeFmtNSecs)
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
func (d *FarcasterDLEq) Verify(p *Proof) (*VerifyResult, error) {
	t := time.Now().Format(common.TimeFmtNSecs)
	path := fmt.Sprintf("%s-verify-%s", defaultProofPath, t)

	if err := ioutil.WriteFile(path, p.proof, os.ModePerm); err != nil {
		return nil, err
	}

	cmd := exec.Command(dleqVerifyBinPath, path)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// slice off \n at the end of string
	out := strings.Split(string(output[:len(output)-1]), " ")
	if len(out) != 3 {
		return nil, errors.New("invalid output from dleq-verify")
	}

	ed25519Pub, err := hex.DecodeString(out[0])
	if err != nil {
		return nil, err
	}

	secp256k1X, err := hex.DecodeString(out[1])
	if err != nil {
		return nil, err
	}

	secp256k1Y, err := hex.DecodeString(out[2])
	if err != nil {
		return nil, err
	}

	res := &VerifyResult{}
	copy(res.ed25519Pub[:], ed25519Pub)

	var x, y [32]byte
	copy(x[:], secp256k1X)
	copy(y[:], secp256k1Y)
	res.secp256k1Pub = secp256k1.NewPublicKey(x, y)

	return res, nil
}
