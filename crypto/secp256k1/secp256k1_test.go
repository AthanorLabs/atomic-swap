package secp256k1

import (
	"encoding/json"
	"fmt"
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

func TestSecp256k1_JSON(t *testing.T) {
	//nolint:lll
	const jsonStr = `{
		"pubKey": "0xde8df52ed48d2c3320c03344a3fe859d61015e5f8d45b0df9aaa8d056c784e7e55a61a53630ee016e0bc8ac21d6ae4cd92e0ef91e74281d9410167b982764a8e"
	}`

	type someStruct struct {
		PubKey PublicKey `json:"pubKey"`
	}

	// Tests UnmarshalText method
	s := new(someStruct)
	err := json.Unmarshal([]byte(jsonStr), s)
	require.NoError(t, err)
	require.Contains(t, jsonStr, fmt.Sprintf("%q", s.PubKey.String()))

	// Tests MarshallText method
	jsonData, err := json.Marshal(s)
	require.NoError(t, err)
	require.JSONEq(t, jsonStr, string(jsonData))
}
