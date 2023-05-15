// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package daemon

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/rpcclient"
	"github.com/athanorlabs/atomic-swap/rpcclient/wsclient"
	"github.com/athanorlabs/atomic-swap/tests"
)

// This test starts a swap between Bob and Alice. The nodes are shut down
// after the key exchange step, while Alice is trying to lock funds,
// and the newSwap tx is in the mempool but not yet included.
// Alice's node is then restarted, and depending on whether the newSwap
// tx was included yet or not, she should be able to cancel or refund the swap.
// In this case, the tx is always included by the time she restarts,
// so she refunds the swap.
// Bob should have aborted the swap in all cases.
func TestXMRTakerCancelOrRefundAfterKeyExchange(t *testing.T) {
	minXMR := coins.StrToDecimal("1")
	maxXMR := minXMR
	exRate := coins.StrToExchangeRate("300")
	providesAmt, err := exRate.ToETH(minXMR)
	require.NoError(t, err)

	ec, err := ethclient.Dial("http://localhost:8545")
	require.NoError(t, err)

	bobConf := CreateTestConf(t, tests.GetMakerTestKey(t))
	monero.MineMinXMRBalance(t, bobConf.MoneroClient, coins.MoneroToPiconero(maxXMR))

	aliceConf := CreateTestConf(t, tests.GetTakerTestKey(t))

	timeout := 7 * time.Minute
	ctx, cancel := LaunchDaemons(t, timeout, bobConf, aliceConf)

	bws, err := wsclient.NewWsClient(ctx, fmt.Sprintf("ws://127.0.0.1:%d/ws", bobConf.RPCPort))
	require.NoError(t, err)
	aws, err := wsclient.NewWsClient(ctx, fmt.Sprintf("ws://127.0.0.1:%d/ws", aliceConf.RPCPort))
	require.NoError(t, err)

	// Use an independent context for these clients that will execute across multiple runs of the daemons
	bc := rpcclient.NewClient(context.Background(), fmt.Sprintf("http://127.0.0.1:%d", bobConf.RPCPort))
	ac := rpcclient.NewClient(context.Background(), fmt.Sprintf("http://127.0.0.1:%d", aliceConf.RPCPort))

	makeResp, bobStatusCh, err := bws.MakeOfferAndSubscribe(minXMR, maxXMR, exRate, types.EthAssetETH, false)
	require.NoError(t, err)

	aliceStatusCh, err := aws.TakeOfferAndSubscribe(makeResp.PeerID, makeResp.OfferID, providesAmt)
	require.NoError(t, err)

	var statusWG sync.WaitGroup
	statusWG.Add(2)

	go func() {
		for {
			count, err := ec.PendingTransactionCount(ctx) //nolint:govet
			require.NoError(t, err)
			if count > 0 {
				// the newSwap tx is in the mempool, shut down the nodes
				cancel()
				t.Log("cancelling context of Alice's and Bob's servers")
				return
			}
			time.Sleep(time.Millisecond * 200)
		}
	}()

	go func() {
		defer statusWG.Done()
		for {
			select {
			case status := <-bobStatusCh:
				t.Log("> Bob got status:", status)
				if !status.IsOngoing() {
					assert.Equal(t, types.CompletedAbort.String(), status.String())
					return
				}
			case <-ctx.Done():
				t.Logf("Bob's context cancelled before he completed the swap [expected]")
				return
			}
		}
	}()

	go func() {
		defer statusWG.Done()
		for {
			select {
			case status := <-aliceStatusCh:
				t.Log("> Alice got status:", status)
				if !status.IsOngoing() {
					return
				}
			case <-ctx.Done():
				t.Logf("Alice's context cancelled before she completed the swap [expected]")
				return
			}
		}
	}()

	statusWG.Wait()
	t.Logf("Both swaps cancelled")
	if t.Failed() {
		return
	}

	// Make sure both servers had time to fully shut down
	time.Sleep(3 * time.Second)

	t.Logf("daemons stopped, now re-launching them")
	_, _ = LaunchDaemons(t, 3*time.Minute, bobConf, aliceConf)

	pastSwap, err := ac.GetPastSwap(&makeResp.OfferID)
	require.NoError(t, err)
	t.Logf("Alice past status: %s", pastSwap.Swaps[0].Status)
	if pastSwap.Swaps[0].Status.String() != types.CompletedRefund.String() {
		t.Errorf("Alice should have refunded the swap")
	}

	pastSwap, err = bc.GetPastSwap(&makeResp.OfferID)
	require.NoError(t, err)
	t.Logf("Bob past status: %s", pastSwap.Swaps[0].Status)
	require.Equal(t, types.CompletedAbort.String(), pastSwap.Swaps[0].Status.String())
}
