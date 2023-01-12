// Package coins provides types, conversions and exchange calculations for dealing
// with cryptocurrency coin and ERC20 token representations.
package coins

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/cockroachdb/apd/v3"
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
	if err := roundToDecimalPlace(pnAmt, pnAmt, 0); err != nil {
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

// WeiAmount represents some amount of ether in the smallest denomination (wei)
type WeiAmount apd.Decimal

// NewWeiAmount converts the passed *big.Int representation of a
// wei amount to the WeiAmount type. The returned value is a copy
// with no references to the passed value.
func NewWeiAmount(amount *big.Int) *WeiAmount {
	a := new(apd.BigInt).SetMathBigInt(amount)
	return ToWeiAmount(apd.NewWithBigInt(a, 0))
}

// Decimal exists to reduce ugly casts
func (a *WeiAmount) Decimal() *apd.Decimal {
	return (*apd.Decimal)(a)
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

// ToWeiAmount casts an *apd.Decimal to *WeiAmount
func ToWeiAmount(wei *apd.Decimal) *WeiAmount {
	return (*WeiAmount)(wei)
}

// EtherToWei converts some amount of standard ether to an WeiAmount.
func EtherToWei(ethAmt *apd.Decimal) *WeiAmount {
	weiAmt := new(apd.Decimal).Set(ethAmt)
	increaseExponent(weiAmt, NumEtherDecimals)
	// We do input validation on provided amounts and prevent values with
	// more than 18 decimal places, so no rounding happens with such values
	// below.
	if err := roundToDecimalPlace(weiAmt, weiAmt, 0); err != nil {
		panic(err) // shouldn't be possible
	}
	return ToWeiAmount(weiAmt)
}

// BigInt returns the given WeiAmount as a *big.Int
func (a *WeiAmount) BigInt() *big.Int {
	// Passing Quantize(...) zero as the exponent sets the coefficient to a whole-number
	// wei value. Round-half-up is used by default. Assuming no rounding occurs, the
	// operation below is the opposite of Reduce(...) which lops off even factors of
	// 10 from the coefficient, placing them on the exponent.
	wholeWeiVal := new(apd.Decimal)
	cond, err := decimalCtx.Quantize(wholeWeiVal, a.Decimal(), 0)
	if err != nil {
		panic(err)
	}
	if cond.Rounded() {
		// We round when converting from Ether to Wei, so we shouldn't see this
		log.Warn("Converting WeiAmount=%s to big.Int required rounding", a)
	}
	return new(big.Int).SetBytes(wholeWeiVal.Coeff.Bytes())
}

// AsEther returns the wei amount as ether
func (a *WeiAmount) AsEther() *apd.Decimal {
	ether := new(apd.Decimal).Set(a.Decimal())
	decreaseExponent(ether, NumEtherDecimals)
	_, _ = ether.Reduce(ether)
	return ether
}

// AsStandard is an alias for AsEther, returning the wei amount as ether
func (a *WeiAmount) AsStandard() *apd.Decimal {
	return a.AsEther()
}

// String returns the wei amount as a base10 string
func (a *WeiAmount) String() string {
	return a.Decimal().Text('f')
}

// ERC20TokenAmount represents some amount of an ERC20 token in the smallest denomination
type ERC20TokenAmount struct {
	amount      *apd.Decimal
	numDecimals uint8 // num digits after the Decimal point needed for smallest denomination
}

// NewERC20TokenAmountFromBigInt converts some amount in the smallest token denomination into an ERC20TokenAmount.
func NewERC20TokenAmountFromBigInt(amount *big.Int, decimals uint8) *ERC20TokenAmount {
	asDecimal := new(apd.Decimal)
	asDecimal.Coeff.SetBytes(amount.Bytes())
	return &ERC20TokenAmount{
		amount:      asDecimal,
		numDecimals: decimals,
	}
}

// NewERC20TokenAmount converts some amount in the smallest token denomination into an ERC20TokenAmount.
func NewERC20TokenAmount(amount int64, decimals uint8) *ERC20TokenAmount {
	return &ERC20TokenAmount{
		amount:      apd.New(amount, 0),
		numDecimals: decimals,
	}
}

// NewERC20TokenAmountFromDecimals converts some amount of standard token in standard format
// to its smaller denomination.
// For example, if amount is 1.99 and decimals is 9, the resulting value stored
// is 1.99 * 10^9.
func NewERC20TokenAmountFromDecimals(amount *apd.Decimal, decimals uint8) *ERC20TokenAmount {
	adjusted := new(apd.Decimal).Set(amount)
	increaseExponent(adjusted, decimals)
	// If we are rejecting token amounts that have too many decimal places on input, rounding
	// below will never occur.
	if err := roundToDecimalPlace(adjusted, adjusted, 0); err != nil {
		panic(err) // this shouldn't be possible
	}
	return &ERC20TokenAmount{
		amount:      adjusted,
		numDecimals: decimals,
	}
}

// BigInt returns the ERC20TokenAmount as a *big.Int
func (a *ERC20TokenAmount) BigInt() *big.Int {
	wholeTokenUnits := new(apd.Decimal)
	cond, err := decimalCtx.Quantize(wholeTokenUnits, a.amount, 0)
	if err != nil {
		panic(err)
	}
	if cond.Rounded() {
		log.Warn("Converting ERC20TokenAmount=%s (digits=%d) to big.Int required rounding", a.amount, a.numDecimals)
	}
	return new(big.Int).SetBytes(wholeTokenUnits.Coeff.Bytes())
}

// AsStandard returns the amount in standard form
func (a *ERC20TokenAmount) AsStandard() *apd.Decimal {
	tokenAmt := new(apd.Decimal).Set(a.amount)
	decreaseExponent(tokenAmt, a.numDecimals)
	_, _ = tokenAmt.Reduce(tokenAmt)
	return tokenAmt
}

// String returns the ERC20TokenAmount as a base10 string
func (a *ERC20TokenAmount) String() string {
	return a.amount.Text('f')
}
