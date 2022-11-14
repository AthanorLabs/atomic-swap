package secp256k1

import (
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/stretchr/testify/require"
)

func TestSecp256k1_Compress(t *testing.T) {
	eckey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)

	key := NewPublicKeyFromBigInt(eckey.X, eckey.Y)
	require.NoError(t, err)

	ckey := key.Compress()
	x, y := secp256k1.DecompressPubkey(ckey[:])
	require.Equal(t, eckey.X, x)
	require.Equal(t, eckey.Y, y)
}
