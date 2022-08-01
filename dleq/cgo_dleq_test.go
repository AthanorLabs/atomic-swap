package dleq

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	ethsecp256k1 "github.com/ethereum/go-ethereum/crypto/secp256k1"

	"github.com/noot/atomic-swap/common"
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

func TestProofSecretComputesVerifyPubKeys(t *testing.T) {
	// It would be nice to increase the number of iterations, but it's pretty slow even at 128. We
	// previously had an issue when X or Y needed at least one high order padding byte. The chance
	// of that happening is around (1/256+1/256)=1/128, so this loop will see values like that
	// frequently, even if it doesn't happen on every run.
	const iterations = 128
	for i := 0; i < iterations; i++ {
		proof, err := (&CGODLEq{}).Prove()
		require.NoError(t, err)
		res, err := (&CGODLEq{}).Verify(proof)
		require.NoError(t, err)

		// The ETH library needs the secret in big-endian format, while the monero library wants it
		// in little endian format.
		secretLE := proof.secret[:]
		secretBE := common.Reverse(secretLE)

		// Secp256k1 check
		ethCurve := ethsecp256k1.S256()
		xSecret, ySecret := ethCurve.ScalarBaseMult(secretBE)
		pubFromSecret := &ecdsa.PublicKey{Curve: ethCurve, X: xSecret, Y: ySecret}
		xVerify := res.Secp256k1PublicKey().X()
		yVerify := res.Secp256k1PublicKey().Y()
		pubFromVerify := &ecdsa.PublicKey{Curve: ethCurve,
			X: new(big.Int).SetBytes(xVerify[:]), Y: new(big.Int).SetBytes(yVerify[:]),
		}
		require.True(t, pubFromSecret.Equal(pubFromVerify))

		// ED25519 Check
		sk, err := mcrypto.NewPrivateSpendKey(secretLE)
		require.NoError(t, err)
		ed25519Pub := sk.Public().Bytes()
		require.Equal(t, res.ed25519Pub[:], ed25519Pub)
	}
}
