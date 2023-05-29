// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package coins

import (
	"github.com/cockroachdb/apd/v3"
)

func roundToDecimalPlace(n *apd.Decimal, decimalPlace uint8) (*apd.Decimal, error) {
	result := new(apd.Decimal).Set(n)

	// Adjust the exponent to the rounding place, round, then adjust the exponent back
	increaseExponent(result, decimalPlace)
	_, err := decimalCtx.RoundToIntegralValue(result, result)
	if err != nil {
		return nil, err
	}
	decreaseExponent(result, decimalPlace)
	_, _ = result.Reduce(result)

	return result, nil
}

// ExceedsDecimals returns `true` if the the number, written without an
// exponent, would require more digits after the decimal place than the passed
// value `decimals`. Otherwise, `false` is returned.
func ExceedsDecimals(val *apd.Decimal, maxDecimals uint8) bool {
	return NumDecimals(val) > int32(maxDecimals)
}

// NumDecimals returns the minimum number digits needed to represent the passed
// value after the decimal point.
func NumDecimals(value *apd.Decimal) int32 {
	// Transfer any rightmost digits in the coefficient to the exponent
	_, _ = value.Reduce(value)

	if value.Exponent >= 0 {
		return 0
	}

	// The number of base-10 digits that we need to shift left by, is the number
	// of digits needed to represent the value after the decimal point.
	return -value.Exponent
}
