package coins

import (
	"github.com/cockroachdb/apd/v3"
)

func roundToDecimalPlace(n *apd.Decimal, decimalPlace uint8) error {
	// Adjust the exponent to the rounding place, round, then adjust the exponent back
	increaseExponent(n, decimalPlace)
	_, err := DecimalCtx.RoundToIntegralValue(n, n)
	if err != nil {
		return err
	}
	decreaseExponent(n, decimalPlace)
	_, _ = n.Reduce(n)
	return nil
}
