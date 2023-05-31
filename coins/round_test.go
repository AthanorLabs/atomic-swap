// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package coins

import (
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_roundToDecimalPlace(t *testing.T) {
	// Round half down
	amt := StrToDecimal("33.4999999999999999999999999999999999")
	amt, err := roundToDecimalPlace(amt, 0)
	require.NoError(t, err)
	assert.Equal(t, "33", amt.String())

	// Round half up
	amt = StrToDecimal("33.5")
	amt, err = roundToDecimalPlace(amt, 0)
	require.NoError(t, err)
	assert.Equal(t, "34", amt.String())

	// Round at Decimal position
	amt = StrToDecimal("0.00009")
	res, err := roundToDecimalPlace(amt, 4) // use a separate result variable this time
	require.NoError(t, err)
	assert.Equal(t, "0.0001", res.String())
	assert.Equal(t, "0.00009", amt.String()) // input value unchanged
}

func Test_exceedsMaxDigitsAfterDecimal(t *testing.T) {
	type testCase struct {
		val      *apd.Decimal
		decimals uint8
		exceeds  bool
	}

	testCases := []testCase{
		{
			val:      StrToDecimal("1234567890"),
			decimals: 0,
			exceeds:  false,
		},
		{
			val:      StrToDecimal("0.0000000000001"), // 13 decimal places
			decimals: 12,
			exceeds:  true,
		},
		{
			val:      StrToDecimal("123456789.999999"), // 6 decimal places
			decimals: 6,
			exceeds:  false,
		},
		{
			val:      StrToDecimal("123456789.999999"), // 6 decimal places
			decimals: 5,
			exceeds:  true,
		},
		{
			val:      StrToDecimal("123.123400000000000000"), // only 4 non-zero decimal places
			decimals: 4,
			exceeds:  false,
		},
	}

	for _, test := range testCases {
		r := ExceedsDecimals(test.val, test.decimals)
		require.Equal(t, test.exceeds, r)
	}
}

func TestExchangeRate_calcAltNumeratorAmount(t *testing.T) {
	type testCase struct {
		numeratorDecimals uint8
		resultDecimals    uint8
		numerator         *apd.Decimal
		denominator       *apd.Decimal
		expectedAltValue  *apd.Decimal
	}

	testCases := []*testCase{
		{
			//
			// 30/20.5 = 1.46341[46341...] (repeats forever, so won't fit in 12 decimals)
			//
			// In order to back calculate the closest numerator that will work
			// (fit within 6 decimals), we need to round the result at 5
			// decimals (6 + -1), where the -1 was contributed by the
			// denominator.
			//
			// 1.46341 * 20.5 = 29.999905
			//
			numeratorDecimals: 6,
			resultDecimals:    12,
			numerator:         StrToDecimal("30"),
			denominator:       StrToDecimal("20.5"),
			expectedAltValue:  StrToDecimal("29.999905"),
		},
		{
			//
			// Same as the previous example, but now the precision of the
			// numerator is 18 decimals, so we can round at the full precision
			// (12) of the result.
			//
			// 1.463414634146 *  20.5 = 29.999999999993
			numeratorDecimals: 18,
			resultDecimals:    12,
			numerator:         StrToDecimal("30"),
			denominator:       StrToDecimal("20.5"),
			expectedAltValue:  StrToDecimal("29.999999999993"),
		},
		{
			//
			// 200/300 = 0.666666666666... (repeats 6 forever)
			//
			// This example is interesting, because the denominator, when reduced,
			// is 3E2, allowing us to round the divided result at 8 decimal (2
			// more decimal places than the precision of the numerator).
			//
			// 0.666667   * 300 = 200.0001   (naive answer, rounding at the numerator's precision)
			// 0.66666667 * 300 = 200.000001 (true closest value)
			//
			numeratorDecimals: 6,
			resultDecimals:    12,
			numerator:         StrToDecimal("200"),
			denominator:       StrToDecimal("300"),
			expectedAltValue:  StrToDecimal("200.000001"),
		},
		{
			//
			// This is a case were we just return nil. The denominator has 6
			// decimal places (what we cap exchange rates at), but the numerator
			// is capped at 4 decimals. We can't just multiply any rounded
			// result by the denominator and get an alternate numerator that
			// fits in 4 decimals.
			//
			numeratorDecimals: 4,
			resultDecimals:    12,
			numerator:         StrToDecimal("0.1"),
			denominator:       StrToDecimal("0.333333"),
			expectedAltValue:  nil,
		},
	}

	for _, tc := range testCases {
		result := new(apd.Decimal)
		_, err := decimalCtx.Quo(result, tc.numerator, tc.denominator)
		require.NoError(t, err)
		altVal := calcAltNumeratorAmount(tc.numeratorDecimals, tc.resultDecimals, tc.denominator, result)
		if tc.expectedAltValue == nil {
			require.Nil(t, altVal)
		} else {
			require.Equal(t, tc.expectedAltValue.Text('f'), altVal.Text('f'))
		}
	}
}
