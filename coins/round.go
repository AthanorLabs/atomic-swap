// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package coins

import (
	"github.com/cockroachdb/apd/v3"
	"github.com/ethereum/go-ethereum/common/math"
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

// calcAltNumeratorAmount attempts to find an alternative, close numerator
// amount that will not exceed the result's decimal precision when divided by
// the denominator and will also not exceed the numerator's precision. If
// no close approximation is found, nil is returned.
//
// Take this example, where the calculation below has already failed due to the
// computed XMR amount exceeding XMR's 12 decimal places.
//
//	TOKEN_AMOUNT / EXCHANGE_RATE = XMR_AMOUNT_TOO_PRECISE
//
// In the above example, we want to calculate the closest possible numerator
// (TOKEN_AMOUNT) by rounding the right-hand-side and multiplying by the
// denominator (EXCHANGE_RATE).
//
//	ROUND(XMR_AMOUNT_TOO_PRECISE, PRECISION) * EXCHANGE_RATE = SUGGESTED_TOKEN_AMOUNT
//	  or rewritten:
//	ROUND(unroundedResult, PRECISION) * denominator = SUGGESTED_ALTERNATE_NUMERATOR
//
// The trick is in computing that maximum value for PRECISION that will yield a
// valid result. Rounding too early increases the chance that we will compute a
// value that is outside of the offer's allowed range (which we don't have
// access to at this point in the code).
func calcAltNumeratorAmount(
	numeratorDecimals uint8,
	resultDecimals uint8,
	denominator *apd.Decimal,
	unroundedResult *apd.Decimal,
) *apd.Decimal {
	// Start with the max precision that the numerator will allow. Note that the
	// denominator's exponent can be positive (300 is 3E2) or negative (0.003 is
	// 3E-3). When it is positive, we can round the result at a larger precision
	// than the numerator's precision and still get a result that fits in the
	// numerator's precision.
	_, _ = denominator.Reduce(denominator)
	roundingPrecision := denominator.Exponent + int32(numeratorDecimals)
	if roundingPrecision < 0 || roundingPrecision > math.MaxUint8 {
		return nil
	}

	// Pick the smaller precision value (precision of result or the precision
	// needed to calculate a valid numerator).
	if roundingPrecision > int32(resultDecimals) {
		roundingPrecision = int32(resultDecimals)
	}

	roundedResult, err := roundToDecimalPlace(unroundedResult, uint8(roundingPrecision))
	if err != nil {
		return nil // not reachable
	}

	closestAltResult := new(apd.Decimal)
	_, err = decimalCtx.Mul(closestAltResult, roundedResult, denominator)
	if err != nil {
		return nil
	}

	_, _ = closestAltResult.Reduce(closestAltResult)

	if closestAltResult.IsZero() {
		return nil
	}

	return closestAltResult
}
