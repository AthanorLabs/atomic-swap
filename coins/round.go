package coins

import (
	"github.com/cockroachdb/apd/v3"
)

func roundToDecimalPlace(n *apd.Decimal, decimalPlace int32) (*apd.Decimal, error) {
	rounded := new(apd.Decimal).Set(n)

	// Adjust the exponent to the rounding place, round, then adjust the exponent back
	rounded.Exponent += decimalPlace
	_, err := DecimalCtx.RoundToIntegralValue(rounded, rounded)
	if err != nil {
		return nil, err
	}
	rounded.Exponent -= decimalPlace
	_, _ = rounded.Reduce(rounded)
	return rounded, nil
}
