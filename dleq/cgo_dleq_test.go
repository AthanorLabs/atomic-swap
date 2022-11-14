//go:build !fakedleq

package dleq

import (
	"bytes"
	"crypto/ecdsa"
	"math/big"
	"testing"

	ethsecp256k1 "github.com/ethereum/go-ethereum/crypto/secp256k1"

	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"

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

func TestProofSecretComputesVerifyPubKeys(t *testing.T) {
	// It would be nice to increase the number of iterations, but it's pretty slow even at 128. We
	// previously had an issue when X or Y needed at least one high order padding byte. The chance
	// of that happening is around (1/256+1/256)=1/128, so this loop will see values like that
	// frequently, even if it doesn't happen on every run.
	const iterations = 128

	toBigInt := func(point [32]byte) *big.Int { return new(big.Int).SetBytes(point[:]) }

	for i := 0; i < iterations; i++ {
		proof, err := (&CGODLEq{}).Prove()
		require.NoError(t, err)
		res, err := (&CGODLEq{}).Verify(proof)
		require.NoError(t, err)

		// The ETH library needs the secret in big-endian format, while the monero library wants it
		// in little endian format.
		secretBE := proof.secret[:]
		secretLE := common.Reverse(secretBE)

		// Secp256k1 check
		ethCurve := ethsecp256k1.S256()
		xPub, yPub := ethCurve.ScalarBaseMult(secretLE)
		ethPubFromSecret := &ecdsa.PublicKey{Curve: ethCurve, X: xPub, Y: yPub}
		ethPubFromVerify := &ecdsa.PublicKey{Curve: ethCurve,
			X: toBigInt(res.Secp256k1PublicKey().X()), Y: toBigInt(res.Secp256k1PublicKey().Y()),
		}
		require.True(t, ethPubFromSecret.Equal(ethPubFromVerify))

		// ED25519 Check
		sk, err := mcrypto.NewPrivateSpendKey(secretBE)
		require.NoError(t, err)
		xmrPubFromSecret := sk.Public().Bytes()
		xmrPubFromVerify := res.ed25519Pub[:]
		require.True(t, bytes.Equal(xmrPubFromSecret, xmrPubFromVerify))
	}
}
