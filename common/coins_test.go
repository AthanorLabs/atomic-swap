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

func TestToDecimals(t *testing.T) {
	val := NewEtherAmount(123456)
	require.Equal(t, fmt.Sprint(val.ToDecimals(5)), "1.23456")
	val = NewEtherAmount(1234567890)
	require.Equal(t, fmt.Sprint(val.ToDecimals(6)), "1234.56789")
}
