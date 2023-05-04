// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package daemon

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/rpcclient"
	"github.com/athanorlabs/atomic-swap/rpcclient/wsclient"
	"github.com/athanorlabs/atomic-swap/tests"
)

const (
	// transferGas is the amount of gas to perform a standard ETH transfer
	transferGas = 21000
)

func init() {
	cliutil.SetLogLevels("debug")
}

func privKeyToAddr(privKey *ecdsa.PrivateKey) ethcommon.Address {
	return crypto.PubkeyToAddress(*privKey.Public().(*ecdsa.PublicKey))
}

func transfer(t *testing.T, fromKey *ecdsa.PrivateKey, toAddress ethcommon.Address, ethAmount *apd.Decimal) {
	ctx := context.Background()
	ec, chainID := tests.NewEthClient(t)
	fromAddress := privKeyToAddr(fromKey)

	gasPrice, err := ec.SuggestGasPrice(ctx)
	require.NoError(t, err)

	nonce, err := ec.PendingNonceAt(ctx, fromAddress)
	require.NoError(t, err)

	weiAmount := coins.EtherToWei(ethAmount).BigInt()

	tx := ethtypes.NewTx(&ethtypes.LegacyTx{
		Nonce:    nonce,
		To:       &toAddress,
		Value:    weiAmount,
		Gas:      transferGas,
		GasPrice: gasPrice,
	})
	signedTx, err := ethtypes.SignTx(tx, ethtypes.LatestSignerForChainID(chainID), fromKey)
	require.NoError(t, err)

	err = ec.SendTransaction(ctx, signedTx)
	require.NoError(t, err)
	_, err = block.WaitForReceipt(ctx, ec, signedTx.Hash())
	require.NoError(t, err)
}

// minimumFundAlice gives Alice enough ETH to do everything but relay a claim
func minimumFundAlice(t *testing.T, ec extethclient.EthClient, providesAmt *apd.Decimal) {
	fundingKey := tests.GetTakerTestKey(t)

	const (
		aliceGasRation = contracts.MaxNewSwapETHGas + contracts.MaxSetReadyGas + contracts.MaxRefundETHGas
	)
	// We give Alice enough gas money to refund if needed, but not enough to
	// relay a claim
	suggestedGasPrice, err := ec.Raw().SuggestGasPrice(context.Background())
	require.NoError(t, err)
	gasCostWei := new(big.Int).Mul(suggestedGasPrice, big.NewInt(aliceGasRation))
	fundAmt := new(apd.Decimal)
	_, err = coins.DecimalCtx().Add(fundAmt, providesAmt, coins.NewWeiAmount(gasCostWei).AsEther())
	require.NoError(t, err)
	transfer(t, fundingKey, ec.Address(), fundAmt)

	bal, err := ec.Balance(context.Background())
	require.NoError(t, err)
	t.Logf("Alice's start balance is: %s ETH", bal.AsEtherString())
}

// Tests the scenario, where Bob has no ETH, there are no advertised relayers in
// the network, and Alice relays Bob's claim.
func TestRunSwapDaemon_SwapBobHasNoEth_AliceRelaysClaim(t *testing.T) {
	minXMR := coins.StrToDecimal("1")
	maxXMR := minXMR
	exRate := coins.StrToExchangeRate("0.1")
	providesAmt, err := exRate.ToETH(minXMR)
	require.NoError(t, err)

	bobEthKey, err := crypto.GenerateKey() // Bob has no ETH (not a ganache key)
	require.NoError(t, err)
	bobConf := CreateTestConf(t, bobEthKey)
	monero.MineMinXMRBalance(t, bobConf.MoneroClient, coins.MoneroToPiconero(maxXMR))

	aliceConf := CreateTestConf(t, tests.GetTakerTestKey(t))

	timeout := 7 * time.Minute
	ctx, _ := LaunchDaemons(t, timeout, bobConf, aliceConf)

	bc, err := wsclient.NewWsClient(ctx, fmt.Sprintf("ws://127.0.0.1:%d/ws", bobConf.RPCPort))
	require.NoError(t, err)
	ac, err := wsclient.NewWsClient(ctx, fmt.Sprintf("ws://127.0.0.1:%d/ws", aliceConf.RPCPort))
	require.NoError(t, err)

	useRelayer := false // Bob will use the relayer regardless, because he has no ETH
	makeResp, bobStatusCh, err := bc.MakeOfferAndSubscribe(minXMR, maxXMR, exRate, types.EthAssetETH, useRelayer)
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
	_, err = coins.DecimalCtx().Sub(expectedBal, providesAmt, coins.RelayerFeeETH)
	require.NoError(t, err)

	bobBalance, err := bobConf.EthereumClient.Balance(ctx)
	require.NoError(t, err)

	require.Equal(t, expectedBal.Text('f'), bobBalance.AsEtherString())
}

