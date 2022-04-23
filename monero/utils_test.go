package monero

import (
	"testing"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"

	"github.com/stretchr/testify/require"
)

func TestWaitForBlocks(t *testing.T) {
	c := NewClient(common.DefaultBobMoneroEndpoint)
	daemon := NewClient(common.DefaultMoneroDaemonEndpoint)

	addr, err := c.callGetAddress(0)
	require.NoError(t, err)

	go func() {
		_ = daemon.callGenerateBlocks(addr.Address, 181)
	}()

	_, err = WaitForBlocks(c)
	require.NoError(t, err)
}

func TestCreateMoneroWallet(t *testing.T) {
	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	c := NewClient(common.DefaultBobMoneroEndpoint)
	addr, err := CreateMoneroWallet("create-wallet-test", common.Development, c, kp)
	require.NoError(t, err)
	require.Equal(t, kp.Address(common.Development), addr)
}
