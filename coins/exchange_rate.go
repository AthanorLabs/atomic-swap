// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package coins

import (
	"errors"
	"fmt"

	"github.com/cockroachdb/apd/v3"
)

// ExchangeRate defines an exchange rate between ETH and XMR.
// It is defined as the ratio of ETH:XMR that the node wishes to provide.
// ie. an ExchangeRate of 0.1 means that the node considers 1 ETH = 10 XMR.
type ExchangeRate apd.Decimal

// CalcExchangeRate computes and returns an exchange rate using ETH and XRM prices. The
// price can be relative to USD, bitcoin or something else, but both values should be
// relative to the same alternate currency.
func CalcExchangeRate(xmrPrice *apd.Decimal, ethPrice *apd.Decimal) (*ExchangeRate, error) {
	rate := new(apd.Decimal)
	_, err := decimalCtx.Quo(rate, xmrPrice, ethPrice)
	if err != nil {
		return nil, err
	}
	if rate, err = roundToDecimalPlace(rate, MaxExchangeRateDecimals); err != nil {
		return nil, err
	}
	return ToExchangeRate(rate), nil
}

// ToExchangeRate casts an *apd.Decimal to *ExchangeRate
func ToExchangeRate(rate *apd.Decimal) *ExchangeRate {
	return (*ExchangeRate)(rate)
}

// Decimal casts *ExchangeRate to *apd.Decimal
func (r *ExchangeRate) Decimal() *apd.Decimal {
	return (*apd.Decimal)(r)
}

// UnmarshalText hands off JSON decoding to apd.Decimal
func (r *ExchangeRate) UnmarshalText(b []byte) error {
	err := r.Decimal().UnmarshalText(b)
	if err != nil {
		return err
	}
	return ValidatePositive("exchangeRate", MaxExchangeRateDecimals, r.Decimal())
}

// MarshalText hands off JSON encoding to apd.Decimal
func (r *ExchangeRate) MarshalText() ([]byte, error) {
	return r.Decimal().MarshalText()
}

// ToXMR converts an ETH amount to an XMR amount with the given exchange rate.
// If the calculated value would have fractional piconeros, an error is
// returned.
func (r *ExchangeRate) ToXMR(ethAssetAmt EthAssetAmount) (*apd.Decimal, error) {
	xmrAmt := new(apd.Decimal)
	_, err := decimalCtx.Quo(xmrAmt, ethAssetAmt.AsStd(), r.Decimal())
	if err != nil {
		return nil, err
	}

	if ExceedsDecimals(xmrAmt, NumMoneroDecimals) {
		errMsg := fmt.Sprintf(
			"%s %s / %s exceeds XMR's %d decimal precision",
			ethAssetAmt.AsStdString(), ethAssetAmt.StdSymbol(), r, NumMoneroDecimals,
		)
		suggestedAltAmt := calcAltNumeratorAmount(ethAssetAmt.NumStdDecimals(), NumMoneroDecimals, r.Decimal(), xmrAmt)
		if suggestedAltAmt != nil {
			errMsg = fmt.Sprintf("%s, try %s", errMsg, suggestedAltAmt.Text('f'))
		}
		return nil, errors.New(errMsg)
	}

	return xmrAmt, nil
}

// ToETH converts an XMR amount to an ETH amount with the given exchange rate.
// If the calculated result would have fractional wei, an error is returned.
func (r *ExchangeRate) ToETH(xmrAmount *apd.Decimal) (*apd.Decimal, error) {
	ethAmt := new(apd.Decimal)
	_, err := decimalCtx.Mul(ethAmt, xmrAmount, r.Decimal())
	if err != nil {
		return nil, err
	}

	// Assuming the xmrAmount was capped at 12 decimal places and the exchange
	// rate was capped at 6 decimal places, you can't generate more than 18
	// decimal places below, so the error below can't happen.
	if ExceedsDecimals(ethAmt, NumEtherDecimals) {
		err := fmt.Errorf("%s XMR * %s exceeds ETH's %d decimal precision",
			xmrAmount.Text('f'), r, NumEtherDecimals)
		return nil, err
	}

	return ethAmt, nil
}

// ToERC20Amount converts an XMR amount to a token amount in standard units with
// the given exchange rate. If the result requires more decimal places than the
// token allows, an error is returned.
func (r *ExchangeRate) ToERC20Amount(xmrAmount *apd.Decimal, token *ERC20TokenInfo) (*apd.Decimal, error) {
	erc20Amount := new(apd.Decimal)
	_, err := decimalCtx.Mul(erc20Amount, xmrAmount, r.Decimal())
	if err != nil {
		return nil, err
	}

	if ExceedsDecimals(erc20Amount, token.NumDecimals) {
		// We could have a suggested value to try, like we have in ToXMR(...),
		// but since this is multiplication and not division, the end user
		// probably doesn't need the hint.
		err := fmt.Errorf("%s XMR * %s exceeds token's %d decimal precision",
			xmrAmount.Text('f'), r, token.NumDecimals)
		return nil, err
	}

	return NewTokenAmountFromDecimals(erc20Amount, token).AsStd(), nil
}

func (r *ExchangeRate) String() string {
	return r.Decimal().Text('f')
}
