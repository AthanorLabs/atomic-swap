// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package daemon

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/relayer"
	"github.com/athanorlabs/atomic-swap/rpcclient"
	"github.com/athanorlabs/atomic-swap/rpcclient/wsclient"
	"github.com/athanorlabs/atomic-swap/tests"
)

func init() {
	// alphabetically ordered
	level := "debug"
	_ = logging.SetLogLevel("cmd", level)
	_ = logging.SetLogLevel("coins", level)
	_ = logging.SetLogLevel("common", level)
	_ = logging.SetLogLevel("contracts", level)
	_ = logging.SetLogLevel("cmd", level)
	_ = logging.SetLogLevel("extethclient", level)
	_ = logging.SetLogLevel("ethereum/watcher", level)
	_ = logging.SetLogLevel("monero", level)
	_ = logging.SetLogLevel("net", level)
	_ = logging.SetLogLevel("offers", level)
	_ = logging.SetLogLevel("p2pnet", level) // external
	_ = logging.SetLogLevel("pricefeed", level)
	_ = logging.SetLogLevel("protocol", level)
	_ = logging.SetLogLevel("relayer", level) // external and internal
	_ = logging.SetLogLevel("rpc", level)
	_ = logging.SetLogLevel("xmrmaker", level)
	_ = logging.SetLogLevel("xmrtaker", level)
}

var _swapFactoryAddress *ethcommon.Address

func getSwapFactoryAddress(t *testing.T, ec *ethclient.Client) ethcommon.Address {
	if _swapFactoryAddress != nil {
		return *_swapFactoryAddress
	}

	ctx := context.Background()
	ethKey := tests.GetTakerTestKey(t) // requester might not have ETH, so we don't pass the key in

	forwarderAddr, err := contracts.DeployGSNForwarderWithKey(ctx, ec, ethKey)
	require.NoError(t, err)

	swapFactoryAddr, _, err := contracts.DeploySwapFactoryWithKey(ctx, ec, ethKey, forwarderAddr)
	require.NoError(t, err)

	_swapFactoryAddress = &swapFactoryAddr
	return swapFactoryAddr
}

func createTestConf(t *testing.T, ethKey *ecdsa.PrivateKey) *SwapdConfig {
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
	envConf.SwapFactoryAddress = getSwapFactoryAddress(t, ec.Raw())

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

// Tests the scenario, where Bob has no ETH and Alice relays his claim.
func TestRunSwapDaemon_SwapBobHasNoEth(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	t.Cleanup(func() {
		cancel()
	})

	minXMR := coins.StrToDecimal("1")
	maxXMR := minXMR
	exRate := coins.StrToExchangeRate("0.1")
	providesAmt, err := exRate.ToETH(minXMR)
	require.NoError(t, err)

	bobEthKey, err := crypto.GenerateKey() // Bob has no ETH (not a ganache key)
	require.NoError(t, err)
	bobConf := createTestConf(t, bobEthKey)
	monero.MineMinXMRBalance(t, bobConf.MoneroClient, coins.MoneroToPiconero(maxXMR))

	var stoppedWG sync.WaitGroup
	t.Cleanup(func() {
		cancel()
		stoppedWG.Wait() // ensure daemons are stopped even if require fails
	})

	stoppedWG.Add(1)
	go func() {
		defer stoppedWG.Done()
		err := RunSwapDaemon(ctx, bobConf) //nolint:govet
		require.ErrorIs(t, err, context.Canceled)
	}()
	WaitForSwapdStart(t, bobConf.RPCPort)

	bc := rpcclient.NewClient(ctx, fmt.Sprintf("http://127.0.0.1:%d", bobConf.RPCPort))
	bws, err := wsclient.NewWsClient(ctx, fmt.Sprintf("ws://127.0.0.1:%d/ws", bobConf.RPCPort))
	require.NoError(t, err)

	// Configure Alice to be a relayer and to use Bob as a bootnode
	aliceConf := createTestConf(t, tests.GetTakerTestKey(t))
	aliceConf.IsRelayer = true
	bobAddrs, err := bc.Addresses()
	require.NoError(t, err)
	require.Greater(t, len(bobAddrs.Addrs), 1)
	aliceConf.EnvConf.Bootnodes = []string{bobAddrs.Addrs[0]}

	stoppedWG.Add(1)
	go func() {
		defer stoppedWG.Done()
		err := RunSwapDaemon(ctx, aliceConf) //nolint:govet
		require.ErrorIs(t, err, context.Canceled)
	}()
	WaitForSwapdStart(t, aliceConf.RPCPort)

	useRelayer := false // Bob will use the relayer regardless, because he has no ETH
	makeResp, bobStatusCh, err := bws.MakeOfferAndSubscribe(minXMR, maxXMR, exRate, types.EthAssetETH, useRelayer)
	require.NoError(t, err)

	ac, err := wsclient.NewWsClient(ctx, fmt.Sprintf("ws://127.0.0.1:%d/ws", aliceConf.RPCPort))
	require.NoError(t, err)

	aliceStatusCh, err := ac.TakeOfferAndSubscribe(makeResp.PeerID, makeResp.OfferID, providesAmt)
	require.NoError(t, err)

	var statusWG sync.WaitGroup
	statusWG.Add(2)

	// Ensure Alice completes the swap successfully
	go func() {
		defer statusWG.Done()
		for {
			select {
			case status := <-aliceStatusCh:
				t.Log("> Alice got status:", status)
				if !status.IsOngoing() {
					assert.Equal(t, types.CompletedSuccess.String(), status.String())
					return
				}
			case <-ctx.Done():
				t.Errorf("Alice's context cancelled before she completed the swap")
				return
			}
		}
	}()

	// Test that Bob completes the swap successfully
	go func() {
		defer statusWG.Done()
		for {
			select {
			case status := <-bobStatusCh:
				t.Log("> Bob got status:", status)
				if !status.IsOngoing() {
					assert.Equal(t, types.CompletedSuccess.String(), status.String())
					return
				}
			case <-ctx.Done():
				t.Errorf("Bob's context cancelled before he completed the swap")
				return
			}
		}
	}()

	statusWG.Wait()
	if t.Failed() {
		return
	}

	//
	// Bob's ending balance should be Alice's provided amount minus the relayer fee
	//
	expectedBal := new(apd.Decimal)
	_, err = coins.DecimalCtx().Sub(expectedBal, providesAmt, relayer.FeeEth)
	require.NoError(t, err)

	bobBalance, err := bobConf.EthereumClient.Balance(ctx)
	require.NoError(t, err)

	require.Equal(t, expectedBal.Text('f'), coins.FmtWeiAsETH(bobBalance))
}
