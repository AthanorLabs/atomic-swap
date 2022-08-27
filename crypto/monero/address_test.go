package mcrypto

import (
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
