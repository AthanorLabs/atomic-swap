package daemon

import (
	"context"
	"sync"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/rpcclient"
	"github.com/athanorlabs/atomic-swap/tests"
)

// Tests the scenario where Bob has XMR and enough ETH to pay gas fees for the token claim. He
// exchanges 2 XMR for 3 of Alice's ERC20 tokens.
func TestRunSwapDaemon_ExchangesXMRForERC20Tokens(t *testing.T) {
	fundingEC := extethclient.CreateTestClient(t, tests.GetTakerTestKey(t))
	tokenAsset := getMockTetherAsset(t, fundingEC)
	tokenAddr := tokenAsset.Address()
	token, err := fundingEC.ERC20Info(context.Background(), tokenAddr)
	require.NoError(t, err)

	minXMR := coins.StrToDecimal("0.1")
	maxXMR := coins.StrToDecimal("0.25")
	exRate := coins.StrToExchangeRate("140")
	providesAmt := coins.NewEthAssetAmount(coins.StrToDecimal("33.999994"), token) // 33.999994 USDT / 140 = 0.2428571 XMR
	gasMoney := coins.EtherToWei(coins.StrToDecimal("0.1"))

	bobEthKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	bobConf := CreateTestConf(t, bobEthKey)
	monero.MineMinXMRBalance(t, bobConf.MoneroClient, coins.MoneroToPiconero(maxXMR))

	// Ensure that Alice has no tokens and definitely no pre-approval to spend
	// any of those tokens by giving her a brand-new ETH key.
	aliceEthKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	aliceConf := CreateTestConf(t, aliceEthKey)
	timeout := 7 * time.Minute

	// Fund Alice and Bob with a little ether for gas. Bob needs gas to claim,
	// as ERC20 token swaps cannot use a relayer.
	_, err = fundingEC.Transfer(context.Background(), aliceConf.EthereumClient.Address(), gasMoney, nil)
	require.NoError(t, err)
	_, err = fundingEC.Transfer(context.Background(), bobConf.EthereumClient.Address(), gasMoney, nil)
	require.NoError(t, err)

	// Fund Alice with the exact amount of token that she'll provide in the swap
	// with Bob. After the swap is over, her token balance should be exactly
	// zero.
	erc20Iface, err := contracts.NewIERC20(tokenAddr, fundingEC.Raw())
	require.NoError(t, err)
	txOpts, err := fundingEC.TxOpts(context.Background())
	require.NoError(t, err)
	_, err = erc20Iface.Transfer(txOpts, aliceConf.EthereumClient.Address(), providesAmt.BigInt())
	require.NoError(t, err)

	ctx, _ := LaunchDaemons(t, timeout, aliceConf, bobConf)

	bc := rpcclient.NewClient(ctx, bobConf.RPCPort)
	ac := rpcclient.NewClient(ctx, aliceConf.RPCPort)

	_, bobStatusCh, err := bc.MakeOfferAndSubscribe(minXMR, maxXMR, exRate, tokenAsset, false)
	require.NoError(t, err)
	time.Sleep(250 * time.Millisecond) // offer propagation time

	// Have Alice query all the offer information back
	peersWithOffers, err := ac.QueryAll(coins.ProvidesXMR, 3)
	require.NoError(t, err)
	require.Len(t, peersWithOffers, 1)
	require.Len(t, peersWithOffers[0].Offers, 1)
	peerID := peersWithOffers[0].PeerID
	offer := peersWithOffers[0].Offers[0]
	require.Equal(t, tokenAddr.String(), offer.EthAsset.Address().String())

	aliceStatusCh, err := ac.TakeOfferAndSubscribe(peerID, offer.ID, providesAmt.AsStd())
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
	// Check final token balances via RPC method instead of doing it directly on
	// the eth client. Bob should have exactly the provided amount and Alice's
	// token balance should now be zero.
	//
	balReq := &rpctypes.BalancesRequest{TokenAddrs: []ethcommon.Address{tokenAddr}}

	bobBal, err := bc.Balances(balReq)
	require.NoError(t, err)
	require.NotEmpty(t, bobBal.TokenBalances)
	require.Equal(t, providesAmt.AsStdString(), bobBal.TokenBalances[0].AsStdString())

	aliceBal, err := ac.Balances(balReq)
	require.NoError(t, err)
	require.NotEmpty(t, aliceBal.TokenBalances)
	require.Equal(t, "0", aliceBal.TokenBalances[0].AsStdString())
}
