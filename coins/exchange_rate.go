package coins

import (
	"github.com/cockroachdb/apd/v3"
)

// ExchangeRate defines an exchange rate between ETH and XMR.
// It is defined as the ratio of ETH:XMR that the node wishes to provide.
// ie. an ExchangeRate of 0.1 means that the node considers 1 ETH = 10 XMR.
type ExchangeRate apd.Decimal

// ToExchangeRate casts an *apd.Decimal to *ExchangeRate
func ToExchangeRate(rate *apd.Decimal) *ExchangeRate {
	return (*ExchangeRate)(rate)
}

func (r *ExchangeRate) decimal() *apd.Decimal {
	return (*apd.Decimal)(r)
}

// UnmarshalText hands off JSON decoding to apd.Decimal
func (r *ExchangeRate) UnmarshalText(b []byte) error {
	err := r.decimal().UnmarshalText(b)
	if err != nil {
		return err
	}
	if r.Negative {
		return errNegativeRate
	}
	return nil
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
	if err = roundToDecimalPlace(xmrAmt, xmrAmt, NumMoneroDecimals); err != nil {
		return nil, err
	}
	return xmrAmt, nil
}

// ToETH converts a monero amount to an eth amount with the given exchange rate
func (r *ExchangeRate) ToETH(xmrAmount *apd.Decimal) (*apd.Decimal, error) {
	ethAmt := new(apd.Decimal)
	_, err := decimalCtx.Mul(ethAmt, r.decimal(), xmrAmount)
	if err != nil {
		return nil, err
	}
	if err = roundToDecimalPlace(ethAmt, ethAmt, NumEtherDecimals); err != nil {
		return nil, err
	}
	return ethAmt, nil
}

func (r *ExchangeRate) String() string {
	return r.decimal().Text('f')
}
