// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/daemon"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/rpcclient"
	"github.com/athanorlabs/atomic-swap/tests"
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
	dataDir := t.TempDir()

	flags := []string{
		"testSwapd",
		fmt.Sprintf("--%s=dev", flagEnv),
		fmt.Sprintf("--%s=debug", flagLogLevel),
		fmt.Sprintf("--%s=true", flagDevXMRTaker),
		fmt.Sprintf("--%s=true", flagDeploy),
		fmt.Sprintf("--%s=%s", flagDataDir, dataDir),
		fmt.Sprintf("--%s=%d", flagRPCPort, rpcPort),
		fmt.Sprintf("--%s=0", flagLibp2pPort),
	}

	ctx, cancel := newTestContext(t)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := cliApp().RunContext(ctx, flags)
		assert.NoError(t, err)
	}()

	// Ensure the daemon fully started before we cancel the context
	daemon.WaitForSwapdStart(t, rpcPort)
	cancel()
	wg.Wait()

	if t.Failed() {
		return
	}

	//
	// Validate that --deploy created a contract address file.
	// At some future point, we will ask the RPC endpoint
	// what the contract addresses are instead of using this file.
	//
	data, err := os.ReadFile(path.Join(dataDir, contractAddressesFile))
	require.NoError(t, err)
	m := make(map[string]string)
	require.NoError(t, json.Unmarshal(data, &m))
	swapCreatorAddr, ok := m["swapCreatorAddr"]
	require.True(t, ok)

	ec, _ := tests.NewEthClient(t)
	ecCtx := context.Background()
	err = contracts.CheckSwapCreatorContractCode(ecCtx, ec, ethcommon.HexToAddress(swapCreatorAddr))
	require.NoError(t, err)
}

func TestDaemon_DevXMRMaker(t *testing.T) {
	rpcPort := getFreePort(t)
	key := tests.GetMakerTestKey(t)
	ec, _ := tests.NewEthClient(t)

	// We tested --deploy with the taker, so test passing the contract address here
	swapCreatorAddr, _, err := deploySwapCreator(context.Background(), ec, key, t.TempDir())
	require.NoError(t, err)

	flags := []string{
		"testSwapd",
		fmt.Sprintf("--%s", flagDevXMRMaker),
		fmt.Sprintf("--%s=dev", flagEnv),
		fmt.Sprintf("--%s=debug", flagLogLevel),
		fmt.Sprintf("--%s=%s", flagContractAddress, swapCreatorAddr),
		fmt.Sprintf("--%s=%s", flagDataDir, t.TempDir()),
		fmt.Sprintf("--%s=%d", flagRPCPort, rpcPort),
		fmt.Sprintf("--%s=0", flagLibp2pPort),
	}

	ctx, cancel := newTestContext(t)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := cliApp().RunContext(ctx, flags)
		assert.NoError(t, err)
	}()

	// Ensure the daemon fully started before we cancel the context
	daemon.WaitForSwapdStart(t, rpcPort)
	cancel()

	wg.Wait()
}

func TestDaemon_BadFlags(t *testing.T) {
	key := tests.GetMakerTestKey(t)
	ec, _ := tests.NewEthClient(t)
	ctx, _ := newTestContext(t)

	swapCreatorAddr, _, err := deploySwapCreator(ctx, ec, key, t.TempDir())
	require.NoError(t, err)

	baseFlags := []string{
		"testSwapd",
		fmt.Sprintf("--%s=dev", flagEnv),
		fmt.Sprintf("--%s=debug", flagLogLevel),
		fmt.Sprintf("--%s=%s", flagDataDir, t.TempDir()),
		fmt.Sprintf("--%s=0", flagRPCPort),
		fmt.Sprintf("--%s=0", flagLibp2pPort),
	}

	type testCase struct {
		description string
		extraFlags  []string
		expectErr   string
	}

	testCases := []testCase{
		{
			description: "no contract and no request to deploy",
			extraFlags:  nil,
			expectErr:   `flag "deploy" or "contract-address" is required for env=dev`,
		},
		{
			description: "pass invalid SwapCreator contract",
			extraFlags: []string{
				fmt.Sprintf("--%s=%s", flagContractAddress, ethcommon.Address{9}), // passing wrong contract
			},
			expectErr: "does not contain correct SwapCreator code",
		},
		{
			description: "pass SwapCreator contract an invalid address (wrong length)",
			extraFlags: []string{
				fmt.Sprintf("--%s=%s", flagContractAddress, "0xFFFF"), // too short
			},
			expectErr: fmt.Sprintf(`"%s" requires a valid ethereum address`, flagContractAddress),
		},
		{
			// this one also happens when people accidentally confuse swapd with swapcli
			description: "forgot to prefix the flag name with dashes",
			extraFlags: []string{
				flagContractAddress, swapCreatorAddr.String(),
			},
			expectErr: fmt.Sprintf("unknown command %q", flagContractAddress),
		},
	}

	for _, tc := range testCases {
		func() { // so we can call defer inside loop
			testCtx, cancel := context.WithTimeout(ctx, time.Second*15)
			defer cancel()
			flags := append(append([]string{}, baseFlags...), tc.extraFlags...)
			err := cliApp().RunContext(testCtx, flags)
			assert.ErrorContains(t, err, tc.expectErr, tc.description)
		}()
	}
}

func TestDaemon_PersistOffers(t *testing.T) {
	dataDir := t.TempDir()
	walletDir := path.Join(dataDir, "wallet")

	defer func() {
		// CI has issues with the filesystem still being written to when it is
		// recursively deleting dataDir. Can't be replicated outside of CI.
		unix.Sync()
		time.Sleep(500 * time.Millisecond)
	}()

	wc := monero.CreateWalletClientWithWalletDir(t, walletDir)
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
		fmt.Sprintf("--%s=%s", flagMoneroWalletPath, path.Join(walletDir, walletName)),
	}

	ctx1, cancel1 := newTestContext(t)

	var wg1 sync.WaitGroup
	wg1.Add(1)

	go func() {
		defer wg1.Done()
		err := cliApp().RunContext(ctx1, flags)
		assert.NoError(t, err)
		t.Logf("initial swapd instance exited")
	}()

	daemon.WaitForSwapdStart(t, rpcPort)
	if t.Failed() {
		return
	}

	// make an offer
	client := rpcclient.NewClient(ctx1, rpcEndpoint)
	balance, err := client.Balances(new(rpctypes.BalancesRequest))
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
		err := cliApp().RunContext(ctx2, flags) //nolint:govet
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
