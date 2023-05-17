// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package daemon

import (
	"sync"
	"testing"
	"time"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/rpcclient"
	"github.com/athanorlabs/atomic-swap/tests"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test if Alice is able to call the Refund() RPC API, used by swapcli refund,
// immediately after locking her ETH.
func TestRunSwapDaemon_ManualRefund(t *testing.T) {
	minXMR := coins.StrToDecimal("1")
	maxXMR := minXMR
	exRate := coins.StrToExchangeRate("0.1")
	providesAmt, err := exRate.ToETH(minXMR)
	require.NoError(t, err)

	bobEthKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	bobConf := CreateTestConf(t, bobEthKey)
	monero.MineMinXMRBalance(t, bobConf.MoneroClient, coins.MoneroToPiconero(maxXMR))

	aliceConf := CreateTestConf(t, tests.GetTakerTestKey(t))

	timeout := 7 * time.Minute
	ctx, _ := LaunchDaemons(t, timeout, bobConf, aliceConf)

	bc := rpcclient.NewClient(ctx, bobConf.RPCPort)
	ac := rpcclient.NewClient(ctx, aliceConf.RPCPort)

	makeResp, bobStatusCh, err := bc.MakeOfferAndSubscribe(minXMR, maxXMR, exRate, types.EthAssetETH, false)
	require.NoError(t, err)

	aliceStatusCh, err := ac.TakeOfferAndSubscribe(makeResp.PeerID, makeResp.OfferID, providesAmt)
	require.NoError(t, err)

	var statusWG sync.WaitGroup
	statusWG.Add(2)

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

				if status == types.ETHLocked {
					// wait for eth lock to get included
					time.Sleep(time.Second)

					// call refund
					t.Log("> Alice calling refund")
					refundResp, err := ac.Refund(makeResp.OfferID)
					require.NoError(t, err)

					ec, err := ethclient.Dial(common.DefaultGanacheEndpoint)
					require.NoError(t, err)

					t.Log("> Alice got refund response tx:", refundResp.TxHash)
					receipt, err := block.WaitForReceipt(ctx, ec, refundResp.TxHash)
					require.NoError(t, err)
					assert.Equal(t, uint64(1), receipt.Status)

					// manually trigger exit, since the xmrtaker doesn't watch for Refunded events.
					status, err := ac.Cancel(makeResp.OfferID)
					require.NoError(t, err)
					assert.Equal(t, types.CompletedRefund.String(), status.String())
					return
				}
			case <-ctx.Done():
				t.Errorf("Alice's context cancelled before she completed the swap")
				return
			}
		}
	}()

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
	if t.Failed() {
		return
	}
}
