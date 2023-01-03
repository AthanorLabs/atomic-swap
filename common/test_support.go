//go:build !prod

package common

import (
	"github.com/cockroachdb/apd/v3"
)

//
// Functions only for tests
//

// Str2Decimal converts strings to big decimal for tests, panicing on error.
// This function is intended for use with string constants, where panic is
// an acceptable behavior.
func Str2Decimal(amount string) *apd.Decimal {
	a, _, err := new(apd.Decimal).SetString(amount)
	if err != nil {
		panic(err)
	}
	return a
}
