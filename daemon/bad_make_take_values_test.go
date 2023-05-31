package daemon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/rpcclient"
	"github.com/athanorlabs/atomic-swap/tests"
)

// Tests end-to-end (client->swapd->client) make/take failures with currency
// precision issues.
func TestBadMakeTakeValues(t *testing.T) {
	bobConf := CreateTestConf(t, tests.GetMakerTestKey(t))
	monero.MineMinXMRBalance(t, bobConf.MoneroClient, coins.MoneroToPiconero(coins.StrToDecimal("10")))

	aliceConf := CreateTestConf(t, tests.GetTakerTestKey(t))

	timeout := 3 * time.Minute
	ctx, _ := LaunchDaemons(t, timeout, bobConf, aliceConf)

	bc := rpcclient.NewClient(ctx, bobConf.RPCPort)
	ac := rpcclient.NewClient(ctx, aliceConf.RPCPort)

	// Trigger a TakeOffer failure because the precision of the min/or max
	// value combined with the exchange rate would exceed the token's precision.
	// 14.979329 * 13.3 = 199.2250757 (7 digits of precision)
	minMaxXMRAmt := coins.StrToDecimal("14.979329")
	exRate := coins.StrToExchangeRate("13.3")
	mockTether := getMockTetherAsset(t, aliceConf.EthereumClient)
	expectedErr := `"net_makeOffer" failed: 14.979329 XMR * 13.3 exceeds token's 6 decimal precision`
	_, err := bc.MakeOffer(minMaxXMRAmt, minMaxXMRAmt, exRate, mockTether, false)
	require.ErrorContains(t, err, expectedErr)
	t.Log(err)

	// Now configure the MakeOffer to succeed, so we can fail some TakeOffer calls
	minXMRAmt := coins.StrToDecimal("1")
	maxXMRAmt := coins.StrToDecimal("10")
	providesAmt := coins.StrToDecimal("5.1234567") // 7 digits, max is 6
	makeResp, err := bc.MakeOffer(minXMRAmt, maxXMRAmt, exRate, mockTether, false)
	require.NoError(t, err)

	// Fail because providesAmount has too much precision in the token's standard units
	err = ac.TakeOffer(makeResp.PeerID, makeResp.OfferID, providesAmt)
	require.ErrorContains(t, err, `"net_takeOffer" failed: "providesAmount" has too many decimal points; found=7 max=6`)

	// Fail because the providesAmount has too much precision when converted into XMR
	// 20.123456/13.3 = 1.51304[180451127819548872] (bracketed sequence repeats forever)
	providesAmt = coins.StrToDecimal("20.123456")
	err = ac.TakeOffer(makeResp.PeerID, makeResp.OfferID, providesAmt)
	expectedErr = `"net_takeOffer" failed: 20.123456 "USDT" / 13.3 exceeds XMR's 12 decimal precision, try 20.123432`
	require.ErrorContains(t, err, expectedErr)
	t.Log(err)
}
