package common

import (
	"math"
	"math/big"
	"strconv"
	"strings"
)

var (
	numEtherUnits  = math.Pow(10, 18)
	numMoneroUnits = math.Pow(10, 12)
)

// PiconeroAmount represents some amount of piconero (the smallest denomination of monero)
type PiconeroAmount uint64

// MoneroToPiconero converts an amount of standard monero and returns it as a PiconeroAmount
func MoneroToPiconero(amount float64) PiconeroAmount {
	return PiconeroAmount(amount * numMoneroUnits)
}

// Uint64 ...
func (a PiconeroAmount) Uint64() uint64 {
	return uint64(a)
}

// AsMonero converts the piconero PiconeroAmount into standard units
func (a PiconeroAmount) AsMonero() float64 {
	return float64(a) / numMoneroUnits
}

// WeiAmount represents some amount of ether in the smallest denomination (wei)
type WeiAmount big.Int

// NewWeiAmount converts some amount of wei into an WeiAmount.
func NewWeiAmount(amount int64) WeiAmount {
	i := big.NewInt(amount)
	return WeiAmount(*i)
}

// EtherToWei converts some amount of standard ether to an WeiAmount.
func EtherToWei(amount float64) WeiAmount {
	amt := big.NewFloat(amount)
	mult := big.NewFloat(numEtherUnits)
	prod := new(big.Float).Mul(amt, mult)
	res := round(prod)
	return WeiAmount(*res)
}

// BigInt returns the given WeiAmount as a *big.Int
func (a WeiAmount) BigInt() *big.Int {
	i := big.Int(a)
	return &i
}

// AsEther returns the wei amount as ether
func (a WeiAmount) AsEther() float64 {
	wei := new(big.Float).SetInt(a.BigInt())
	mult := big.NewFloat(numEtherUnits)
	ether := new(big.Float).Quo(wei, mult)
	res, _ := ether.Float64()
	return res
}

// AsStandard returns the wei amount as ether
func (a WeiAmount) AsStandard() float64 {
	return a.AsEther()
}

// String ...
func (a WeiAmount) String() string {
	return a.BigInt().String()
}

// FmtFloat creates a string from a floating point value that keeps enough precision to
// represent a single wei, does not use exponent notation, and has no trailing zeros after
// the decimal point.
func FmtFloat(f float64) string {
	// 18 decimal points is needed to represent 1 wei in ETH
	s := strconv.FormatFloat(f, 'f', 18, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
	}
	return s
}

// ERC20TokenAmount represents some amount of an ERC20 token in the smallest denomination
type ERC20TokenAmount struct {
	amount   *big.Int
	numUnits float64 // 10^decimals
}

// NewERC20TokenAmountFromBigInt converts some amount in the smallest token denomination into an ERC20TokenAmount.
func NewERC20TokenAmountFromBigInt(amount *big.Int, decimals int) *ERC20TokenAmount {
	return &ERC20TokenAmount{
		amount:   amount,
		numUnits: math.Pow10(decimals),
	}
}

// NewERC20TokenAmount converts some amount in the smallest token denomination into an ERC20TokenAmount.
func NewERC20TokenAmount(amount int64, decimals int) *ERC20TokenAmount {
	return &ERC20TokenAmount{
		amount:   big.NewInt(amount),
		numUnits: math.Pow10(decimals),
	}
}

// NewERC20TokenAmountFromDecimals converts some amount of standard token in standard format
// to its smaller denomination.
// For example, if amount is 1.99 and decimals is 9, the resulting value stored
// is 1.99 * 10^9.
func NewERC20TokenAmountFromDecimals(amount float64, decimals int) *ERC20TokenAmount {
	numUnits := math.Pow10(decimals)
	amt := big.NewFloat(amount)
	mult := big.NewFloat(numUnits)
	prod := new(big.Float).Mul(amt, mult)
	res := round(prod)
	return &ERC20TokenAmount{
		amount:   res,
		numUnits: numUnits,
	}
}

// BigInt returns the given ERC20TokenAmount as a *big.Int
func (a *ERC20TokenAmount) BigInt() *big.Int {
	return a.amount
}

// AsStandard returns the amount in standard form
func (a *ERC20TokenAmount) AsStandard() float64 {
	wei := new(big.Float).SetInt(a.BigInt())
	mult := big.NewFloat(a.numUnits)
	ether := new(big.Float).Quo(wei, mult)
	res, _ := ether.Float64()
	return res
}

// String ...
func (a *ERC20TokenAmount) String() string {
	return a.BigInt().String()
}

// round rounds the input *big.Float to a *big.Int.
// eg. if the input is 33.49, it returns 33.
// if the input is 33.5, it returns 34.
func round(num *big.Float) *big.Int {
	// ignore overflow, we only care about the least significant decimals
	numAsFloat, _ := num.Float64()
	rounded := math.Round(numAsFloat)
	if rounded <= numAsFloat {
		res, _ := num.Int(nil)
		return res
	}

	res, _ := new(big.Float).Add(num, big.NewFloat(1)).Int(nil)
	return res
}
