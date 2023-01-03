package common

import (
	"fmt"
	"math/big"

	"github.com/cockroachdb/apd/v3"
	"github.com/ethereum/go-ethereum/log"
)

const (
	// NumEtherDecimals is the number of Decimal points needed to represent whole units of Wei in Ether
	NumEtherDecimals = 18
	// NumMoneroDecimals is the number of Decimal points needed to represent whole units of piconero in XMR
	NumMoneroDecimals = 12

	// MaxCoinPrecision is a somewhat arbitrary precision upper bound (2^256 consumes 78 digits)
	MaxCoinPrecision = 100
)

var (
	// DecimalCtx is the apd context used for math operations on our coins
	DecimalCtx          = apd.BaseContext.WithPrecision(MaxCoinPrecision)
	nonFractionalWeiCtx = apd.BaseContext.WithPrecision(NumEtherDecimals)
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
	// 2^63 / 10^12 is over 9 million XMR, so there is no point in adding complicated code to
	// handle all 64 bits.
	amtInt64, err := a.Decimal().Int64()
	if err != nil {
		return 0, err
	}
	// Hopefully, the rest of our code is doing input validation and the error below
	// never gets triggered.
	if amtInt64 < 0 {
		return 0, fmt.Errorf("can not convert %d to unsigned value", amtInt64)
	}
	return uint64(amtInt64), nil
}

// UnmarshalText hands off JSON decoding to apd.Decimal
func (a *PiconeroAmount) UnmarshalText(b []byte) error {
	return a.Decimal().UnmarshalText(b)
}

// MarshalText hands off JSON encoding to apd.Decimal
func (a *PiconeroAmount) MarshalText() ([]byte, error) {
	return a.Decimal().MarshalText()
}

