package protocol

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeysAndProof(t *testing.T) {
	kp, err := GenerateKeysAndProof()
	require.NoError(t, err)

	pk, err := VerifyKeysAndProof(hex.EncodeToString(kp.DLEqProof.Proof()), kp.Secp256k1PublicKey.String())
	require.NoError(t, err)
	require.Equal(t, kp.Secp256k1PublicKey.String(), pk.String())
}
