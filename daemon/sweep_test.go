package daemon

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/rpcclient"
	"github.com/athanorlabs/atomic-swap/tests"
)

// This test starts a swap between Bob and Alice. Bob completes the swap
// successfully, and then both swapd daemons are shut down before Alice
// completes the sweep from the swap wallet back to her primary wallet.
// The Alice restarts, she should not try to sweep again and instead know
// that the swap is already complete.
func TestAliceStoppedAndRestartedDuringXMRSweep(t *testing.T) {
	minXMR := coins.StrToDecimal("1")
	maxXMR := minXMR
	exRate := coins.StrToExchangeRate("300")
	providesAmt, err := exRate.ToETH(minXMR)
	require.NoError(t, err)

	bobConf := CreateTestConf(t, tests.GetMakerTestKey(t))
	monero.MineMinXMRBalance(t, bobConf.MoneroClient, coins.MoneroToPiconero(maxXMR))

	aliceConf := CreateTestConf(t, tests.GetTakerTestKey(t))

	timeout := 7 * time.Minute
	ctx, cancel := LaunchDaemons(t, timeout, bobConf, aliceConf)

	bws, err := rpcclient.NewWsClient(ctx, bobConf.RPCPort)
	require.NoError(t, err)
	aws, err := rpcclient.NewWsClient(ctx, aliceConf.RPCPort)
	require.NoError(t, err)

	// Use an independent context for these clients that will execute across 2 runs of the daemons
	bc := rpcclient.NewClient(context.Background(), bobConf.RPCPort)
	ac := rpcclient.NewClient(context.Background(), aliceConf.RPCPort)

	tokenAddr := GetMockTokens(t, aliceConf.EthereumClient)[MockTether]
	tokenAsset := types.EthAsset(tokenAddr)

	makeResp, bobStatusCh, err := bws.MakeOfferAndSubscribe(minXMR, maxXMR, exRate, tokenAsset, false)
	require.NoError(t, err)

	aliceStatusCh, err := aws.TakeOfferAndSubscribe(makeResp.PeerID, makeResp.OfferID, providesAmt)
	require.NoError(t, err)

	var statusWG sync.WaitGroup
	statusWG.Add(2)

	// Test that Bob completes the swap successfully
	go func() {
		defer statusWG.Done()
		for {
			select {
			case status := <-bobStatusCh:
				t.Log("> Bob got status:", status)
				if !status.IsOngoing() {
					assert.Equal(t, types.CompletedSuccess.String(), status.String())
					// note: we can cancel the context here, but there's a chance
					// Alice hasn't started sweeping yet, in which case she'll start
					// the sweep on restart. either way, the test passes
					return
				}
			case <-ctx.Done():
				t.Errorf("Bob's context cancelled before he completed the swap [unexpected]")
				return
			}
		}
	}()

	// In theory, Alice won't complete the swap in the background goroutine
	// below, because we shut down her server as soon as Bob's end succeeded,
	// and Bob would normally succeed first, while Alice is still sweeping XMR
	// funds from the swap wallet back to her primary wallet.
	go func() {
		defer statusWG.Done()
		for {
			select {
			case status := <-aliceStatusCh:
				t.Log("> Alice got status:", status)
				if !status.IsOngoing() {
					return
				}

				if status == types.SweepingXMR {
					// Bob should be done at this point, and Alice is sweeping
					// cancel and ensure the swaps are both complete on daemon restart
					cancel()
					t.Log("cancelling context of Alice's and Bob's servers")
					return
				}
			case <-ctx.Done():
				t.Logf("Alice's context cancelled before she completed the swap [expected]")
				return
			}
		}
	}()

	statusWG.Wait()
	t.Logf("Both swaps completed or cancelled")
	if t.Failed() {
		return
	}

	// Make sure both servers had time to fully shut down
	time.Sleep(3 * time.Second)

	// relaunch the daemons
	t.Logf("daemons stopped, now re-launching them")
	_, _ = LaunchDaemons(t, 3*time.Minute, bobConf, aliceConf)

	t.Logf("daemon's relaunched, giving Alice 10 seconds to complete swap before query")
	time.Sleep(10 * time.Second) // give alice time to complete the swap

	pastSwap, err := ac.GetPastSwap(&makeResp.OfferID)
	require.NoError(t, err)
	t.Logf("Alice past status: %s", pastSwap.Swaps[0].Status)
	require.Equal(t, types.CompletedSuccess, pastSwap.Swaps[0].Status)

	pastSwap, err = bc.GetPastSwap(&makeResp.OfferID)
	require.NoError(t, err)
	t.Logf("Bob past status: %s", pastSwap.Swaps[0].Status)
	require.Equal(t, types.CompletedSuccess, pastSwap.Swaps[0].Status)
}
