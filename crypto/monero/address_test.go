package mcrypto

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/athanorlabs/atomic-swap/common"

	"github.com/stretchr/testify/require"
)

func TestValidateAddress(t *testing.T) {
	kp, err := GenerateKeys()
	require.NoError(t, err)
	addr := kp.Address(common.Mainnet)
	err = ValidateAddress(string(addr), common.Mainnet)
	require.NoError(t, err)
	err = ValidateAddress(string(addr), common.Stagenet)
	require.Equal(t, err, errInvalidPrefixGotMainnet)

	addr = kp.Address(common.Stagenet)
	err = ValidateAddress(string(addr), common.Stagenet)
	require.NoError(t, err)
	err = ValidateAddress(string(addr), common.Mainnet)
	require.Equal(t, err, errInvalidPrefixGotStagenet)

	err = ValidateAddress("fake", common.Mainnet)
	require.True(t, errors.Is(err, errInvalidAddressLength))
}

func TestValidateAddress_loop(t *testing.T) {
	// Tests our address encoding/decoding with randomised data
	for i := 0; i < 1000; i++ {
		kp, err := GenerateKeys() // create random key
		require.NoError(t, err)
		addrHex := hex.EncodeToString(kp.AddressBytes(common.Mainnet))
		addr := kp.Address(common.Mainnet)                  // generates a base58 encoded address
		err = ValidateAddress(string(addr), common.Mainnet) // decodes base58 as part of validation
		require.NoError(t, err, addrHex)                    // save this hex address if the test fails!
	}
}
