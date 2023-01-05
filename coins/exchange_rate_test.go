package coins

import (
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strToDec(amount string) *apd.Decimal {
	a, _, err := new(apd.Decimal).SetString(amount)
	if err != nil {
		panic(err)
	}
	return a
}

func TestExchangeRate_ToXMR(t *testing.T) {
	rate := ToExchangeRate(strToDec("0.25")) // 4 XMR * 0.25 = 1 ETH
	ethAmount := strToDec("1")
	const expectedXMRAmount = "4"
	xmrAmount, err := rate.ToXMR(ethAmount)
	require.NoError(t, err)
	assert.Equal(t, expectedXMRAmount, xmrAmount.String())
}

func TestExchangeRate_ToXMR_roundDown(t *testing.T) {
	// XMR has 12 decimal points of accuracy, 1/3 below is represented with 14 decimal points
	rate := ToExchangeRate(strToDec("0.33333333333333")) // 9 XMR * 1/3 = 3 ETH
	ethAmount := strToDec("3")
	const expectedXMRAmount = "9"
	xmrAmount, err := rate.ToXMR(ethAmount)
	require.NoError(t, err)
	assert.Equal(t, expectedXMRAmount, xmrAmount.String())
}

func TestExchangeRate_ToXMR_roundUp(t *testing.T) {
	// XMR has 12 decimal points of accuracy, 2/3 below is represented with 14 decimal points.
	rate := ToExchangeRate(strToDec("0.66666666666666")) // 9 XMR * 2/3 = 6 ETH
	ethAmount := strToDec("6")
	const expectedXMRAmount = "9"
	xmrAmount, err := rate.ToXMR(ethAmount)
	require.NoError(t, err)
	assert.Equal(t, expectedXMRAmount, xmrAmount.String())
}

func TestExchangeRate_ToETH(t *testing.T) {
	rate := ToExchangeRate(strToDec("0.25")) // 4 XMR * 0.25 = 1 ETH
	xmrAmount := strToDec("4")
	const expectedETHAmount = "1"
	ethAmount, err := rate.ToETH(xmrAmount)
	require.NoError(t, err)
	assert.Equal(t, expectedETHAmount, ethAmount.String())
}

func TestExchangeRate_String(t *testing.T) {
	rate := ToExchangeRate(apd.New(3, -4)) // 0.0003
	assert.Equal(t, "0.0003", rate.String())
}
