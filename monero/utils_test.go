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

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_ = daemon.callGenerateBlocks(addr.Address, 181)
		wg.Done()
	}()

	_, err = WaitForBlocks(c, 1)
	require.NoError(t, err)
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
