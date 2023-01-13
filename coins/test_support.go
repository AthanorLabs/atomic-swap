//go:build !prod

package coins

import (
	"fmt"
	"math/big"

	"github.com/cockroachdb/apd/v3"
)

//
// FUNCTIONS ONLY FOR UNIT TESTS
//

// StrToDecimal converts strings to apd.Decimal for tests, panicking on error.
// This function is intended for use with string constants, so panic is arguably
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
	r := new(ExchangeRate)
	if err := r.UnmarshalText([]byte(rate)); err != nil {
		panic(err) // test only function
	}
	return r
}

// IntToWei converts some amount of wei into an WeiAmount for unit tests.
func IntToWei(amount int64) *WeiAmount {
	if amount < 0 {
		panic(fmt.Sprintf("Wei amount %d is negative", amount)) // test only function
	}
	return NewWeiAmount(big.NewInt(amount))
}