// Tests the scenario where Bob has no ETH, he can't find an advertised relayer,
// and Alice does not have enough ETH to relay his claim. The end result should
// be a refund. Note that this test has a long pause, as the refund cannot
// happen until T2 expires.
func TestRunSwapDaemon_NoRelayersAvailable_Refund(t *testing.T) {
	minXMR := coins.StrToDecimal("1")
	maxXMR := minXMR
	exRate := coins.StrToExchangeRate("0.1")
	providesAmt, err := exRate.ToETH(minXMR)
	require.NoError(t, err)

	bobEthKey, err := crypto.GenerateKey() // Bob has no ETH (not a ganache key)
	require.NoError(t, err)
	bobConf := CreateTestConf(t, bobEthKey)
	monero.MineMinXMRBalance(t, bobConf.MoneroClient, coins.MoneroToPiconero(maxXMR))

	aliceEthKey, err := crypto.GenerateKey() // Alice has non-ganache key that we fund
	require.NoError(t, err)
	aliceConf := CreateTestConf(t, aliceEthKey)
	minimumFundAlice(t, aliceConf.EthereumClient, providesAmt)

	timeout := 8 * time.Minute
	ctx, _ := LaunchDaemons(t, timeout, bobConf, aliceConf)

	bc, err := wsclient.NewWsClient(ctx, fmt.Sprintf("ws://127.0.0.1:%d/ws", bobConf.RPCPort))
	require.NoError(t, err)
	ac, err := wsclient.NewWsClient(ctx, fmt.Sprintf("ws://127.0.0.1:%d/ws", aliceConf.RPCPort))
	require.NoError(t, err)

	useRelayer := false // Bob will use unsuccessfully use the relayer regardless, because he has no ETH
	makeResp, bobStatusCh, err := bc.MakeOfferAndSubscribe(minXMR, maxXMR, exRate, types.EthAssetETH, useRelayer)
	require.NoError(t, err)

	aliceStatusCh, err := ac.TakeOfferAndSubscribe(makeResp.PeerID, makeResp.OfferID, providesAmt)
	require.NoError(t, err)

	var statusWG sync.WaitGroup
	statusWG.Add(2)

	// Ensure Alice completes the swap with a refund
	go func() {
		defer statusWG.Done()
		for {
			select {
			case status := <-aliceStatusCh:
				t.Log("> Alice got status:", status)
				if !status.IsOngoing() {
					assert.Equal(t, types.CompletedRefund.String(), status.String())
					return
				}
			case <-ctx.Done():
				t.Errorf("Alice's context cancelled before she completed the swap")
				return
			}
		}
	}()

	// Test that Bob completes the swap as a refund
	go func() {
		defer statusWG.Done()
		for {
			select {
			case status := <-bobStatusCh:
				t.Log("> Bob got status:", status)
				if !status.IsOngoing() {
					assert.Equal(t, types.CompletedRefund.String(), status.String())
					return
				}
			case <-ctx.Done():
				t.Errorf("Bob's context cancelled before he completed the swap")
				return
			}
		}
	}()

	statusWG.Wait()
}

// Tests the scenario where Bob has no ETH and Charlie, an advertised relayer,
// performs the relay so Bob can get his ETH. To ensure that the test does not
// succeed by Alice relaying the claim, we ensure that Alice does not have
// enough ETH left over after the swap to relay.
func TestRunSwapDaemon_CharlieRelays(t *testing.T) {
	minXMR := coins.StrToDecimal("1")
	maxXMR := minXMR
	exRate := coins.StrToExchangeRate("0.1")
	providesAmt, err := exRate.ToETH(minXMR)
	require.NoError(t, err)

	bobEthKey, err := crypto.GenerateKey() // Bob has no ETH (not a ganache key)
	require.NoError(t, err)
	bobConf := CreateTestConf(t, bobEthKey)
	monero.MineMinXMRBalance(t, bobConf.MoneroClient, coins.MoneroToPiconero(maxXMR))

	// Configure Alice with enough funds to complete the swap, but not to relay Bob's claim
	aliceEthKey, err := crypto.GenerateKey() // Alice gets a key without enough funds to relay
	require.NoError(t, err)
	aliceConf := CreateTestConf(t, aliceEthKey)
	minimumFundAlice(t, aliceConf.EthereumClient, providesAmt)

	// Charlie can safely use the taker key, as Alice is not using it.
	charlieConf := CreateTestConf(t, tests.GetTakerTestKey(t))
	charlieConf.IsRelayer = true
	charlieStartBal, err := charlieConf.EthereumClient.Balance(context.Background())
	require.NoError(t, err)

	timeout := 7 * time.Minute
	ctx, _ := LaunchDaemons(t, timeout, bobConf, aliceConf, charlieConf)

	bc, err := wsclient.NewWsClient(ctx, fmt.Sprintf("ws://127.0.0.1:%d/ws", bobConf.RPCPort))
	require.NoError(t, err)
	ac, err := wsclient.NewWsClient(ctx, fmt.Sprintf("ws://127.0.0.1:%d/ws", aliceConf.RPCPort))
	require.NoError(t, err)

	useRelayer := false // Bob will use the relayer regardless, because he has no ETH
	makeResp, bobStatusCh, err := bc.MakeOfferAndSubscribe(minXMR, maxXMR, exRate, types.EthAssetETH, useRelayer)
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

	// Ensure Bob completes the swap successfully
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
	bobExpectedBal := new(apd.Decimal)
	_, err = coins.DecimalCtx().Sub(bobExpectedBal, providesAmt, coins.RelayerFeeETH)
	require.NoError(t, err)
	bobBalance, err := bobConf.EthereumClient.Balance(ctx)
	require.NoError(t, err)
	require.Equal(t, bobExpectedBal.Text('f'), bobBalance.AsEtherString())

	//
	// Charlie should be wealthier now than at the start, despite paying the claim
	// gas, because he received the relayer fee.
	//
	charlieEC := charlieConf.EthereumClient
	charlieBal, err := charlieEC.Balance(ctx)
	require.NoError(t, err)
	require.Greater(t, charlieBal.Cmp(charlieStartBal), 0)
	charlieProfitWei := charlieBal.Sub(charlieStartBal)
	t.Logf("Charlie earned %s ETH", charlieProfitWei.AsEtherString())
}

