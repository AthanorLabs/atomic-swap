package types

import (
	"github.com/cockroachdb/apd/v3"
)

const (
	// NumEtherDecimals is the number of Decimal points needed to represent whole units of Wei in Ether
	NumEtherDecimals = 18
	// NumMoneroDecimals is the number of Decimal points needed to represent whole units of piconero in XMR
	NumMoneroDecimals = 12

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

var decimalCtx = apd.BaseContext.WithPrecision(MaxCoinPrecision)

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
func (r *ExchangeRate) ToXMR(ethAmount *apd.Decimal) (*apd.Decimal, error) {
	xmrAmt := new(apd.Decimal)
	_, err := decimalCtx.Quo(xmrAmt, ethAmount, r.decimal())
	if err != nil {
		return nil, err
	}
	return roundToDecimalPlace(xmrAmt, NumMoneroDecimals)
}

// ToETH converts a monero amount to an eth amount with the given exchange rate
func (r *ExchangeRate) ToETH(xmrAmount *apd.Decimal) (*apd.Decimal, error) {
	ethAmt := new(apd.Decimal)
	_, err := decimalCtx.Mul(ethAmt, r.decimal(), xmrAmount)
	if err != nil {
		return nil, err
	}
	return roundToDecimalPlace(ethAmt, NumEtherDecimals)
}

func roundToDecimalPlace(n *apd.Decimal, decimalPlace int32) (*apd.Decimal, error) {
	rounded := new(apd.Decimal).Set(n)

	// Adjust the exponent to the rounding place, round, then adjust the exponent back
	rounded.Exponent += decimalPlace
	_, err := decimalCtx.RoundToIntegralValue(rounded, rounded)
	if err != nil {
		return nil, err
	}
	rounded.Exponent -= decimalPlace
	_, _ = rounded.Reduce(rounded)
	return rounded, nil
}

func (r *ExchangeRate) String() string {
	if r == nil {
		return ""
	}
	return r.decimal().Text('f')
}
