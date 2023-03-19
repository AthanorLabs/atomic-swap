package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/daemon"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/rpcclient"
)

func newTestContext(t *testing.T) (context.Context, context.CancelFunc) {
	// The only external program any test in this package calls is monero-wallet-rpc, so we
	// make monero-bin the only directory in our path.
	curDir, err := os.Getwd()
	require.NoError(t, err)
	projectRoot := path.Dir(path.Dir(curDir)) // 2 dirs up from cmd/swaprecover
	t.Setenv("PATH", path.Join(projectRoot, "monero-bin"))

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
	})

	return ctx, cancel
}

func getFreePort(t *testing.T) uint16 {
	port, err := common.GetFreeTCPPort()
	require.NoError(t, err)
	return uint16(port)
}

func TestDaemon_DevXMRTaker(t *testing.T) {
	rpcPort := getFreePort(t)

	flags := []string{
		"testSwapd",
		fmt.Sprintf("--%s=dev", flagEnv),
		fmt.Sprintf("--%s=debug", flagLogLevel),
		fmt.Sprintf("--%s=true", flagDevXMRTaker),
		fmt.Sprintf("--%s=true", flagDeploy),
		fmt.Sprintf("--%s=%s", flagDataDir, t.TempDir()),
		fmt.Sprintf("--%s=%d", flagRPCPort, rpcPort),
		fmt.Sprintf("--%s=0", flagLibp2pPort),
	}

	ctx, cancel := newTestContext(t)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := app.RunContext(ctx, flags)
		assert.NoError(t, err)
	}()

	// Ensure the daemon fully started before we cancel the context
	daemon.WaitForSwapdStart(t, rpcPort)
	cancel()

	wg.Wait()
}

func TestDaemon_DevXMRMaker(t *testing.T) {
	rpcPort := getFreePort(t)

	flags := []string{
		"testSwapd",
		fmt.Sprintf("--%s", flagDevXMRMaker),
		fmt.Sprintf("--%s=dev", flagEnv),
		fmt.Sprintf("--%s=debug", flagLogLevel),
		fmt.Sprintf("--%s", flagDeploy),
		fmt.Sprintf("--%s=%s", flagDataDir, t.TempDir()),
		fmt.Sprintf("--%s=%d", flagRPCPort, rpcPort),
		fmt.Sprintf("--%s=0", flagLibp2pPort),
	}

	ctx, cancel := newTestContext(t)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := app.RunContext(ctx, flags)
		assert.NoError(t, err)
	}()

	// Ensure the daemon fully started before we cancel the context
	daemon.WaitForSwapdStart(t, rpcPort)
	cancel()

	wg.Wait()
}

func TestDaemon_PersistOffers(t *testing.T) {
	dataDir := t.TempDir()
	defer func() {
		// CI has issues with the filesystem still being written to when it is
		// recursively deleting dataDir. Can't be replicated outside of CI.
		unix.Sync()
		time.Sleep(500 * time.Millisecond)
	}()

	wc := monero.CreateWalletClientWithWalletDir(t, dataDir)
	one := apd.New(1, 0)
	monero.MineMinXMRBalance(t, wc, coins.MoneroToPiconero(one))
	walletName := wc.WalletName()
	wc.Close() // wallet file stays in place with mined monero

	rpcPort := getFreePort(t)
	rpcEndpoint := fmt.Sprintf("http://127.0.0.1:%d", rpcPort)

	flags := []string{
		"testSwapd",
		fmt.Sprintf("--%s", flagDevXMRMaker),
		fmt.Sprintf("--%s=dev", flagEnv),
		fmt.Sprintf("--%s=debug", flagLogLevel),
		fmt.Sprintf("--%s", flagDeploy),
		fmt.Sprintf("--%s=%s", flagDataDir, dataDir),
		fmt.Sprintf("--%s=%d", flagRPCPort, rpcPort),
		fmt.Sprintf("--%s=0", flagLibp2pPort),
		fmt.Sprintf("--%s=%s", flagMoneroWalletPath, path.Join(dataDir, walletName)),
	}

	ctx1, cancel1 := newTestContext(t)

	var wg1 sync.WaitGroup
	wg1.Add(1)

	go func() {
		defer wg1.Done()
		err := app.RunContext(ctx1, flags)
		assert.NoError(t, err)
		t.Logf("initial swapd instance exited")
	}()

	daemon.WaitForSwapdStart(t, rpcPort)
	if t.Failed() {
		return
	}

	// make an offer
	client := rpcclient.NewClient(ctx1, rpcEndpoint)
	balance, err := client.Balances()
	require.NoError(t, err)
	require.GreaterOrEqual(t, balance.PiconeroUnlockedBalance.Cmp(coins.MoneroToPiconero(one)), 0)

	minXMRAmt := coins.StrToDecimal("0.1")
	maxXMRAmt := one
	xRate := coins.ToExchangeRate(one)

	offerResp, err := client.MakeOffer(minXMRAmt, maxXMRAmt, xRate, types.EthAssetETH, false)
	require.NoError(t, err)

	// shut down the daemon to verify that the offer still exists on restart
	t.Logf("shutting down initial swapd instance")
	cancel1()
	wg1.Wait()

	// restart daemon
	t.Log("restarting daemon")
	ctx2, cancel2 := newTestContext(t)

	var wg2 sync.WaitGroup
	wg2.Add(1)
	t.Cleanup(func() {
		cancel2()
		wg2.Wait()
	})

	go func() {
		defer wg2.Done()
		err := app.RunContext(ctx2, flags) //nolint:govet
		assert.NoError(t, err)
	}()

	daemon.WaitForSwapdStart(t, rpcPort)

	client = rpcclient.NewClient(ctx2, rpcEndpoint)
	resp, err := client.GetOffers()
	require.NoError(t, err)
	require.Equal(t, offerResp.PeerID, resp.PeerID)
	require.Equal(t, 1, len(resp.Offers))
	require.Equal(t, offerResp.OfferID, resp.Offers[0].ID)
}
