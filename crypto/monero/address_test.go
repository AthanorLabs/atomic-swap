package mcrypto

import (
	"testing"

	"github.com/athanorlabs/atomic-swap/common"

	"github.com/stretchr/testify/require"
)

func TestValidateAddress(t *testing.T) {
	kp, err := GenerateKeys()
	require.NoError(t, err)
	addr := kp.PublicKeyPair().Address(common.Mainnet)
	require.NoError(t, addr.Validate(common.Mainnet))
	require.ErrorIs(t, addr.Validate(common.Stagenet), errInvalidPrefixGotMainnet)

	addr = kp.PublicKeyPair().Address(common.Stagenet)
	require.NoError(t, addr.Validate(common.Stagenet))
	require.ErrorIs(t, addr.Validate(common.Mainnet), errInvalidPrefixGotStagenet)

	_, err = NewAddress("fake", common.Mainnet)
	require.ErrorIs(t, err, errInvalidAddressLength)
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
