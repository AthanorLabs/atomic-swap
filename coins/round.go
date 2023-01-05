package coins

import (
	"github.com/cockroachdb/apd/v3"
)

func roundToDecimalPlace(result *apd.Decimal, n *apd.Decimal, decimalPlace uint8) error {
	result.Set(n) // already optimizes result == n

	// Adjust the exponent to the rounding place, round, then adjust the exponent back
	increaseExponent(result, decimalPlace)
	_, err := DecimalCtx.RoundToIntegralValue(result, result)
	if err != nil {
		return err
	}
	decreaseExponent(result, decimalPlace)
	_, _ = result.Reduce(result)
	return nil
}
