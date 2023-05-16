// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

//go:build !prod

package daemon

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net"
	"path"
	"sync"
	"syscall"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/bootnode"
	"github.com/athanorlabs/atomic-swap/common"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/rpcclient"
	"github.com/athanorlabs/atomic-swap/tests"
)

// map indexes for our mock tokens
const (
	MockDAI    = "DAI"
	MockTether = "USDT"
)

// This file is only for test support. Use the build tag "prod" to prevent
// symbols in this file from consuming space in production binaries.

// CreateTestConf creates a localhost-only dev environment SwapdConfig config
// for testing
func CreateTestConf(t *testing.T, ethKey *ecdsa.PrivateKey) *SwapdConfig {
	ctx := context.Background()
	ec, err := extethclient.NewEthClient(ctx, common.Development, common.DefaultGanacheEndpoint, ethKey)
	require.NoError(t, err)
	t.Cleanup(func() {
		ec.Close()
	})

	rpcPort, err := common.GetFreeTCPPort()
	require.NoError(t, err)

	// We need a copy of the environment conf, as it is no longer a singleton
	// when we are testing it here.
	envConf := new(common.Config)
	*envConf = *common.ConfigDefaultsForEnv(common.Development)
	envConf.DataDir = t.TempDir()
	// Passed in ETH key may not have funds, deploy contract with the funded taker key
	envConf.SwapCreatorAddr, _ = contracts.DevDeploySwapCreator(t, ec.Raw(), tests.GetTakerTestKey(t))

	return &SwapdConfig{
		EnvConf:        envConf,
		MoneroClient:   monero.CreateWalletClient(t),
		EthereumClient: ec,
		Libp2pPort:     0,
		Libp2pKeyfile:  "",
		RPCPort:        uint16(rpcPort),
		IsRelayer:      false,
		NoTransferBack: false,
	}
}

// CreateTestBootnode creates a bootnode for unit tests that is automatically
// cleaned up when the test completes. Returns the local RPC port and P2P
// address for the node.
func CreateTestBootnode(t *testing.T) (uint16, string) {
	rpcPort, err := common.GetFreeTCPPort()
	require.NoError(t, err)

	// The bootnode uses an independent context from any swapd instances and
	// will not exit until the end of the test. To shut it down early, you can
	// use the shutdown RPC method on it.
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	t.Cleanup(func() {
		cancel()
		wg.Wait()
	})

	dataDir := t.TempDir()

	conf := &bootnode.Config{
		Env:           common.Development,
		DataDir:       t.TempDir(),
		Bootnodes:     nil,
		P2PListenIP:   "127.0.0.1",
		Libp2pPort:    0,
		Libp2pKeyFile: path.Join(dataDir, common.DefaultLibp2pKeyFileName),
		RPCPort:       uint16(rpcPort),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := bootnode.RunBootnode(ctx, conf) //nolint:govet
		require.ErrorIs(t, err, context.Canceled)
		t.Log("bootnode exited")
	}()
	WaitForSwapdStart(t, conf.RPCPort)

	endpoint := conf.RPCPort
	addresses, err := rpcclient.NewClient(ctx, endpoint).Addresses()
	require.NoError(t, err)
	require.NotEmpty(t, addresses)

	return conf.RPCPort, addresses.Addrs[0]
}

// LaunchDaemons launches one or more swapd daemons and blocks until they are
// started. If more than one config is passed, the bootnode settings of the
// passed config are modified to make the first daemon the bootnode for the
// remaining daemons.
func LaunchDaemons(t *testing.T, timeout time.Duration, configs ...*SwapdConfig) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	// We'll use the bootnode(s) in the first config for any config that does
	// not have the bootnodes field set. If the first config has no configured
	// bootnodes, we create one.
	require.NotEmpty(t, configs)
	bootnodes := configs[0].EnvConf.Bootnodes
	if len(bootnodes) == 0 {
		_, node := CreateTestBootnode(t)
		bootnodes = []string{node}
	}

	// ensure all swapd instances have exited before we let the test complete
	var wg sync.WaitGroup
	t.Cleanup(func() {
		cancel()
		wg.Wait()
	})

	for n, conf := range configs {
		if len(conf.EnvConf.Bootnodes) == 0 {
			conf.EnvConf.Bootnodes = bootnodes
		}

		wg.Add(1)
		go func(confIndex int) {
			defer wg.Done()
			err := RunSwapDaemon(ctx, conf)
			require.ErrorIs(t, err, context.Canceled)
			t.Logf("swapd#%d exited", confIndex)
		}(n)
		WaitForSwapdStart(t, conf.RPCPort)
	}

	return ctx, cancel
}

// WaitForSwapdStart takes the rpcPort of a swapd instance and waits for it to
// be in a listening state. Fails the test if the server isn't listening after a
// little over 60 seconds.
func WaitForSwapdStart(t *testing.T, rpcPort uint16) {
	const maxSeconds = 60
	addr := fmt.Sprintf("127.0.0.1:%d", rpcPort)

	startTime := time.Now()

	for i := 0; i < maxSeconds; i++ {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			startupTime := time.Since(startTime).Round(time.Second)
			t.Logf("daemon on rpc port %d started after %s", rpcPort, startupTime)
			require.NoError(t, conn.Close())
			return
		}
		// DialTimeout doesn't do retries. If the connection was refused, it happened
		// almost immediately, so we still need to sleep.
		require.ErrorIs(t, err, syscall.ECONNREFUSED)
		time.Sleep(time.Second)
	}
	t.Fatalf("giving up, swapd RPC port %d is not listening after %d seconds", rpcPort, maxSeconds)
}

// these variables are only for use by GetMockTokens
var _mockTokens map[string]ethcommon.Address
var _mockTokensMu sync.Mutex

// GetMockTokens returns a symbol=>address map of our mock ERC20 tokens,
// deploying them if they haven't already been deployed. Use the constants
// defined earlier to access the map elements.
func GetMockTokens(t *testing.T, ec extethclient.EthClient) map[string]ethcommon.Address {
	_mockTokensMu.Lock()
	defer _mockTokensMu.Unlock()

	if _mockTokens == nil {
		_mockTokens = make(map[string]ethcommon.Address)
	}

	calcSupply := func(numStdUnits int64, decimals int64) *big.Int {
		return new(big.Int).Mul(big.NewInt(numStdUnits), new(big.Int).Exp(big.NewInt(10), big.NewInt(decimals), nil))
	}

	ctx := context.Background()
	txOpts, err := ec.TxOpts(ctx)
	require.NoError(t, err)

	mockDaiAddr, mockDaiTx, _, err := contracts.DeployTestERC20(
		txOpts,
		ec.Raw(),
		"Dai Stablecoin",
		"DAI",
		18,
		ec.Address(),
		calcSupply(1000, 18),
	)
	require.NoError(t, err)
	tests.MineTransaction(t, ec.Raw(), mockDaiTx)
	_mockTokens[MockDAI] = mockDaiAddr

	txOpts, err = ec.TxOpts(ctx)
	require.NoError(t, err)

	mockTetherAddr, mockTetherTx, _, err := contracts.DeployTestERC20(
		txOpts,
		ec.Raw(),
		"Tether USD",
		"USDT",
		6,
		ec.Address(),
		calcSupply(1000, 6),
	)
	require.NoError(t, err)
	tests.MineTransaction(t, ec.Raw(), mockTetherTx)
	_mockTokens[MockTether] = mockTetherAddr

	return _mockTokens
}
