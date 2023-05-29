// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package coins provides types, conversions and exchange calculations for dealing
// with cryptocurrency coin and ERC20 token representations.
package coins

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

// PiconeroAmount represents some amount of piconero (the smallest denomination of monero)
type PiconeroAmount apd.Decimal

// NewPiconeroAmount converts piconeros in uint64 to PicnoneroAmount
func NewPiconeroAmount(amount uint64) *PiconeroAmount {
	// apd.New(...) takes signed int64, so we need to use string initialization
	// to avoid error handling on values greater than 2^63-1.
	a, _, err := apd.NewFromString(fmt.Sprintf("%d", amount))
	if err != nil {
		panic(err) // can't happen, since we generated the string
	}
	return (*PiconeroAmount)(a)
}

// Decimal casts *PiconeroAmount to *apd.Decimal
func (a *PiconeroAmount) Decimal() *apd.Decimal {
	return (*apd.Decimal)(a)
}

// Uint64 converts piconero amount to uint64. Errors if a is negative or larger than 2^63-1.
func (a *PiconeroAmount) Uint64() (uint64, error) {
	// Hopefully, the rest of our code is doing input validation and the error below
	// never gets triggered.
	if a.Negative {
		return 0, fmt.Errorf("cannot convert %s to unsigned", a.String())
	}

	// Decimal has an Int64() method, but not a UInt64() method, so we are converting to
	// a string and back (optimizing for least code instead of speed).
	return strconv.ParseUint(a.String(), 10, 64)
}

// UnmarshalText hands off JSON decoding to apd.Decimal
func (a *PiconeroAmount) UnmarshalText(b []byte) error {
	err := a.Decimal().UnmarshalText(b)
	if err != nil {
		return err
	}
	if a.Negative {
		return errNegativePiconeros
	}
	return nil
}

// MarshalText hands off JSON encoding to apd.Decimal
func (a *PiconeroAmount) MarshalText() ([]byte, error) {
	return a.Decimal().MarshalText()
}

// MoneroToPiconero converts an amount in Monero to Piconero
func MoneroToPiconero(xmrAmt *apd.Decimal) *PiconeroAmount {
	pnAmt := new(apd.Decimal).Set(xmrAmt)
	increaseExponent(pnAmt, NumMoneroDecimals)
	// We do input validation and reject XMR values with more than 12 decimal
	// places from external sources, so no rounding will happen with those
	// values below.
	pnAmt, err := roundToDecimalPlace(pnAmt, 0)
	if err != nil {
		panic(err) // shouldn't be possible
	}
	return (*PiconeroAmount)(pnAmt)
}

// Cmp compares a and other and returns:
//
//	-1 if a < other
//	 0 if a == other
//	+1 if a > other
func (a *PiconeroAmount) Cmp(other *PiconeroAmount) int {
	return a.Decimal().Cmp(other.Decimal())
}

// CmpU64 compares a and other and returns:
//
//	-1 if a < other
//	 0 if a == other
//	+1 if a > other
func (a *PiconeroAmount) CmpU64(other uint64) int {
	return a.Cmp(NewPiconeroAmount(other))
}

// String returns the PiconeroAmount as a base10 string
func (a *PiconeroAmount) String() string {
	// If you call Decimal's String() method, it calls Text('G'), but
	// we'd rather 0.001 instead of 1E-3.
	return a.Decimal().Text('f')
}

// AsMonero converts the piconero PiconeroAmount into standard units
func (a *PiconeroAmount) AsMonero() *apd.Decimal {
	xmrAmt := new(apd.Decimal).Set(a.Decimal())
	decreaseExponent(xmrAmt, NumMoneroDecimals)
	_, _ = xmrAmt.Reduce(xmrAmt)
	return xmrAmt
}

// AsMoneroString converts a PiconeroAmount into a formatted XMR amount string.
func (a *PiconeroAmount) AsMoneroString() string {
	return a.AsMonero().Text('f')
}

// FmtPiconeroAsXMR takes piconeros as input and produces a formatted string of the
// amount in XMR.
func FmtPiconeroAsXMR(piconeros uint64) string {
	return NewPiconeroAmount(piconeros).AsMoneroString()
}

// EthAssetAmount represents an amount of an Ethereum asset (ie. ether or an ERC20)
type EthAssetAmount interface {
	BigInt() *big.Int
	AsStd() *apd.Decimal
	AsStdString() string
	StdSymbol() string
	IsToken() bool
	TokenAddress() ethcommon.Address
}

