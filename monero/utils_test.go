package monero

import (
	"context"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
)

func TestWaitForBlocks(t *testing.T) {
	c := CreateWalletClient(t)

	heightBefore, err := c.GetHeight()
	require.NoError(t, err)

	heightAfter, err := WaitForBlocks(context.Background(), c, 2)
	require.NoError(t, err)
	require.GreaterOrEqual(t, heightAfter-heightBefore, uint64(2))
}

func TestCreateMoneroWallet(t *testing.T) {
	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	c, err := NewWalletClient(&WalletClientConf{
		Env:                 common.Development,
		WalletFilePath:      path.Join(t.TempDir(), "wallet", "not-used"),
		MoneroWalletRPCPath: moneroWalletRPCPath,
	})
	require.NoError(t, err)

	height, err := c.GetHeight()
	require.NoError(t, err)

	addr, err := CreateWallet("create-wallet-test", common.Development, c, kp, height)
	require.NoError(t, err)
	require.Equal(t, kp.Address(common.Development), addr)
}
