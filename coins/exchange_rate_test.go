// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package coins

import (
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExchangeRate_ToXMR(t *testing.T) {
	rate := StrToExchangeRate("0.25") // 4 XMR * 0.25 = 1 ETH
	ethAssetAmt := StrToETHAsset("1", nil)
	const expectedXMRAmount = "4"
	xmrAmount, err := rate.ToXMR(ethAssetAmt)
	require.NoError(t, err)
	assert.Equal(t, expectedXMRAmount, xmrAmount.String())
}

func TestExchangeRate_ToXMR_exceedsXMRPrecision(t *testing.T) {
	// 3.1/0.333333 calculated to 13 decimals is 9.3000093000093 (300009 repeats indefinitely)
	rate := StrToExchangeRate("0.333333")
	ethAssetAmt := StrToETHAsset("3.1", nil)

	_, err := rate.ToXMR(ethAssetAmt)
	expectedErr := "3.1 ETH / 0.333333 exceeds XMR's 12 decimal precision, try 3.099999999999899997"
	require.ErrorContains(t, err, expectedErr)

	// 6.6/0.666666 to 13 decimal places is 9.9000099000099 (900009 repeats indefinitely)
	rate = StrToExchangeRate("0.666666")
	ethAssetAmt = StrToETHAsset("6.6", nil)

	_, err = rate.ToXMR(ethAssetAmt)
	expectedErr = "6.6 ETH / 0.666666 exceeds XMR's 12 decimal precision, try 6.60000000000006666"
	require.ErrorContains(t, err, expectedErr)
}

func TestExchangeRate_ToXMR_fail(t *testing.T) {
	rateZero := ToExchangeRate(new(apd.Decimal)) // zero exchange rate
	_, err := rateZero.ToXMR(StrToETHAsset("0.1", nil))
	require.ErrorContains(t, err, "division by zero")
}

func TestExchangeRate_ToETH(t *testing.T) {
	rate := StrToExchangeRate("0.25") // 4 XMR * 0.25 = 1 ETH
	xmrAmount := StrToDecimal("4")
	const expectedETHAmount = "1"
	ethAmount, err := rate.ToETH(xmrAmount)
	require.NoError(t, err)
	assert.Equal(t, expectedETHAmount, ethAmount.String())
}

func TestExchangeRate_ToERC20Amount(t *testing.T) {
	rate := StrToExchangeRate("1.5") // 1.5 XMR * 2 = 3 Standard token units
	xmrAmount := StrToDecimal("2")
	const tokenDecimals = 10
	const expectedTokenStandardAmount = "3"
	erc20Info := &ERC20TokenInfo{NumDecimals: tokenDecimals}

	erc20Amt, err := rate.ToERC20Amount(xmrAmount, erc20Info)
	require.NoError(t, err)
	assert.Equal(t, expectedTokenStandardAmount, erc20Amt.Text('f'))
}

func TestExchangeRate_ToERC20Amount_exceedsTokenPrecision(t *testing.T) {
	const tokenDecimals = 6
	token := &ERC20TokenInfo{NumDecimals: tokenDecimals}

	// 1.0000015 * 0.333333 = 0.3333334999995
	xmrAmount := StrToDecimal("1.0000015")
	rate := StrToExchangeRate("0.333333")
	_, err := rate.ToERC20Amount(xmrAmount, token)
	require.ErrorContains(t, err, "1.0000015 XMR * 0.333333 exceeds token's 6 decimal precision")
}

func TestExchangeRate_String(t *testing.T) {
	rate := ToExchangeRate(apd.New(3, -4)) // 0.0003
	assert.Equal(t, "0.0003", rate.String())
}

func TestCalcExchangeRate(t *testing.T) {
	xmrPrice := StrToDecimal("200")
	ethPrice := StrToDecimal("300")
	rate, err := CalcExchangeRate(xmrPrice, ethPrice)
	require.NoError(t, err)
	assert.Equal(t, "0.666667", rate.String())
}

func TestCalcExchangeRate_fail(t *testing.T) {
	xmrPrice := StrToDecimal("1.0")
	ethPrice := StrToDecimal("0") // create a division by zero error
	_, err := CalcExchangeRate(xmrPrice, ethPrice)
	require.ErrorContains(t, err, "division by zero")
}
