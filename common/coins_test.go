package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPiconeroAmount(t *testing.T) {
	const preciseAmount = "666.666666666666666666" // 18 sixes after the Decimal
	const moneroAmount = "666.666666666667"        // 12 digits after Decimal saved
	const piconeroAmount = "666666666666667"       // 15 digits rounded

	amount := Str2Decimal(preciseAmount)
	piconero := MoneroToPiconero(amount)
	require.Equal(t, moneroAmount, piconero.AsMonero().String())
	require.Equal(t, piconeroAmount, piconero.String())
}

func TestWeiAmount(t *testing.T) {
	amount := Str2Decimal("33.3")
	wei := EtherToWei(amount)
	require.Equal(t, "33300000000000000000", wei.String())
	require.Equal(t, "33.3", wei.AsEther().String())

	amountUint := int64(8181)
	WeiAmount := NewWeiAmount(amountUint)
	require.Equal(t, amountUint, WeiAmount.BigInt().Int64())
}

func TestERC20TokenAmount(t *testing.T) {
	amount := Str2Decimal("33.999999999")
	wei := NewERC20TokenAmountFromDecimals(amount, 9)
	require.Equal(t, amount.String(), wei.AsStandard().String())

	amount = Str2Decimal("33.000000005")
	wei = NewERC20TokenAmountFromDecimals(amount, 9)
	require.Equal(t, "33.000000005", wei.AsStandard().String())

	amount = Str2Decimal("33.0000000005")
	wei = NewERC20TokenAmountFromDecimals(amount, 9)
	require.Equal(t, "33.000000001", wei.AsStandard().String())

	amount = Str2Decimal("999999999999999999.0000000005")
	wei = NewERC20TokenAmountFromDecimals(amount, 9)
	require.Equal(t, "999999999999999999.000000001", wei.AsStandard().String())

	amountUint := int64(8181)
	tokenAmt := NewERC20TokenAmount(amountUint, 9)
	require.Equal(t, amountUint, tokenAmt.BigInt().Int64())
}
