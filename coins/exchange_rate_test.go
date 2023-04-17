// Copyright 2023 Athanor Labs (ON)
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
	ethAmount := StrToDecimal("1")
	const expectedXMRAmount = "4"
	xmrAmount, err := rate.ToXMR(ethAmount)
	require.NoError(t, err)
	assert.Equal(t, expectedXMRAmount, xmrAmount.String())
}

func TestExchangeRate_ToXMR_roundDown(t *testing.T) {
	rate := StrToExchangeRate("0.333333")
	ethAmount := StrToDecimal("3.1")

	// 3.1/0.333333 calculated to 13 decimals is 9.3000093000093 (300009 repeats indefinitely)
	// This calculator goes to 200 decimals: https://www.mathsisfun.com/calculator-precision.html
	// XMR rounds at 12 decimal places to:
	const expectedXMRAmount = "9.300009300009"

	xmrAmount, err := rate.ToXMR(ethAmount)
	require.NoError(t, err)
	assert.Equal(t, expectedXMRAmount, xmrAmount.String())
}

func TestExchangeRate_ToXMR_roundUp(t *testing.T) {
	rate := StrToExchangeRate("0.666666")
	ethAmount := StrToDecimal("6.6")
	// 6.6/0.666666 to 13 decimal places is 9.9000099000099 (900009 repeats indefinitely)
	// The 9 in the 12th position goes to zero changing 11th position to 1:
	const expectedXMRAmount = "9.90000990001" // only 11 decimal places shown as 12th is 0
	xmrAmount, err := rate.ToXMR(ethAmount)
	require.NoError(t, err)
	assert.Equal(t, expectedXMRAmount, xmrAmount.String())
}

func TestExchangeRate_ToXMR_fail(t *testing.T) {
	rateZero := ToExchangeRate(new(apd.Decimal)) // zero exchange rate
	_, err := rateZero.ToXMR(StrToDecimal("0.1"))
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

func TestExchangeRate_ToERC20Amount_roundDown(t *testing.T) {
	// 0.333333 * 1.0000015 = 0.333333499...
	//                      = 0.333333 (token only supports 6 decimals)
	rate := StrToExchangeRate("0.333333")
	xmrAmount := StrToDecimal("1.0000015")

	const tokenDecimals = 6
	const expectedTokenStandardAmount = "0.333333"
	erc20Info := &ERC20TokenInfo{NumDecimals: tokenDecimals}

	erc20Amt, err := rate.ToERC20Amount(xmrAmount, erc20Info)
	require.NoError(t, err)
	assert.Equal(t, expectedTokenStandardAmount, erc20Amt.Text('f'))
}

func TestExchangeRate_ToERC20Amount_roundUp(t *testing.T) {
	// 0.333333 * 1.000001501 = 0.333333500..
	//                        = 0.333334 (token only supports 6 decimals)

	rate := StrToExchangeRate("0.333333")
	xmrAmount := StrToDecimal("1.000001501")

	const tokenDecimals = 6
	const expectedTokenStandardAmount = "0.333334"
	erc20Info := &ERC20TokenInfo{NumDecimals: tokenDecimals}

	erc20Amt, err := rate.ToERC20Amount(xmrAmount, erc20Info)
	require.NoError(t, err)
	assert.Equal(t, expectedTokenStandardAmount, erc20Amt.Text('f'))
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
