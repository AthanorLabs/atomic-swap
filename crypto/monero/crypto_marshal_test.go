package mcrypto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKey_Marshal_success(t *testing.T) {
	type SomeStruct struct {
		PrivSpendKey   *PrivateSpendKey `json:"privSpendKey"`
		PrivateViewKey *PrivateViewKey  `json:"privViewKey"`
		PublicSpendKey *PublicKey       `json:"pubSpendKey"`
	}

	const (
		expectedPrivSpendKey = "0xab0000000000000000000000000000000000000000000000000000000000cd00"
		expectedPrivViewKey  = "0xcd0000000000000000000000000000000000000000000000000000000000ef00"
		expectedPubKey       = "0x5866666666666666666666666666666666666666666666666666666666666666" // generator point
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

	data, err := json.Marshal(&SomeStruct{
		PrivSpendKey:   spendKey,
		PrivateViewKey: viewKey,
		PublicSpendKey: pubKey,
	})
	require.NoError(t, err)
	require.JSONEq(t, expectJSON, string(data))

	s := new(SomeStruct)
	err = json.Unmarshal(data, s)
	require.NoError(t, err)

	require.Equal(t, expectedPrivSpendKey, s.PrivSpendKey.String())
	require.Equal(t, expectedPrivViewKey, s.PrivateViewKey.String())
	require.Equal(t, expectedPubKey, s.PublicSpendKey.String())
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
	err = json.Unmarshal(jsonData, kp2)
	require.NoError(t, err)

	assert.Equal(t, kp.SpendKey().String(), kp2.SpendKey().String())
	assert.Equal(t, kp.ViewKey().String(), kp2.ViewKey().String())
}

func TestPublicKeyPair_Marshal(t *testing.T) {
	kp, err := GenerateKeys()
	require.NoError(t, err)
	pubKP1 := kp.PublicKeyPair()

	// serialize, deserialize, and make sure the result is the same as the original
	jsonData, err := json.Marshal(pubKP1)
	require.NoError(t, err)
	pubKP2 := new(PublicKeyPair)
	err = json.Unmarshal(jsonData, pubKP2)
	require.NoError(t, err)

	assert.Equal(t, pubKP1.SpendKey().String(), pubKP2.SpendKey().String())
	assert.Equal(t, pubKP1.ViewKey().String(), pubKP2.ViewKey().String())
}