// Tests the scenario where Charlie, an advertised relayer, has run out of ETH
// and cannot relay Alice's request. Bob falls back to Alice as the relayer of
// last resort, and she relays his claim.
func TestRunSwapDaemon_CharlieIsBroke_AliceRelays(t *testing.T) {
	minXMR := coins.StrToDecimal("1")
	maxXMR := minXMR
	exRate := coins.StrToExchangeRate("0.1")
	providesAmt, err := exRate.ToETH(minXMR)
	require.NoError(t, err)

	bobEthKey, err := crypto.GenerateKey() // Bob has no ETH (not a ganache key)
	require.NoError(t, err)
	bobConf := CreateTestConf(t, bobEthKey)
	monero.MineMinXMRBalance(t, bobConf.MoneroClient, coins.MoneroToPiconero(maxXMR))

	// Alice is fully funded with the taker key
	aliceConf := CreateTestConf(t, tests.GetTakerTestKey(t))

	// Charlie is a relayer, but he has no ETH
	charlieEthKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	charlieConf := CreateTestConf(t, charlieEthKey)
	charlieConf.IsRelayer = true

	timeout := 7 * time.Minute
	ctx, _ := LaunchDaemons(t, timeout, bobConf, aliceConf, charlieConf)

	bc, err := wsclient.NewWsClient(ctx, fmt.Sprintf("ws://127.0.0.1:%d/ws", bobConf.RPCPort))
	require.NoError(t, err)
	ac, err := wsclient.NewWsClient(ctx, fmt.Sprintf("ws://127.0.0.1:%d/ws", aliceConf.RPCPort))
	require.NoError(t, err)

	useRelayer := false // Bob will use the relayer regardless, because he has no ETH
	makeResp, bobStatusCh, err := bc.MakeOfferAndSubscribe(minXMR, maxXMR, exRate, types.EthAssetETH, useRelayer)
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

	// Ensure Bob completes the swap successfully
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
	bobExpectedBal := new(apd.Decimal)
	_, err = coins.DecimalCtx().Sub(bobExpectedBal, providesAmt, coins.RelayerFeeETH)
	require.NoError(t, err)
	bobBalance, err := bobConf.EthereumClient.Balance(ctx)
	require.NoError(t, err)
	require.Equal(t, bobExpectedBal.Text('f'), bobBalance.AsEtherString())
}

// Tests the version and shutdown RPC methods
func TestRunSwapDaemon_RPC_Version(t *testing.T) {
	conf := CreateTestConf(t, tests.GetMakerTestKey(t))
	protocolVersion := fmt.Sprintf("%s/%d", net.ProtocolID, conf.EthereumClient.ChainID())
	timeout := time.Minute
	ctx, _ := LaunchDaemons(t, timeout, conf)

	c := rpcclient.NewClient(ctx, fmt.Sprintf("http://127.0.0.1:%d", conf.RPCPort))
	versionResp, err := c.Version()
	require.NoError(t, err)

	require.Equal(t, conf.EnvConf.Env, versionResp.Env)
	require.NotEmpty(t, versionResp.SwapdVersion)
	require.Equal(t, conf.EnvConf.SwapCreatorAddr, versionResp.SwapCreatorAddr)
	require.Equal(t, protocolVersion, versionResp.P2PVersion)
}

// Tests the shutdown RPC method
func TestRunSwapDaemon_RPC_Shutdown(t *testing.T) {
	conf := CreateTestConf(t, tests.GetMakerTestKey(t))
	timeout := time.Minute
	ctx, _ := LaunchDaemons(t, timeout, conf)

	c := rpcclient.NewClient(ctx, fmt.Sprintf("http://127.0.0.1:%d", conf.RPCPort))
	err := c.Shutdown()
	require.NoError(t, err)

	err = c.Shutdown()
	require.ErrorIs(t, err, syscall.ECONNREFUSED)
}
