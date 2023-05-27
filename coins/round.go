// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package coins

import (
	"github.com/cockroachdb/apd/v3"
)

func roundToDecimalPlace(result *apd.Decimal, n *apd.Decimal, decimalPlace uint8) error {
	result.Set(n) // already optimizes result == n

	// Adjust the exponent to the rounding place, round, then adjust the exponent back
	increaseExponent(result, decimalPlace)
	_, err := decimalCtx.RoundToIntegralValue(result, result)
	if err != nil {
		return err
	}
	decreaseExponent(result, decimalPlace)
	_, _ = result.Reduce(result)
	return nil
}

// ExceedsDecimals returns `true` if the the number, written without an
// exponent, would require more digits after the decimal place than the passed
// value `decimals`. Otherwise, `false` is returned.
func ExceedsDecimals(val *apd.Decimal, maxDecimals uint8) bool {
	// Reduce strips trailing zeros from the coefficient and subtracts them from
	// the exponent. If the exponent is more negative than -`decimals`, we would
	// require more digits after the decimal point than we have available to
	// represent it.
	_, _ = val.Reduce(val)
	return val.Exponent < -int32(maxDecimals)
}
