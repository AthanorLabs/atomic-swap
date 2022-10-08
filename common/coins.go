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

// ToDecimals returns the amount rounded to a fixed number of decimal places
func (a EtherAmount) ToDecimals(decimals uint8) float64 {
	decimalsValue := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	wei := big.NewFloat(0).SetInt(a.BigInt())
	ether := big.NewFloat(0).Quo(wei, big.NewFloat(0).SetInt(decimalsValue))
	res, _ := ether.Float64()
	return res
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
