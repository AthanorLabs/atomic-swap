package dleq

import (
	"testing"

	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestGoDLEq(t *testing.T) {
	proof, err := (&GoDLEq{}).Prove()
	require.NoError(t, err)

	res, err := (&GoDLEq{}).Verify(proof)
	require.NoError(t, err)

	cpk := res.secp256k1Pub.Compress()
	_, err = ethcrypto.DecompressPubkey(cpk[:])
	require.NoError(t, err)

	sk, err := mcrypto.NewPrivateSpendKey(proof.secret[:])
	require.NoError(t, err)
	ed25519Pub := sk.Public().Bytes()
	require.Equal(t, res.ed25519Pub.Bytes(), ed25519Pub)
}
