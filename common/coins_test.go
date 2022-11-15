package common

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMoneroAmount(t *testing.T) {
	amount := float64(33.3)
	piconero := MoneroToPiconero(amount)
	require.Equal(t, fmt.Sprintf("%.11f", amount), fmt.Sprintf("%.11f", piconero.AsMonero()))

	amountUint := piconero.Uint64()
	amountUint2 := MoneroAmount(amountUint)
	require.Equal(t, amountUint, amountUint2.Uint64())
}

func TestEtherAmount(t *testing.T) {
	amount := float64(33.3)
	wei := EtherToWei(amount)
	require.Equal(t, fmt.Sprintf("%.18f", amount), fmt.Sprintf("%.18f", wei.AsEther()))

	amountUint := int64(8181)
	etherAmount := NewEtherAmount(amountUint)
	require.Equal(t, amountUint, etherAmount.BigInt().Int64())
}

func TestERC20TokenAmount(t *testing.T) {
	amount := float64(33.999999999)
	wei := NewERC20TokenAmountFromDecimals(amount, 9)
	require.Equal(t, fmt.Sprintf("%.9f", amount), fmt.Sprintf("%.9f", wei.AsStandard()))

	amountUint := int64(8181)
	tokenAmt := NewERC20TokenAmount(amountUint, 9)
	require.Equal(t, amountUint, tokenAmt.BigInt().Int64())
}
