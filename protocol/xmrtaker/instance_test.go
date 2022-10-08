package xmrtaker

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/monero"
)

func TestGetAddress(t *testing.T) {
	c := monero.CreateWalletClient(t)
	addr, err := getAddress(c, "", "")
	require.NoError(t, err)

	addr2, err := getAddress(c, swapDepositWallet, "")
	require.NoError(t, err)
	require.Equal(t, addr, addr2)
}
