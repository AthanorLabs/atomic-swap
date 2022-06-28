package xmrtaker

import (
	"testing"

	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/tests"

	"github.com/stretchr/testify/require"
)

func TestGetAddress(t *testing.T) {
	c := monero.NewClient(tests.CreateWalletRPCService(t))
	addr, err := getAddress(c, "", "")
	require.NoError(t, err)

	addr2, err := getAddress(c, swapDepositWallet, "")
	require.NoError(t, err)
	require.Equal(t, addr, addr2)
}
