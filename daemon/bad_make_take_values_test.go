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

func TestBadMakeTakeValues(t *testing.T) {
	bobConf := CreateTestConf(t, tests.GetMakerTestKey(t))
	monero.MineMinXMRBalance(t, bobConf.MoneroClient, coins.MoneroToPiconero(coins.StrToDecimal("10")))

	aliceConf := CreateTestConf(t, tests.GetTakerTestKey(t))

	timeout := 7 * time.Minute
	ctx, _ := LaunchDaemons(t, timeout, bobConf, aliceConf)

	bc := rpcclient.NewClient(ctx, bobConf.RPCPort)
	ac := rpcclient.NewClient(ctx, aliceConf.RPCPort)
	_ = ac

	minMaxAmt := coins.StrToDecimal("14.979329")
	exRate := coins.StrToExchangeRate("13.3")

	tokenAsset := getMockTetherAsset(t, aliceConf.EthereumClient)

	_, err := bc.MakeOffer(minMaxAmt, minMaxAmt, exRate, tokenAsset, false)
	require.Error(t, err)
}
