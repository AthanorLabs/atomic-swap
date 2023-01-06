//go:build !prod

package coins

import (
	"math/big"

	"github.com/cockroachdb/apd/v3"
)

//
// FUNCTIONS ONLY FOR UNIT TESTS
//

// StrToDecimal converts strings to apd.Decimal for tests, panicking on error.
// This function is intended for use with string constants, so panic arguably
// correct and allows variables to be declared outside a test function.
func StrToDecimal(amount string) *apd.Decimal {
	a, _, err := new(apd.Decimal).SetString(amount)
	if err != nil {
		panic(err)
	}
	return a
}

// StrToExchangeRate converts strings to ExchangeRate for tests, panicking on error.
func StrToExchangeRate(rate string) *ExchangeRate {
	return ToExchangeRate(StrToDecimal(rate))
}

// IntToWei converts some amount of wei into an WeiAmount for unit tests.
func IntToWei(amount int64) *WeiAmount {
	return NewWeiAmount(big.NewInt(amount))
}
