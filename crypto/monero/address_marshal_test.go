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

	require.True(t, s1.XMRAddress.Equal(s2.XMRAddress))
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

func TestAddress_UnmarshalText_fail(t *testing.T) {
	address := new(Address) // non-initialized address should not marshal
	_, err := address.MarshalText()
	require.ErrorIs(t, err, errChecksumMismatch)
}

func TestAddress_UnmarshalText_badChecksum(t *testing.T) {
	keys, err := GenerateKeys()
	require.NoError(t, err)

	// Generate a good address, then change the checksum to create
	// a new address with a bad checksum
	address := keys.PublicKeyPair().Address(common.Development)
	address.decoded[addressBytesLen-1]++ // overflow fine, 255 goes to 0
	badChecksumAddr := address.String()

	err = address.UnmarshalText([]byte(badChecksumAddr))
	require.ErrorIs(t, err, errChecksumMismatch)
}

func TestAddress_UnmarshalText_badNetworkPrefix(t *testing.T) {
	keys, err := GenerateKeys()
	require.NoError(t, err)

	// Generate a good address, then change the network prefix and adjust the
	// checksum to get an address that is otherwise good, except for the prefix.
	address := keys.PublicKeyPair().Address(common.Development)
	address.decoded[0] = 255
	checksum := getChecksum(address.decoded[0:65])
	copy(address.decoded[65:69], checksum[:])

	badPrefixAddr := address.String()

	err = address.UnmarshalText([]byte(badPrefixAddr))
	require.ErrorContains(t, err, "monero address has unknown network prefix 255")
}

func TestAddress_UnmarshalText_integratedAddress(t *testing.T) {
	const integratedAddress = "4BxSHvcgTwu25WooY4BVmgdcKwZu5EksVZSZkDd6ooxSVVqQ4ubxXkhLF6hEqtw96i9cf3cVfLw8UWe95bdDKfRQeYtPwLm1Jiw7AKt2LY" //nolint:lll
	address := new(Address)
	err := address.UnmarshalText([]byte(integratedAddress))
	require.ErrorIs(t, err, errInvalidAddressLength)
	require.ErrorContains(t, err, "integrated addresses not supported")
}
