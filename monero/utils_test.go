package monero

import (
	"sync"
	"testing"

	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/tests"

	"github.com/stretchr/testify/require"
)

func TestWaitForBlocks(t *testing.T) {
	c := NewWalletClient(tests.CreateWalletRPCService(t))
	require.NoError(t, c.CreateWallet("wallet", ""))
	daemon := NewDaemonClient(common.DefaultMoneroDaemonEndpoint)

	addr, err := c.GetAddress(0)
	require.NoError(t, err)

	heightBefore, err := c.GetHeight()
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		errGen := daemon.GenerateBlocks(addr.Address, 181)
		require.NoError(t, errGen)
		wg.Done()
	}()
	heightAfter, err := WaitForBlocks(c, 1)
	require.NoError(t, err)
	require.GreaterOrEqual(t, heightAfter-heightBefore, uint64(1))
	wg.Wait()
}

func TestCreateMoneroWallet(t *testing.T) {
	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	c := NewWalletClient(tests.CreateWalletRPCService(t))
	addr, err := CreateWallet("create-wallet-test", common.Development, c, kp)
	require.NoError(t, err)
	require.Equal(t, kp.Address(common.Development), addr)
}
