package coins

import (
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/require"
)

func Test_roundToDecimalPlace(t *testing.T) {
	// Round half down
	amt := Str2Decimal("33.4999999999999999999999999999999999")
	err := roundToDecimalPlace(amt, amt, 0)
	require.NoError(t, err)
	require.Equal(t, "33", amt.String())

	// Round half up
	amt = Str2Decimal("33.5")
	err = roundToDecimalPlace(amt, amt, 0)
	require.NoError(t, err)
	require.Equal(t, "34", amt.String())

	// Round at decimal position
	amt = Str2Decimal("0.00009")
	res := new(apd.Decimal) // use a separate result variable this time
	err = roundToDecimalPlace(res, amt, 4)
	require.NoError(t, err)
	require.Equal(t, "0.0001", res.String())
	require.Equal(t, "0.00009", amt.String()) // input value unchanged
}
