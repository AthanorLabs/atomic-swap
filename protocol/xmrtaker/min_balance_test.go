package xmrtaker

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/tests"
)

func deployTestToken(t *testing.T, ownerKey *ecdsa.PrivateKey) (ethcommon.Address, *coins.ERC20TokenInfo) {
	ctx := context.Background()
	ec := extethclient.CreateTestClient(t, ownerKey)

	txOpts, err := ec.TxOpts(ctx)
	require.NoError(t, err)

	tokenAddr, tx, _, err := contracts.DeployTestERC20(
		txOpts,
		ec.Raw(),
		"Min Max Testing Token",
		"MMTT",
		6,
		ec.Address(),
		big.NewInt(1000e6), // 1000 standard units
	)
	require.NoError(t, err)
	_, err = block.WaitForReceipt(ctx, ec.Raw(), tx.Hash())
	require.NoError(t, err)

	tokenInfo, err := ec.ERC20Info(ctx, tokenAddr)
	require.NoError(t, err)

	return tokenAddr, tokenInfo
}

func Test_validateMinBalForETHSwap(t *testing.T) {
	balanceWei := coins.EtherToWei(coins.StrToDecimal("0.5"))
	providesAmt := coins.StrToDecimal("0.25")
	gasPriceWei := big.NewInt(35e9) // 35 GWei
	err := validateMinBalForETHSwap(balanceWei, providesAmt, gasPriceWei)
	require.NoError(t, err)
}

func Test_validateMinBalForETHSwap_InsufficientBalance(t *testing.T) {
	balanceWei := coins.EtherToWei(coins.StrToDecimal("0.5"))
	providesAmt := balanceWei.AsEther() // provided amount consumes whole balance leaving nothing for gas
	gasPriceWei := big.NewInt(35e9)     // 35 GWei
	err := validateMinBalForETHSwap(balanceWei, providesAmt, gasPriceWei)
	// Amount in check below is truncated, so minor adjustments in the
	// expected gas won't break the test. The full message looks like:
	// "balance of 0.5 ETH is under required amount of 0.504541005 ETH"
	require.ErrorContains(t, err, "balance of 0.5 ETH is under required amount of 0.504")
}

func Test_validateMinBalForTokenSwap(t *testing.T) {
	tokenInfo := &coins.ERC20TokenInfo{
		Address:     ethcommon.Address{0x1},
		NumDecimals: 6,
		Name:        "Token",
		Symbol:      "TK",
	}
	balanceWei := coins.EtherToWei(coins.StrToDecimal("0.5"))
	tokenBalance := coins.NewTokenAmountFromDecimals(coins.StrToDecimal("10"), tokenInfo)
	providesAmt := coins.StrToDecimal("5")
	gasPriceWei := big.NewInt(35e9) // 35 GWei
	err := validateMinBalForTokenSwap(balanceWei, tokenBalance, providesAmt, gasPriceWei)
	require.NoError(t, err)
}

func Test_validateMinBalForTokenSwap_InsufficientTokenBalance(t *testing.T) {
	tokenInfo := &coins.ERC20TokenInfo{
		Address:     ethcommon.Address{0x1},
		NumDecimals: 6,
		Name:        "Token",
		Symbol:      "TK",
	}
	balanceWei := coins.EtherToWei(coins.StrToDecimal("0.5"))
	tokenBalance := coins.NewTokenAmountFromDecimals(coins.StrToDecimal("10"), tokenInfo)
	providesAmt := coins.StrToDecimal("20")
	gasPriceWei := big.NewInt(35e9) // 35 GWei
	err := validateMinBalForTokenSwap(balanceWei, tokenBalance, providesAmt, gasPriceWei)
	require.ErrorContains(t, err, `balance of 10 "TK" is below provided 20 "TK"`)
}

func Test_validateMinBalForTokenSwap_InsufficientETHBalance(t *testing.T) {
	tokenInfo := &coins.ERC20TokenInfo{
		Address:     ethcommon.Address{0x1},
		NumDecimals: 6,
		Name:        "Token",
		Symbol:      "TK",
	}
	balanceWei := coins.EtherToWei(coins.StrToDecimal("0.007"))
	tokenBalance := coins.NewTokenAmountFromDecimals(coins.StrToDecimal("10"), tokenInfo)
	providesAmt := coins.StrToDecimal("1")
	gasPriceWei := big.NewInt(35e9) // 35 GWei
	err := validateMinBalForTokenSwap(balanceWei, tokenBalance, providesAmt, gasPriceWei)
	// Amount in check below is truncated, so minor adjustments in the
	// expected gas won't break the test. The full message looks like:
	// "balance of 0.007 ETH is under required amount of 0.00743302 ETH"
	require.ErrorContains(t, err, `balance of 0.007 ETH is under required amount of 0.0074`)
}

func Test_validateMinBalanceETH(t *testing.T) {
	ctx := context.Background()
	pk := tests.GetTakerTestKey(t)
	ec := extethclient.CreateTestClient(t, pk)
	providesAmt := coins.StrToDecimal("0.3") // ganache taker key will have more than this
	err := validateMinBalance(ctx, ec, providesAmt, types.EthAssetETH)
	require.NoError(t, err)
}

func Test_validateMinBalanceToken(t *testing.T) {
	ctx := context.Background()
	pk := tests.GetTakerTestKey(t)
	ec := extethclient.CreateTestClient(t, pk)
	tokenAddr, _ := deployTestToken(t, pk)
	providesAmt := coins.StrToDecimal("1") // less than the 1k tokens belonging to the taker
	err := validateMinBalance(ctx, ec, providesAmt, types.EthAsset(tokenAddr))
	require.NoError(t, err)
}
