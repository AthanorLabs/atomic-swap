package coins

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_roundToDecimalPlace(t *testing.T) {
	// Round half down
	amt := Str2Decimal("33.4999999999999999999999999999999999")
	err := roundToDecimalPlace(amt, 0)
	require.NoError(t, err)
	require.Equal(t, "33", amt.String())

	// Round half up
	amt = Str2Decimal("33.5")
	err = roundToDecimalPlace(amt, 0)
	require.NoError(t, err)
	require.Equal(t, "34", amt.String())

	// Round at decimal position
	amt = Str2Decimal("0.00009")
	err = roundToDecimalPlace(amt, 4)
	require.NoError(t, err)
	require.Equal(t, "0.0001", amt.String())
}
