package mcrypto

import (
	"encoding/json"
	"testing"

	ed25519 "filippo.io/edwards25519"
	"github.com/stretchr/testify/require"
)

func TestP_Marshal_success(t *testing.T) {
	// Using a struct and json.Marshal (instead of directly invoking psk.MarshalText()) so
	// the reader can easily see what the generated JSON looks like.
	type SomeStruct struct {
		PrivSpendKey   *PrivateSpendKey `json:"privSpendKey"`
		PrivateViewKey *PrivateViewKey  `json:"privViewKey"`
		PublicSpendKey *PublicKey       `json:"pubSpendKey"`
	}
	ed25519.NewIdentityPoint().Bytes()
	const (
		expectedPrivSpendKey = "ab0000000000000000000000000000000000000000000000000000000000cd00"
		expectedPrivViewKey  = "cd0000000000000000000000000000000000000000000000000000000000ef00"
		expectedPubKey       = "5866666666666666666666666666666666666666666666666666666666666666" // generator point
		expectJSON           = `{
			"privSpendKey": "ab0000000000000000000000000000000000000000000000000000000000cd00",
			"privViewKey":  "cd0000000000000000000000000000000000000000000000000000000000ef00",
			"pubSpendKey":  "5866666666666666666666666666666666666666666666666666666666666666"
		}`
	)
	spendKey, err := NewPrivateSpendKeyFromHex(expectedPrivSpendKey)
	require.NoError(t, err)
	viewKey, err := NewPrivateViewKeyFromHex(expectedPrivViewKey)
	require.NoError(t, err)
	pubKey, err := NewPublicKeyFromHex(expectedPubKey)
	require.NoError(t, err)
	s1 := &SomeStruct{
		PrivSpendKey:   spendKey,
		PrivateViewKey: viewKey,
		PublicSpendKey: pubKey,
	}
	data, err := json.Marshal(s1)
	require.NoError(t, err)
	require.JSONEq(t, expectJSON, string(data))

	s2 := &SomeStruct{
		PrivSpendKey:   &PrivateSpendKey{},
		PrivateViewKey: &PrivateViewKey{ed25519.NewScalar()},
		PublicSpendKey: &PublicKey{ed25519.NewIdentityPoint()},
	}
	err = json.Unmarshal(data, s2)
	require.NoError(t, err)
	require.Equal(t, expectedPrivSpendKey, s2.PrivSpendKey.Hex())
	require.Equal(t, expectedPrivViewKey, s2.PrivateViewKey.Hex())
	require.Equal(t, expectedPubKey, s2.PublicSpendKey.Hex())
}

func TestPrivateSpendKey_MarshalText_uninitialized(t *testing.T) {
	psk := &PrivateSpendKey{} // key inside is nil
	_, err := psk.MarshalText()
	require.ErrorContains(t, err, "uninitialized")
}

func TestPrivateSpendKey_UnmarshalText_nil(t *testing.T) {
	psk := &PrivateSpendKey{}
	err := psk.UnmarshalText([]byte(""))
	require.ErrorContains(t, err, "invalid scalar length")
}
