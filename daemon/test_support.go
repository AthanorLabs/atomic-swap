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
	"sync"
	"syscall"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

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
	ec, err := extethclient.NewEthClient(ctx, common.Development, common.DefaultEthEndpoint, ethKey)
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
	envConf.SwapCreatorAddr = getSwapCreatorAddress(t, ec.Raw())

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

// LaunchDaemons launches one or more swapd daemons and blocks until they are
// started. If more than one config is passed, the bootnode settings of the
// passed config are modified to make the first daemon the bootnode for the
// remaining daemons.
func LaunchDaemons(t *testing.T, timeout time.Duration, configs ...*SwapdConfig) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	var wg sync.WaitGroup
	t.Cleanup(func() {
		cancel()
		wg.Wait()
	})

	var bootNodes []string // First daemon to launch has no bootnodes

	for n, conf := range configs {
		conf.EnvConf.Bootnodes = bootNodes

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := RunSwapDaemon(ctx, conf)
			require.ErrorIs(t, err, context.Canceled)
		}()
		WaitForSwapdStart(t, conf.RPCPort)

		// Configure remaining daemons to use the first one a bootnode
		if n == 0 {
			c := rpcclient.NewClient(ctx, fmt.Sprintf("http://127.0.0.1:%d", conf.RPCPort))
			addresses, err := c.Addresses()
			require.NoError(t, err)
			require.Greater(t, len(addresses.Addrs), 1)
			bootNodes = []string{addresses.Addrs[0]}
		}
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

// these variables are only for use by getSwapCreatorAddress
var _swapCreatorAddr *ethcommon.Address
var _swapCreatorAddrMu sync.Mutex

func getSwapCreatorAddress(t *testing.T, ec *ethclient.Client) ethcommon.Address {
	_swapCreatorAddrMu.Lock()
	defer _swapCreatorAddrMu.Unlock()

	if _swapCreatorAddr != nil {
		return *_swapCreatorAddr
	}

	ctx := context.Background()
	ethKey := tests.GetTakerTestKey(t) // requester might not have ETH, so we don't pass the key in

	forwarderAddr, err := contracts.DeployGSNForwarderWithKey(ctx, ec, ethKey)
	require.NoError(t, err)

	swapCreatorAddr, _, err := contracts.DeploySwapCreatorWithKey(ctx, ec, ethKey, forwarderAddr)
	require.NoError(t, err)

	_swapCreatorAddr = &swapCreatorAddr
	return swapCreatorAddr
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
