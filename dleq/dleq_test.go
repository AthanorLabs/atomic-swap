package dleq

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	"github.com/stretchr/testify/require"
)

func TestFarcasterDLEqProof(t *testing.T) {
	f := &FarcasterDLEq{}
	proof, err := f.Prove()
	require.NoError(t, err)
	res, err := f.Verify(proof)
	require.NoError(t, err)
	require.NotEqual(t, [32]byte{}, res.ed25519Pub)
	require.NotEqual(t, [32]byte{}, res.secp256k1Pub.X())
	require.NotEqual(t, [32]byte{}, res.secp256k1Pub.Y())
}

func TestFarcasterDLEqProof_invalid(t *testing.T) {
	f := &FarcasterDLEq{}
	proof, err := f.Prove()
	require.NoError(t, err)
	proof.proof[0] = 0xff
	_, err = f.Verify(proof)
	require.Error(t, err)
}

func TestFarcasterDLEqProof_createKeys(t *testing.T) {
	f := &FarcasterDLEq{}
	proof, err := f.Prove()
	require.NoError(t, err)

	sk, err := mcrypto.NewPrivateSpendKey(proof.secret[:])
	require.NoError(t, err)

	res, err := f.Verify(proof)
	require.NoError(t, err)
	require.Equal(t, res.ed25519Pub[:], sk.Public().Bytes())

	curve := secp256k1.S256()

	xb := res.secp256k1Pub.X()
	yb := res.secp256k1Pub.Y()
	x := big.NewInt(0).SetBytes(xb[:])
	y := big.NewInt(0).SetBytes(yb[:])
	ok := curve.IsOnCurve(x, y)
	require.True(t, ok)
}
