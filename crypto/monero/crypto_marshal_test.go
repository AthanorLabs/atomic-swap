package mcrypto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/vjson"
)

func TestKey_Marshal_success(t *testing.T) {
	type SomeStruct struct {
		PrivSpendKey   *PrivateSpendKey `json:"privSpendKey" validate:"required"`
		PrivateViewKey *PrivateViewKey  `json:"privViewKey" validate:"required"`
		PublicSpendKey *PublicKey       `json:"pubSpendKey" validate:"required"`
	}

	const (
		expectedPrivSpendKey = "ab0000000000000000000000000000000000000000000000000000000000cd00"
		expectedPrivViewKey  = "cd0000000000000000000000000000000000000000000000000000000000ef00"
		expectedPubKey       = "5866666666666666666666666666666666666666666666666666666666666666" // generator point
		expectJSON           = `{
			"privSpendKey": "0xab0000000000000000000000000000000000000000000000000000000000cd00",
			"privViewKey":  "0xcd0000000000000000000000000000000000000000000000000000000000ef00",
			"pubSpendKey":  "0x5866666666666666666666666666666666666666666666666666666666666666"
		}`
	)

	spendKey := new(PrivateSpendKey)
	err := spendKey.UnmarshalText([]byte(expectedPrivSpendKey))
	require.NoError(t, err)

	viewKey := new(PrivateViewKey)
	err = viewKey.UnmarshalText([]byte(expectedPrivViewKey))
	require.NoError(t, err)

	pubKey := new(PublicKey)
	err = pubKey.UnmarshalText([]byte(expectedPubKey))
	require.NoError(t, err)

	data, err := vjson.MarshalStruct(&SomeStruct{
		PrivSpendKey:   spendKey,
		PrivateViewKey: viewKey,
		PublicSpendKey: pubKey,
	})
	require.NoError(t, err)
	require.JSONEq(t, expectJSON, string(data))

	s := new(SomeStruct)
	err = vjson.UnmarshalStruct(data, s)
	require.NoError(t, err)

	require.Equal(t, expectedPrivSpendKey, s.PrivSpendKey.Hex())
	require.Equal(t, expectedPrivViewKey, s.PrivateViewKey.Hex())
	require.Equal(t, expectedPubKey, s.PublicSpendKey.Hex())
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

func TestPrivateKeyPair_Marshal(t *testing.T) {
	kp, err := GenerateKeys()
	require.NoError(t, err)

	// serialize, deserialize, and make sure the result is the same as the original
	jsonData, err := json.Marshal(kp)
	require.NoError(t, err)
	kp2 := new(PrivateKeyPair)
	err = vjson.UnmarshalStruct(jsonData, kp2)
	require.NoError(t, err)

	assert.Equal(t, kp.SpendKey().Hex(), kp2.SpendKey().Hex())
	assert.Equal(t, kp.ViewKey().Hex(), kp2.ViewKey().Hex())
}

func TestPublicKeyPair_Marshal(t *testing.T) {
	kp, err := GenerateKeys()
	require.NoError(t, err)
	pubKP1 := kp.PublicKeyPair()

	// serialize, deserialize, and make sure the result is the same as the original
	jsonData, err := vjson.MarshalStruct(pubKP1)
	require.NoError(t, err)
	pubKP2 := new(PublicKeyPair)
	err = vjson.UnmarshalStruct(jsonData, pubKP2)
	require.NoError(t, err)

	assert.Equal(t, pubKP1.SpendKey().Hex(), pubKP2.SpendKey().Hex())
	assert.Equal(t, pubKP1.ViewKey().Hex(), pubKP2.ViewKey().Hex())
}
