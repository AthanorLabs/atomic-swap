package coins

import (
	"fmt"

	"github.com/cockroachdb/apd/v3"
)

// ValidatePositive is for doing additional input validation on apd.Decimal values
// that should only contain positive values (like min, max and provided amounts).
func ValidatePositive(jsonFieldName string, value *apd.Decimal) error {
	if value == nil {
		return fmt.Errorf("%q is not set", jsonFieldName)
	}
	if value.IsZero() {
		return fmt.Errorf("%q must be non-zero", jsonFieldName)
	}
	if value.Negative {
		return fmt.Errorf("%q can not be negative", jsonFieldName)
	}
	// We could probably go lower, but something is definitely suspicious if
	// someone is sending us amounts with more than 100 digits. The check
	// below does not differentiate between digits before/after the Decimal
	// point.
	if value.NumDigits() > MaxCoinPrecision {
		return fmt.Errorf("%q is too large or precise", jsonFieldName)
	}

	return nil
}