// NewEthAssetAmount accepts an amount, in standard units, for ETH or a token and
// returns a type implementing EthAssetAmount. If the token is nil, we assume
// the asset is ETH.
func NewEthAssetAmount(amount *apd.Decimal, token *ERC20TokenInfo) EthAssetAmount {
	if token == nil {
		return EtherToWei(amount)
	}
	return NewTokenAmountFromDecimals(amount, token)
}

// WeiAmount represents some amount of ETH in the smallest denomination (Wei)
type WeiAmount apd.Decimal

// NewWeiAmount converts the passed *big.Int representation of a
// Wei amount to the WeiAmount type. The returned value is a copy
// with no references to the passed value.
func NewWeiAmount(amount *big.Int) *WeiAmount {
	a := new(apd.BigInt).SetMathBigInt(amount)
	return ToWeiAmount(apd.NewWithBigInt(a, 0))
}

// Decimal exists to reduce ugly casts
func (a *WeiAmount) Decimal() *apd.Decimal {
	return (*apd.Decimal)(a)
}

// Cmp compares a and b and returns:
//
//	-1 if a <  b
//	 0 if a == b
//	+1 if a >  b
func (a *WeiAmount) Cmp(b *WeiAmount) int {
	return a.Decimal().Cmp(b.Decimal())
}

// UnmarshalText hands off JSON decoding to apd.Decimal
func (a *WeiAmount) UnmarshalText(b []byte) error {
	err := a.Decimal().UnmarshalText(b)
	if err != nil {
		return err
	}
	if a.Negative {
		return errNegativeWei
	}
	return nil
}

// MarshalText hands off JSON encoding to apd.Decimal
func (a *WeiAmount) MarshalText() ([]byte, error) {
	return a.Decimal().MarshalText()
}

// ToWeiAmount casts an *apd.Decimal that is already in Wei to *WeiAmount
func ToWeiAmount(wei *apd.Decimal) *WeiAmount {
	return (*WeiAmount)(wei)
}

// EtherToWei converts some amount of standard ETH to a WeiAmount.
func EtherToWei(ethAmt *apd.Decimal) *WeiAmount {
	weiAmt := new(apd.Decimal).Set(ethAmt)
	increaseExponent(weiAmt, NumEtherDecimals)
	// We do input validation on provided amounts and prevent values with
	// more than 18 decimal places, so no rounding happens with such values
	// below.
	weiAmt, err := roundToDecimalPlace(weiAmt, 0)
	if err != nil {
		panic(err) // shouldn't be possible
	}
	return ToWeiAmount(weiAmt)
}

// BigInt returns the given WeiAmount as a *big.Int
func (a *WeiAmount) BigInt() *big.Int {
	// Passing Quantize(...) zero as the exponent sets the coefficient to a whole-number
	// Wei value. Round-half-up is used by default. Assuming no rounding occurs, the
	// operation below is the opposite of Reduce(...) which lops off even factors of
	// 10 from the coefficient, placing them on the exponent.
	wholeWeiVal := new(apd.Decimal)
	cond, err := decimalCtx.Quantize(wholeWeiVal, a.Decimal(), 0)
	if err != nil {
		panic(err)
	}
	if cond.Inexact() {
		// We round when converting from ETH to Wei, so we shouldn't see this
		log.Warnf("converting WeiAmount=%s to big.Int required rounding", a.String())
	}
	return new(big.Int).SetBytes(wholeWeiVal.Coeff.Bytes())
}

// AsEther returns the Wei amount as ETH
func (a *WeiAmount) AsEther() *apd.Decimal {
	ether := new(apd.Decimal).Set(a.Decimal())
	decreaseExponent(ether, NumEtherDecimals)
	_, _ = ether.Reduce(ether)
	return ether
}

// AsEtherString converts the Wei amount to an ETH amount string
func (a *WeiAmount) AsEtherString() string {
	return a.AsEther().Text('f')
}

// AsStd is an alias for AsEther, returning the Wei amount as ETH
func (a *WeiAmount) AsStd() *apd.Decimal {
	return a.AsEther()
}

// AsStdString is an alias for AsEtherString, returning the Wei amount as
// an ETH string
func (a *WeiAmount) AsStdString() string {
	return a.AsEther().Text('f')
}

// StdSymbol returns the string "ETH"
func (a *WeiAmount) StdSymbol() string {
	return "ETH"
}

// IsToken returns false, as WeiAmount is not an ERC20 token
func (a *WeiAmount) IsToken() bool {
	return false
}

// TokenAddress returns the all-zero address as WeiAmount is not an ERC20 token
func (a *WeiAmount) TokenAddress() ethcommon.Address {
	return ethcommon.Address{}
}

// String returns the Wei amount as a base10 string
func (a *WeiAmount) String() string {
	return a.Decimal().Text('f')
}

