package common

import (
	"math"
	"math/big"
)

const (
	MainnetChainID = 1 //nolint
	RopstenChainID = 3
	GanacheChainID = 1337

	DefaultAliceMoneroEndpoint  = "http://127.0.0.1:18084/json_rpc"
	DefaultBobMoneroEndpoint    = "http://127.0.0.1:18083/json_rpc"
	DefaultMoneroDaemonEndpoint = "http://127.0.0.1:18081/json_rpc"
	DefaultEthEndpoint          = "ws://localhost:8545"

	// DefaultPrivKeyAlice is the private key at index 0 from `ganache-cli -d`
	DefaultPrivKeyAlice = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"

	// DefaultPrivKeyBob is the private key at index 1 from `ganache-cli -d`
	DefaultPrivKeyBob = "6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1"
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

// EtherAmount represents some amout of ether in the smallest denomination (wei)
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

// String ...
func (a EtherAmount) String() string {
	return a.BigInt().String()
}
