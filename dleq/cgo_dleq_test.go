package dleq

import (
	"testing"

	mcrypto "github.com/noot/atomic-swap/crypto/monero"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestCGODLEq(t *testing.T) {
	proof, err := (&CGODLEq{}).Prove()
	require.NoError(t, err)

	res, err := (&CGODLEq{}).Verify(proof)
	require.NoError(t, err)

	cpk := res.secp256k1Pub.Compress()
	_, err = ethcrypto.DecompressPubkey(cpk[:])
	require.NoError(t, err)

	sk, err := mcrypto.NewPrivateSpendKey(proof.secret[:])
	require.NoError(t, err)
	ed25519Pub := sk.Public().Bytes()
	require.Equal(t, res.ed25519Pub[:], ed25519Pub)
}

func TestCGODLEq_Invalid(t *testing.T) {
	proof, err := (&CGODLEq{}).Prove()
	require.NoError(t, err)
	proof.proof[0] = 0

	_, err = (&CGODLEq{}).Verify(proof)
	require.Error(t, err)
}
