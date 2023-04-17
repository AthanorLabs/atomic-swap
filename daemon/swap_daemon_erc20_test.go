package daemon

import (
	"fmt"
	"sync"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/rpcclient"
	"github.com/athanorlabs/atomic-swap/rpcclient/wsclient"
	"github.com/athanorlabs/atomic-swap/tests"
)

// Tests the scenario where Bob has XMR and enough ETH to pay gas fees for the token claim. He
// exchanges 2 XMR for 3 of Alice's ERC20 tokens.
func TestRunSwapDaemon_ExchangesXMRForERC20Tokens(t *testing.T) {
	minXMR := coins.StrToDecimal("1")
	maxXMR := coins.StrToDecimal("2")
	exRate := coins.StrToExchangeRate("1.5")

	bobConf := CreateTestConf(t, tests.GetMakerTestKey(t))
	monero.MineMinXMRBalance(t, bobConf.MoneroClient, coins.MoneroToPiconero(maxXMR))

	aliceConf := CreateTestConf(t, tests.GetTakerTestKey(t))

	tokenAddr := GetMockTokens(t, aliceConf.EthereumClient)[MockTether]
	tokenAsset := types.EthAsset(tokenAddr)

	timeout := 7 * time.Minute
	ctx := LaunchDaemons(t, timeout, aliceConf, bobConf)

	bc, err := wsclient.NewWsClient(ctx, fmt.Sprintf("ws://127.0.0.1:%d/ws", bobConf.RPCPort))
	require.NoError(t, err)
	ac, err := wsclient.NewWsClient(ctx, fmt.Sprintf("ws://127.0.0.1:%d/ws", aliceConf.RPCPort))
	require.NoError(t, err)

	_, bobStatusCh, err := bc.MakeOfferAndSubscribe(minXMR, maxXMR, exRate, tokenAsset, false)
	require.NoError(t, err)
	time.Sleep(250 * time.Millisecond) // offer propagation time

	// Have Alice query all the offer information back
	aRPC := rpcclient.NewClient(ctx, fmt.Sprintf("http://127.0.0.1:%d", aliceConf.RPCPort))
	peersWithOffers, err := aRPC.QueryAll(coins.ProvidesXMR, 3)
	require.NoError(t, err)
	require.Len(t, peersWithOffers, 1)
	require.Len(t, peersWithOffers[0].Offers, 1)
	peerID := peersWithOffers[0].PeerID
	offer := peersWithOffers[0].Offers[0]
	tokenInfo, err := aRPC.TokenInfo(offer.EthAsset.Address())
	require.NoError(t, err)
	providesAmt, err := exRate.ToERC20Amount(offer.MaxAmount, tokenInfo)
	require.NoError(t, err)

	aliceStatusCh, err := ac.TakeOfferAndSubscribe(peerID, offer.ID, providesAmt)
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
	// Check Bob's token balance via RPC method instead of doing it directly
	//
	bRPC := rpcclient.NewClient(ctx, fmt.Sprintf("http://127.0.0.1:%d", bobConf.RPCPort))
	balances, err := bRPC.Balances(&rpctypes.BalancesRequest{TokenAddrs: []ethcommon.Address{tokenAddr}})
	require.NoError(t, err)
	t.Logf("Balances: %#v", balances)

	require.NotEmpty(t, balances.TokenBalances)
	require.Equal(t, providesAmt.Text('f'), balances.TokenBalances[0].AsStandardString())
}
