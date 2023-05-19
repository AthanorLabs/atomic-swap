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

func TestXMRNotLockedAndETHRefundedAfterAliceRestarts(t *testing.T) {
	minXMR := coins.StrToDecimal("1")
	maxXMR := minXMR
	exRate := coins.StrToExchangeRate("0.1")
	providesAmt, err := exRate.ToETH(minXMR)
	require.NoError(t, err)

	bobConf := CreateTestConf(t, tests.GetMakerTestKey(t))
	monero.MineMinXMRBalance(t, bobConf.MoneroClient, coins.MoneroToPiconero(maxXMR))

	aliceConf := CreateTestConf(t, tests.GetTakerTestKey(t))

	timeout := 7 * time.Minute
	ctx, cancel := LaunchDaemons(t, timeout, aliceConf, bobConf)

	// clients use a separate context and will work across server restarts
	clientCtx := context.Background()
	bc := rpcclient.NewWsClient(clientCtx, bobConf.RPCPort)
	ac := rpcclient.NewWsClient(clientCtx, aliceConf.RPCPort)

	// Bob makes an offer
	makeResp, bobStatusCh, err := bc.MakeOfferAndSubscribe(minXMR, maxXMR, exRate, types.EthAssetETH, false)
	require.NoError(t, err)

	// Alice takes the offer
	aliceStatusCh, err := ac.TakeOfferAndSubscribe(makeResp.PeerID, makeResp.OfferID, providesAmt)
	require.NoError(t, err)

	var statusWG sync.WaitGroup
	statusWG.Add(2)

	// Alice shuts down both servers as soon as she locks her ETH
	go func() {
		defer statusWG.Done()
		for {
			select {
			case status := <-aliceStatusCh:
				t.Log("> Alice got status:", status)
				switch status {
				case types.ExpectingKeys:
					continue
				case types.ETHLocked:
					cancel() // stop both Alice's and Bob's daemons
				default:
					cancel()
					t.Errorf("Alice should not have reached status=%s", status)
				}
			case <-ctx.Done():
				t.Logf("Alice's context cancelled (expected)")
				return
			}
		}
	}()

	// Bob is not playing a significant role in this test. His swapd instance is
	// shut down before he can lock any XMR and we don't bring it back online.
	go func() {
		defer statusWG.Done()
		for {
			select {
			case status := <-bobStatusCh:
				t.Log("> Bob got status:", status)
				switch status {
				case types.KeysExchanged:
					continue
				default:
					cancel()
					t.Errorf("Bob should not have reached status=%s", status)
				}
			case <-ctx.Done():
				t.Logf("Bob's context cancelled (expected)")
				return
			}
		}
	}()

	statusWG.Wait()
	if t.Failed() {
		return
	}

	// Make sure both servers had time to fully shut down
	time.Sleep(3 * time.Second)

	// relaunch Alice's daemon
	t.Logf("daemons stopped, now re-launching Alice's daemon in isolation")
	ctx, cancel = LaunchDaemons(t, 3*time.Minute, aliceConf)

	aliceStatusCh, err = ac.SubscribeSwapStatus(makeResp.OfferID)
	require.NoError(t, err)

	// Ensure Alice completes the swap with a refund
	statusWG.Add(1)
	go func() {
		defer statusWG.Done()
		for {
			select {
			case status := <-aliceStatusCh:
				t.Log("> Alice, after restart, got status:", status)
				if !status.IsOngoing() {
					assert.Equal(t, types.CompletedRefund.String(), status.String())
					return
				}
			case <-ctx.Done():
				// Alice's context has a deadline. If we get here, the context
				// expired before we got any Refund status update.
				t.Errorf("Alice's context cancelled before she completed the swap")
				return
			}
		}
	}()

	statusWG.Wait()
	// TODO: Add some additional checks here when the rest of the test is working
}
