//go:build !prod

package monero

//
// This file is only for test support when working with monerod in regtest mode. Use the build
// tag "prod" to prevent symbols in this file from consuming space (or mentioning mining) in
// production binaries.
//

import (
	"context"
	"path"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/MarinX/monerorpc"
	"github.com/MarinX/monerorpc/daemon"
	"github.com/MarinX/monerorpc/wallet"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
)

const (
	// MonerodRegtestEndpoint is the RPC endpoint used by monerod in the dev environment's regtest mode.
	MonerodRegtestEndpoint = "http://127.0.0.1:18081/json_rpc"

	backgroundMineInterval = 500 * time.Millisecond
	// Mastering monero example address (we don't use the background mining block rewards in tests)
	blockRewardAddress = "4BKjy1uVRTPiz4pHyaXXawb82XpzLiowSDd8rEQJGqvN6AD6kWosLQ6VJXW9sghopxXgQSh1RTd54JdvvCRsXiF41xvfeW5"
)

// CreateWalletClientWithWalletDir creates a WalletClient with the given wallet directory.
func CreateWalletClientWithWalletDir(t *testing.T, walletDir string) WalletClient {
	_, filename, _, ok := runtime.Caller(0) // this test file path
	require.True(t, ok)
	packageDir := path.Dir(filename)
	repoBaseDir := path.Dir(packageDir)
	moneroWalletRPCPath := path.Join(repoBaseDir, "monero-bin", "monero-wallet-rpc")

	c, err := NewWalletClient(&WalletClientConf{
		Env:                 common.Development,
		WalletFilePath:      path.Join(walletDir, "test-wallet"),
		MoneroWalletRPCPath: moneroWalletRPCPath,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Close()
	})
	BackgroundMineBlocks(t)

	return c
}

// CreateWalletClient starts a monero-wallet-rpc listening on a random port for tests and
// returns the client interface for using it. Background mining is initiated so created transactions
// get mined into blocks.
func CreateWalletClient(t *testing.T) WalletClient {
	return CreateWalletClientWithWalletDir(t, t.TempDir())
}

// GetBalance is a convenience method for tests that assumes you want the primary
// address, that you want to refresh, and that errors should fail the test.
func GetBalance(t *testing.T, wc WalletClient) *wallet.GetBalanceResponse {
	err := wc.Refresh()
	require.NoError(t, err)
	balance, err := wc.GetBalance(0)
	require.NoError(t, err)
	return balance
}

var mineMu sync.Mutex

// BackgroundMineBlocks starts a background go routine to mine blocks in a monerod instance
// that is in regtest mode. If there is an existing go routine that is already mining from
// a previous call, no new go routine is created.
func BackgroundMineBlocks(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancelFunc()
		wg.Wait()
	})
	go func() {
		defer wg.Done()
		if !mineMu.TryLock() {
			return // If there are multiple clients in a test, only let one of them mine.
		}
		defer mineMu.Unlock()
		for {
			time.Sleep(backgroundMineInterval)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backgroundMineInterval):
				// not cancelled, mine another block below
			}

			daemonCli := monerorpc.New(MonerodRegtestEndpoint, nil).Daemon
			resp, err := daemonCli.GenerateBlocks(&daemon.GenerateBlocksRequest{
				AmountOfBlocks: 1,
				WalletAddress:  blockRewardAddress,
			})
			if err != nil && err.Error() == "Block not accepted" {
				// This probably happens when something else is simultaneously generating
				// blocks, not an error that matters unless it is happening frequently.
				t.Logf("Background mining had non-accepted block")
				continue
			}
			require.NoError(t, err)
			if false { // change to true if debugging and you want to see when new blocks are generated
				t.Logf("Block generated height=%d", resp.Height)
			}
		}
	}()
}

// MineMinXMRBalance enables mining for the passed wc wallet until it has an unlocked balance greater
// than or equal to minBalance.
func MineMinXMRBalance(t *testing.T, wc WalletClient, minBalance common.MoneroAmount) {
	daemonCli := monerorpc.New(MonerodRegtestEndpoint, nil).Daemon
	addr, err := wc.GetAddress(0)
	require.NoError(t, err)
	t.Log("mining to address:", addr.Address)

	for {
		require.NoError(t, wc.Refresh())
		balance, err := wc.GetBalance(0)
		require.NoError(t, err)
		if balance.UnlockedBalance > uint64(minBalance) {
			break
		}
		_, err = daemonCli.GenerateBlocks(&daemon.GenerateBlocksRequest{
			AmountOfBlocks: 32,
			WalletAddress:  addr.Address,
		})
		if err != nil && err.Error() == "Block not accepted" {
			continue
		}
		require.NoError(t, err)
	}
}
