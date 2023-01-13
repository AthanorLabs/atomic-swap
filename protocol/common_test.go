package protocol

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeysAndProof(t *testing.T) {
	kp, err := GenerateKeysAndProof()
	require.NoError(t, err)

	res, err := VerifyKeysAndProof(
		hex.EncodeToString(kp.DLEqProof.Proof()),
		kp.Secp256k1PublicKey.String(),
		kp.PublicKeyPair.SpendKey().Hex(),
	)
	require.NoError(t, err)
	require.Equal(t, kp.Secp256k1PublicKey.String(), res.Secp256k1PublicKey.String())
	require.Equal(t, kp.PublicKeyPair.SpendKey().Hex(), res.Ed25519PublicKey.Hex())
}
