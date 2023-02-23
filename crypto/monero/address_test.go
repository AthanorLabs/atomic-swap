package mcrypto

import (
	"testing"

	"github.com/athanorlabs/atomic-swap/common"

	"github.com/stretchr/testify/require"
)

func TestNewAddress(t *testing.T) {
	const addrStr = "42ey1afDFnn4886T7196doS9GPMzexD9gXpsZJDwVjeRVdFCSoHnv7KPbBeGpzJBzHRCAs9UxqeoyFQMYbqSWYTfJJQAWDm"
	addr, err := NewAddress(addrStr, common.Mainnet)
	require.NoError(t, err)
	require.Equal(t, addrStr, addr.String())
}

func TestNewAddress_fail(t *testing.T) {
	_, err := NewAddress("fake", common.Mainnet)
	require.ErrorIs(t, err, errInvalidAddressLength)
}

func TestValidateAddress(t *testing.T) {
	kp, err := GenerateKeys()
	require.NoError(t, err)
	pubKeys := kp.PublicKeyPair()

	// mainnet address checks
	addr := pubKeys.Address(common.Mainnet)
	require.NoError(t, addr.ValidateEnv(common.Mainnet))
	require.ErrorIs(t, addr.ValidateEnv(common.Stagenet), errInvalidPrefixGotMainnet)

	// stagenet address checks
	addr = pubKeys.Address(common.Stagenet)
	require.NoError(t, addr.ValidateEnv(common.Stagenet))
	require.ErrorIs(t, addr.ValidateEnv(common.Mainnet), errInvalidPrefixGotStagenet)

	// testnet address check
	const testnetAddress = "9ujeXrjzf7bfeK3KZdCqnYaMwZVFuXemPU8Ubw335rj2FN1CdMiWNyFV3ksEfMFvRp9L9qum5UxkP5rN9aLcPxbH1au4WAB" //nolint:lll
	require.NoError(t, addr.UnmarshalText([]byte(testnetAddress)))
	require.ErrorIs(t, addr.ValidateEnv(common.Mainnet), errInvalidPrefixGotTestnet)

	// uninitialized address validation
	addr = new(Address) // empty
	require.ErrorIs(t, addr.ValidateEnv(common.Development), errAddressNotInitialized)
}

func TestValidateAddress_loop(t *testing.T) {
	// Tests our address encoding/decoding with randomised data
	for i := 0; i < 1000; i++ {
		kp, err := GenerateKeys() // create random key
		require.NoError(t, err)
		// Generate the address, convert it to its base58 string form,
		// then convert the base58 form back into a new address, then
		// verify that the bytes of the 2 addresses are identical.
		addr1 := kp.PublicKeyPair().Address(common.Mainnet)
		addr2, err := NewAddress(addr1.String(), common.Mainnet)
		require.NoError(t, err)
		require.Equal(t, addr1.String(), addr2.String())
	}
}

func TestAddress_Equal(t *testing.T) {
	kp, err := GenerateKeys() // create random key
	require.NoError(t, err)
	pubKeys := kp.PublicKeyPair()

	addr1 := pubKeys.Address(common.Mainnet)
	addr2 := pubKeys.Address(common.Mainnet)
	addr3 := pubKeys.Address(common.Stagenet)

	require.False(t, addr1.Equal(nil))
	require.True(t, addr1.Equal(addr1)) // identity

	require.False(t, addr1 == addr2)    // the pointers are unique,
	require.True(t, addr1.Equal(addr2)) // but the values are the same

	require.False(t, addr1.Equal(addr3)) // same keys, but different network
}
