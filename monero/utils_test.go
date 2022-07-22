package monero

import (
	"sync"
	"testing"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	"github.com/noot/atomic-swap/tests"

	"github.com/stretchr/testify/require"
)

func TestWaitForBlocks(t *testing.T) {
	c := NewClient(tests.CreateWalletRPCService(t))
	require.NoError(t, c.CreateWallet("wallet", ""))
	daemon := NewClient(common.DefaultMoneroDaemonEndpoint)

	addr, err := c.callGetAddress(0)
	require.NoError(t, err)

	heightBefore, err := c.GetHeight()
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		errGen := daemon.callGenerateBlocks(addr.Address, 181)
		require.NoError(t, errGen)
		wg.Done()
	}()
	heightAfter, err := WaitForBlocks(c, 1)
	require.NoError(t, err)
	require.GreaterOrEqual(t, heightAfter-heightBefore, uint(1))
	wg.Wait()
}

func TestCreateMoneroWallet(t *testing.T) {
	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	c := NewClient(tests.CreateWalletRPCService(t))
	addr, err := CreateMoneroWallet("create-wallet-test", common.Development, c, kp)
	require.NoError(t, err)
	require.Equal(t, kp.Address(common.Development), addr)
}