// MoneroToPiconero converts an amount in Monero to Piconero
func MoneroToPiconero(xmrAmt *apd.Decimal) *PiconeroAmount {
	pnAmt := new(apd.Decimal).Set(xmrAmt)
	pnAmt.Exponent += NumMoneroDecimals
	_, err := DecimalCtx.Round(pnAmt, pnAmt)
	if err != nil {
		// This could only happen on over or underflow. We vet the ranges of all external
		// inputs, so this cannot happen.
		panic(err)
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
	return a.Decimal().String()
}

// AsMonero converts the piconero PiconeroAmount into standard units
func (a *PiconeroAmount) AsMonero() *apd.Decimal {
	xmrAmt := new(apd.Decimal).Set(a.Decimal())
	xmrAmt.Exponent -= NumMoneroDecimals
	_, _ = xmrAmt.Reduce(xmrAmt)
	return xmrAmt
}

// WeiAmount represents some amount of ether in the smallest denomination (wei)
type WeiAmount apd.Decimal

// Decimal exists to reduce ugly casts
func (a *WeiAmount) Decimal() *apd.Decimal {
	return (*apd.Decimal)(a)
}

// UnmarshalText hands off JSON decoding to apd.Decimal
func (a *WeiAmount) UnmarshalText(b []byte) error {
	return a.Decimal().UnmarshalText(b)
}

// MarshalText hands off JSON encoding to apd.Decimal
func (a *WeiAmount) MarshalText() ([]byte, error) {
	return a.Decimal().MarshalText()
}

// NewWeiAmount converts some amount of wei into an WeiAmount. (Only used by unit tests.)
// TODO: Should this method take *big.Int instead?
func NewWeiAmount(amount int64) *WeiAmount {
	if amount < 0 {
		panic("negative wei amounts are not supported")
	}
	return (*WeiAmount)(apd.New(amount, 0))
}

// ToWeiAmount casts an *apd.Decimal to *WeiAmount
func ToWeiAmount(wei *apd.Decimal) *WeiAmount {
	return (*WeiAmount)(wei)
}

// BigInt2Wei converts the passed *big.Int representation of a
// wei amount to the WeiAmount type. The returned value is a copy
// with no references to the passed value.
func BigInt2Wei(amount *big.Int) *WeiAmount {
	a := new(apd.BigInt).SetMathBigInt(amount)
	return (*WeiAmount)(apd.NewWithBigInt(a, 0))
}

// EtherToWei converts some amount of standard ether to an WeiAmount.
func EtherToWei(ethAmt *apd.Decimal) *WeiAmount {
	weiAmt := new(apd.Decimal).Set(ethAmt)
	weiAmt.Exponent += NumEtherDecimals
	if _, err := nonFractionalWeiCtx.Round(weiAmt, weiAmt); err != nil {
		panic(err)
	}
	return (*WeiAmount)(weiAmt)
}

// BigInt returns the given WeiAmount as a *big.Int
func (a *WeiAmount) BigInt() *big.Int {
	// Passing Quantize zero for the exponent set the coefficient to the whole-number
	// wei value. Round-half-up is used by default.
	wholeWeiVal := new(apd.Decimal)
	cond, err := DecimalCtx.Quantize(wholeWeiVal, a.Decimal(), 0)
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
	ether.Exponent -= NumEtherDecimals
	_, _ = ether.Reduce(ether)
	return ether
}

// AsStandard is an alias for AsEther, returning the wei amount as ether
func (a *WeiAmount) AsStandard() *apd.Decimal {
	return a.AsEther()
}

// String returns the wei amount as a base10 string
func (a *WeiAmount) String() string {
	return a.Decimal().String()
}

// ERC20TokenAmount represents some amount of an ERC20 token in the smallest denomination
type ERC20TokenAmount struct {
	amount      *apd.Decimal
	numDecimals int // num digits after the Decimal point needed for smallest denomination
}

// NewERC20TokenAmountFromBigInt converts some amount in the smallest token denomination into an ERC20TokenAmount.
func NewERC20TokenAmountFromBigInt(amount *big.Int, decimals int) *ERC20TokenAmount {
	asDecimal := new(apd.Decimal)
	asDecimal.Coeff.SetBytes(amount.Bytes())
	return &ERC20TokenAmount{
		amount:      asDecimal,
		numDecimals: decimals,
	}
}

// NewERC20TokenAmount converts some amount in the smallest token denomination into an ERC20TokenAmount.
func NewERC20TokenAmount(amount int64, decimals int) *ERC20TokenAmount {
	return &ERC20TokenAmount{
		amount:      apd.New(amount, 0),
		numDecimals: decimals,
	}
}

// NewERC20TokenAmountFromDecimals converts some amount of standard token in standard format
// to its smaller denomination.
// For example, if amount is 1.99 and decimals is 9, the resulting value stored
// is 1.99 * 10^9.
func NewERC20TokenAmountFromDecimals(amount *apd.Decimal, decimals int) *ERC20TokenAmount {
	adjusted := new(apd.Decimal).Set(amount)
	adjusted.Exponent += int32(decimals)
	return &ERC20TokenAmount{
		amount:      adjusted,
		numDecimals: decimals,
	}
}

// BigInt returns the ERC20TokenAmount as a *big.Int
func (a *ERC20TokenAmount) BigInt() *big.Int {
	noExponent := new(apd.Decimal)
	cond, err := nonFractionalWeiCtx.Quantize(noExponent, a.amount, 0)
	if err != nil {
		panic(err)
	}
	if cond.Rounded() {
		log.Warn("Converting ERC20TokenAmount=%s (digits=%d) to big.Int required rounding", a.amount, a.numDecimals)
	}
	return new(big.Int).SetBytes(noExponent.Coeff.Bytes())
}

// AsStandard returns the amount in standard form
func (a *ERC20TokenAmount) AsStandard() *apd.Decimal {
	tokenAmt := new(apd.Decimal).Set(a.amount)
	tokenAmt.Exponent -= NumEtherDecimals
	return tokenAmt
}

// String returns the ERC20TokenAmount as a base10 string
func (a *ERC20TokenAmount) String() string {
	return a.amount.String()
}