// FmtWeiAsETH takes Wei as input and produces a formatted string of the amount
// in ETH.
func FmtWeiAsETH(wei *big.Int) string {
	return NewWeiAmount(wei).AsEther().Text('f')
}

// ERC20TokenInfo stores the token contract address and basic info that most
// ERC20 tokens support
type ERC20TokenInfo struct {
	Address     ethcommon.Address `json:"address" validate:"required"`
	NumDecimals uint8             `json:"decimals"` // digits after the Decimal point needed for smallest denomination
	Name        string            `json:"name"`
	Symbol      string            `json:"symbol"`
}

// NewERC20TokenInfo constructs and returns a new ERC20TokenInfo object
func NewERC20TokenInfo(address ethcommon.Address, decimals uint8, name string, symbol string) *ERC20TokenInfo {
	return &ERC20TokenInfo{
		Address:     address,
		NumDecimals: decimals,
		Name:        name,
		Symbol:      symbol,
	}
}

// SanitizedSymbol quotes the symbol ensuring escape sequences, newlines, etc. are escaped.
func (t *ERC20TokenInfo) SanitizedSymbol() string {
	return strconv.Quote(t.Symbol)
}

// ERC20TokenAmount represents some amount of an ERC20 token in the smallest denomination
type ERC20TokenAmount struct {
	Amount    *apd.Decimal    `json:"amount" validate:"required"` // in standard units
	TokenInfo *ERC20TokenInfo `json:"tokenInfo" validate:"required"`
}

// NewERC20TokenAmountFromBigInt converts some amount in the smallest token denomination
// into an ERC20TokenAmount.
func NewERC20TokenAmountFromBigInt(amount *big.Int, token *ERC20TokenInfo) *ERC20TokenAmount {
	asDecimal := new(apd.Decimal)
	asDecimal.Coeff.SetBytes(amount.Bytes())
	decreaseExponent(asDecimal, token.NumDecimals)
	_, _ = asDecimal.Reduce(asDecimal)

	return &ERC20TokenAmount{
		Amount:    asDecimal,
		TokenInfo: token,
	}
}

// NewTokenAmountFromDecimals converts an amount in standard units from
// apd.Decimal into the ERC20TokenAmount type. During the conversion, rounding
// may occur if the input value is too precise for the token's decimals.
func NewTokenAmountFromDecimals(amount *apd.Decimal, token *ERC20TokenInfo) *ERC20TokenAmount {
	if ExceedsDecimals(amount, token.NumDecimals) {
		log.Warn("Converting amount=%s (digits=%d) to token amount required rounding",
			amount.Text('f'), token.NumDecimals)
		roundedAmt, err := roundToDecimalPlace(amount, token.NumDecimals)
		if err != nil {
			panic(err) // shouldn't be possible
		}
		amount = roundedAmt
	}

	return &ERC20TokenAmount{
		Amount:    amount,
		TokenInfo: token,
	}
}

// BigInt returns the ERC20TokenAmount as a *big.Int
func (a *ERC20TokenAmount) BigInt() *big.Int {
	wholeTokenUnits := new(apd.Decimal).Set(a.Amount)
	increaseExponent(wholeTokenUnits, a.TokenInfo.NumDecimals)
	cond, err := decimalCtx.Quantize(wholeTokenUnits, wholeTokenUnits, 0)
	if err != nil {
		panic(err)
	}
	if cond.Inexact() {
		log.Warn("Converting ERC20TokenAmount=%s (digits=%d) to big.Int required rounding",
			a.Amount, a.TokenInfo.NumDecimals)
	}

	return new(big.Int).SetBytes(wholeTokenUnits.Coeff.Bytes())
}

// AsStd returns the amount in standard units
func (a *ERC20TokenAmount) AsStd() *apd.Decimal {
	return a.Amount
}

// AsStdString returns the ERC20TokenAmount as a base10 string in standard units.
func (a *ERC20TokenAmount) AsStdString() string {
	return a.String()
}

// StdSymbol returns the token's symbol in a format that is safe to log and display
func (a *ERC20TokenAmount) StdSymbol() string {
	return a.TokenInfo.SanitizedSymbol()
}

// IsToken returns true, because ERC20TokenAmount represents and ERC20 token
func (a *ERC20TokenAmount) IsToken() bool {
	return true
}

// TokenAddress returns the ERC20 token's ethereum contract address
func (a *ERC20TokenAmount) TokenAddress() ethcommon.Address {
	return a.TokenInfo.Address
}

// String returns the ERC20TokenAmount as a base10 string in standard units.
func (a *ERC20TokenAmount) String() string {
	return a.Amount.Text('f')
}
