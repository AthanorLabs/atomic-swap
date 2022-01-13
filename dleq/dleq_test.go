package dleq

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFarcasterDLEqProof(t *testing.T) {
	f := &FarcasterDLEq{}
	proof, err := f.Prove()
	require.NoError(t, err)
	err = f.Verify(proof)
	require.NoError(t, err)
}

func TestFarcasterDLEqProof_invalid(t *testing.T) {
	f := &FarcasterDLEq{}
	proof, err := f.Prove()
	require.NoError(t, err)
	proof.proof[0] = 0xff
	err = f.Verify(proof)
	require.Error(t, err)
}
