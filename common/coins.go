package common

import (
	"math"
	"math/big"
	"strconv"
)

var (
	numEtherUnits  = math.Pow(10, 18)
	numMoneroUnits = math.Pow(10, 12)
)

// MoneroAmount represents some amount of piconero (the smallest denomination of monero)
type MoneroAmount uint64

// MoneroToPiconero converts an amount of standard monero and returns it as a MoneroAmount
func MoneroToPiconero(amount float64) MoneroAmount {
	return MoneroAmount(amount * numMoneroUnits)
}

// Uint64 ...
func (a MoneroAmount) Uint64() uint64 {
	return uint64(a)
}

// AsMonero converts the piconero MoneroAmount into standard units
func (a MoneroAmount) AsMonero() float64 {
	return float64(a) / numMoneroUnits
}

// EtherAmount represents some amount of ether in the smallest denomination (wei)
type EtherAmount big.Int

// NewEtherAmount converts some amount of wei into an EtherAmount.
func NewEtherAmount(amount int64) EtherAmount {
	i := big.NewInt(amount)
	return EtherAmount(*i)
}

// EtherToWei converts some amount of standard ether to an EtherAmount.
func EtherToWei(amount float64) EtherAmount {
	amt := big.NewFloat(amount)
	mult := big.NewFloat(numEtherUnits)
	res, _ := big.NewFloat(0).Mul(amt, mult).Int(nil)
	return EtherAmount(*res)
}

// BigInt returns the given EtherAmount as a *big.Int
func (a EtherAmount) BigInt() *big.Int {
	i := big.Int(a)
	return &i
}

// AsEther returns the wei amount as ether
func (a EtherAmount) AsEther() float64 {
	wei := big.NewFloat(0).SetInt(a.BigInt())
	mult := big.NewFloat(numEtherUnits)
	ether := big.NewFloat(0).Quo(wei, mult)
	res, _ := ether.Float64()
	return res
}

// AsStandard returns the wei amount as ether
func (a EtherAmount) AsStandard() float64 {
	return a.AsEther()
}

// String ...
func (a EtherAmount) String() string {
	return a.BigInt().String()
}

// FmtFloat creates a string from a floating point value that keeps maximum precision,
// does not use exponent notation, and has no trailing zeros after the decimal point.
func FmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// ERC20TokenAmount represents some amount of an ERC20 token in the smallest denomination
type ERC20TokenAmount struct {
	amount   *big.Int
	numUnits float64 // 10^decimals
}

func NewERC20TokenAmountFromBigInt(amount *big.Int, decimals float64) *ERC20TokenAmount {
	return &ERC20TokenAmount{
		amount:   amount,
		numUnits: math.Pow(10, decimals),
	}
}

// NewERC20TokenAmount converts some amount of wei into an EtherAmount.
func NewERC20TokenAmount(amount int64, decimals float64) *ERC20TokenAmount {
	return &ERC20TokenAmount{
		amount:   big.NewInt(amount),
		numUnits: math.Pow(10, decimals),
	}
}

// NewERC20TokenAmountFromDecimals converts some amount of standard token in standard format
// to its smaller denomination.
func NewERC20TokenAmountFromDecimals(amount float64, decimals float64) *ERC20TokenAmount {
	numUnits := math.Pow(10, decimals)
	amt := big.NewFloat(amount)
	mult := big.NewFloat(numUnits)
	res, _ := big.NewFloat(0).Mul(amt, mult).Int(nil)
	return &ERC20TokenAmount{
		amount:   res,
		numUnits: numUnits,
	}
}

// BigInt returns the given EtherAmount as a *big.Int
func (a *ERC20TokenAmount) BigInt() *big.Int {
	return a.amount
}

// AsStandard returns the amount in standard form
func (a *ERC20TokenAmount) AsStandard() float64 {
	wei := big.NewFloat(0).SetInt(a.BigInt())
	mult := big.NewFloat(a.numUnits)
	ether := big.NewFloat(0).Quo(wei, mult)
	res, _ := ether.Float64()
	return res
}

// String ...
func (a *ERC20TokenAmount) String() string {
	return a.BigInt().String()
}
