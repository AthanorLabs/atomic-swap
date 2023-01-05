package coins

import (
	"math"
	"testing"

	"github.com/cockroachdb/apd/v3"
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

func TestMoneroToPiconero(t *testing.T) {
	xrmAmount := Str2Decimal("2")
	const expectedPiconeros = "2000000000000"
	piconeroAmount := MoneroToPiconero(xrmAmount)
	require.Equal(t, expectedPiconeros, piconeroAmount.String())
}

func TestMoneroToPiconero_roundUp(t *testing.T) {
	//
	// This test is merely demonstrating the current behavior. It is not
	// entirely clear if the ideal behavior is to round-half-up, truncate,
	// or just store fractional piconeros.
	//
	xrmAmount := Str2Decimal("1.0000000000005") // 12 zeros, then "5"
	const expectedPiconeros = "1000000000001"
	piconeroAmount := MoneroToPiconero(xrmAmount)
	require.Equal(t, expectedPiconeros, piconeroAmount.String())
}

func TestMoneroToPiconero_roundDown(t *testing.T) {
	xrmAmount := Str2Decimal("1.00000000000049") // 12 zeros, then "49"
	const expectedPiconeros = "1000000000000"
	piconeroAmount := MoneroToPiconero(xrmAmount)
	require.Equal(t, expectedPiconeros, piconeroAmount.String())
}

func TestNewPiconeroAmount(t *testing.T) {
	onePn := NewPiconeroAmount(1)
	oneU64, err := onePn.Uint64()
	require.NoError(t, err)
	require.Equal(t, oneU64, uint64(1))
}

func TestPiconeroAmount_Uint64(t *testing.T) {
	// MaxUint64 should work
	piconeros := NewPiconeroAmount(math.MaxUint64)
	piconerosU64, err := piconeros.Uint64()
	require.NoError(t, err)
	require.Equal(t, uint64(math.MaxUint64), piconerosU64)

	// MaxUint64+1 should return an error
	one := apd.New(1, 0)
	_, err = decimalCtx.Add(piconeros.Decimal(), piconeros.Decimal(), one)
	require.NoError(t, err)
	_, err = piconeros.Uint64()
	require.ErrorContains(t, err, "value out of range")

	// Negative values, which we should never have, return an error
	piconeros.Decimal().Set(apd.New(-1, 0))
	_, err = piconeros.Uint64()
	require.ErrorContains(t, err, "can not convert")
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
