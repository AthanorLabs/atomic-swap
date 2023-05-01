package xmrmaker

import (
	"context"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/tests"
)

func Test_validateMinBalance(t *testing.T) {
	ctx := context.Background()
	mc := monero.CreateWalletClient(t)
	ec := extethclient.CreateTestClient(t, tests.GetMakerTestKey(t))
	offerMax := coins.StrToDecimal("0.4")
	tokenAsset := types.EthAsset(ethcommon.Address{0x1}) // arbitrary token asset

	monero.MineMinXMRBalance(t, mc, coins.MoneroToPiconero(offerMax))

	err := validateMinBalance(ctx, mc, ec, offerMax, tokenAsset)
	require.NoError(t, err)
}

func Test_validateMinBalance_insufficientXMR(t *testing.T) {
	ctx := context.Background()
	ec := extethclient.CreateTestClient(t, tests.GetMakerTestKey(t))
	mc := monero.CreateWalletClient(t)
	offerMax := coins.StrToDecimal("0.5")

	// We didn't mine any XMR, so balance is zero

	err := validateMinBalance(ctx, mc, ec, offerMax, types.EthAssetETH)
	require.ErrorContains(t, err, "balance 0 XMR is too low for maximum offer amount of 0.5 XMR")
}

func Test_validateMinBalance_insufficientETH(t *testing.T) {
	ctx := context.Background()

	mc := monero.CreateWalletClient(t)
	pk, err := crypto.GenerateKey() // new eth key with no balance
	require.NoError(t, err)
	ec := extethclient.CreateTestClient(t, pk)

	offerMax := coins.StrToDecimal("0.5")
	tokenAsset := types.EthAsset(ethcommon.Address{0x1}) // arbitrary token asset

	monero.MineMinXMRBalance(t, mc, coins.MoneroToPiconero(offerMax))

	err = validateMinBalance(ctx, mc, ec, offerMax, tokenAsset)
	require.Error(t, err)
	require.Regexp(t, "balance of 0 ETH insufficient for token swap, 0.000\\d+ ETH required to claim", err.Error())
}
