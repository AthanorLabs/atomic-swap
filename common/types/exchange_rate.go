package types

import (
	"github.com/cockroachdb/apd/v3"
)

const (
	// MaxCoinPrecision is a somewhat arbitrary precision upper bound (2^256 consumes 78 digits)
	MaxCoinPrecision = 100
)

// ExchangeRate defines an exchange rate between ETH and XMR.
// It is defined as the ratio of ETH:XMR that the node wishes to provide.
// ie. an ExchangeRate of 0.1 means that the node considers 1 ETH = 10 XMR.
type ExchangeRate apd.Decimal

// ToExchangeRate casts an *apd.Decimal to *ExchangeRate
func ToExchangeRate(rate *apd.Decimal) *ExchangeRate {
	return (*ExchangeRate)(rate)
}

var decCtx = apd.BaseContext.WithPrecision(MaxCoinPrecision)

func (r *ExchangeRate) decimal() *apd.Decimal {
	return (*apd.Decimal)(r)
}

// UnmarshalText hands off JSON decoding to apd.Decimal
func (r *ExchangeRate) UnmarshalText(b []byte) error {
	return r.decimal().UnmarshalText(b)
}

// MarshalText hands off JSON encoding to apd.Decimal
func (r *ExchangeRate) MarshalText() ([]byte, error) {
	return r.decimal().MarshalText()
}

// ToXMR converts an ether amount to a monero amount with the given exchange rate
func (r *ExchangeRate) ToXMR(ethAmount *apd.Decimal) *apd.Decimal {
	xmrAmt := new(apd.Decimal)
	// TODO: return error? round?
	_, err := decCtx.Quo(xmrAmt, r.decimal(), ethAmount)
	if err != nil {
		panic(err)
	}
	return xmrAmt
}

// ToETH converts a monero amount to an eth amount with the given exchange rate
func (r *ExchangeRate) ToETH(xmrAmount *apd.Decimal) *apd.Decimal {
	ethAmt := new(apd.Decimal)
	// TODO: return error? round?
	// TODO: Min should round up, max should round down???
	_, err := decCtx.Mul(ethAmt, r.decimal(), xmrAmount)
	if err != nil {
		panic(err)
	}
	return ethAmt
}

func (r *ExchangeRate) String() string {
	if r == nil {
		return ""
	}
	return ((*apd.Decimal)(r)).String()
}
