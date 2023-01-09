package coins

import (
	"fmt"

	"github.com/cockroachdb/apd/v3"
)

// ValidatePositive is for doing additional input validation on apd.Decimal values
// that should only contain positive values (like min, max and provided amounts).
func ValidatePositive(jsonFieldName string, maxDecimals uint8, value *apd.Decimal) error {
	if value == nil {
		return fmt.Errorf("%q is not set", jsonFieldName)
	}
	if value.IsZero() {
		return fmt.Errorf("%q must be non-zero", jsonFieldName)
	}
	if value.Negative {
		return fmt.Errorf("%q cannot be negative", jsonFieldName)
	}

	// In most cases, this line won't do anything. If the coefficient is divisible
	// by one or more multiples of 10, the zeros are chopped off and added to the
	// exponent (same external value, but different internal representation).
	_, _ = value.Reduce(value)

	// We could probably go lower, but something is definitely suspicious if
	// someone is sending us amounts with more than 100 digits. The check
	// below does not differentiate between digits before/after the Decimal
	// point.
	numDigits := value.NumDigits() // number of digits in the coefficient
	if numDigits > MaxCoinPrecision {
		return fmt.Errorf("%q has too many digits", jsonFieldName)
	}

	// We are calling digits after the decimal point "decimals". Since we reduced
	// the value above, each negative exponent value represents a decimal point.
	if value.Exponent < -int32(maxDecimals) {
		return fmt.Errorf("%q has too many decimal points; found=%d max=%d", jsonFieldName,
			-value.Exponent, maxDecimals)
	}

	return nil
}
