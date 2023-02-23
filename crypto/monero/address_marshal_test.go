package mcrypto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
)

func TestAddress_MarshalText_roundTrip(t *testing.T) {
	keys, err := GenerateKeys()
	require.NoError(t, err)
	addr := keys.PublicKeyPair().Address(common.Development)

	type MyStruct struct {
		XMRAddress *Address `json:"xmrAddress"`
	}

	s1 := &MyStruct{XMRAddress: addr}
	data, err := json.Marshal(s1)
	require.NoError(t, err)

	s2 := new(MyStruct)
	err = json.Unmarshal(data, s2)
	require.NoError(t, err)

	require.Equal(t, s1.XMRAddress, s2.XMRAddress)
}

func TestAddress_UnmarshalText(t *testing.T) {
	for _, test := range addressEncodingTests {
		address := new(Address)
		err := address.UnmarshalText([]byte(test.address))
		require.NoError(t, err)
		require.Equal(t, test.network, address.Network())
		require.Equal(t, test.addressType, address.Type())
	}
}
