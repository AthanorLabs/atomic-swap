package common

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPiconeroAmount(t *testing.T) {
	amount := float64(33.3)
	piconero := MoneroToPiconero(amount)
	require.Equal(t, fmt.Sprintf("%.11f", amount), fmt.Sprintf("%.11f", piconero.AsMonero()))

	amountUint := piconero.Uint64()
	amountUint2 := PiconeroAmount(amountUint)
	require.Equal(t, amountUint, amountUint2.Uint64())
}

func TestWeiAmount(t *testing.T) {
	amount := float64(33.3)
	wei := EtherToWei(amount)
	require.Equal(t, fmt.Sprintf("%.18f", amount), fmt.Sprintf("%.18f", wei.AsEther()))

	amountUint := int64(8181)
	WeiAmount := NewWeiAmount(amountUint)
	require.Equal(t, amountUint, WeiAmount.BigInt().Int64())
}

func TestERC20TokenAmount(t *testing.T) {
	amount := float64(33.999999999)
	wei := NewERC20TokenAmountFromDecimals(amount, 9)
	require.Equal(t, fmt.Sprintf("%.9f", amount), fmt.Sprintf("%.9f", wei.AsStandard()))

	amount = float64(33.000000005)
	wei = NewERC20TokenAmountFromDecimals(amount, 9)
	require.Equal(t, fmt.Sprintf("%.9f", amount), fmt.Sprintf("%.9f", wei.AsStandard()))

	amount = float64(33.0000000005)
	wei = NewERC20TokenAmountFromDecimals(amount, 9)
	require.Equal(t, fmt.Sprintf("%.9f", amount), fmt.Sprintf("%.9f", wei.AsStandard()))

	amount = float64(999999999999999999.0000000005)
	wei = NewERC20TokenAmountFromDecimals(amount, 9)
	require.Equal(t, fmt.Sprintf("%.9f", amount), fmt.Sprintf("%.9f", wei.AsStandard()))

	amountUint := int64(8181)
	tokenAmt := NewERC20TokenAmount(amountUint, 9)
	require.Equal(t, amountUint, tokenAmt.BigInt().Int64())
}

func TestRound(t *testing.T) {
	amt := big.NewFloat(33.49)
	expected := big.NewInt(33)
	require.Equal(t, expected, round(amt))

	amt = big.NewFloat(33.5)
	expected = big.NewInt(34)
	require.Equal(t, expected, round(amt))

	amt = big.NewFloat(33.51)
	expected = big.NewInt(34)
	require.Equal(t, expected, round(amt))
}
