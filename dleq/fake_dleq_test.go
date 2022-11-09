package dleq

import (
	"testing"

	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestFakeDLEq(t *testing.T) {
	proof, err := (&FakeDLEq{}).Prove()
	require.NoError(t, err)

	res, err := (&FakeDLEq{}).Verify(proof)
	require.NoError(t, err)

	cpk := res.secp256k1Pub.Compress()
	_, err = ethcrypto.DecompressPubkey(cpk[:])
	require.NoError(t, err)

	sk, err := mcrypto.NewPrivateSpendKey(proof.secret[:])
	require.NoError(t, err)
	ed25519Pub := sk.Public().Bytes()
	require.Equal(t, res.ed25519Pub[:], ed25519Pub)
}
