package common

import (
	"math"
	"math/big"
)

const (
	MainnetChainID = 1
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

type MoneroAmount uint64

func MoneroToPiconero(amount float64) MoneroAmount {
	return MoneroAmount(amount * numMoneroUnits)
}

func (a MoneroAmount) Uint64() uint64 {
	return uint64(a)
}

type EtherAmount big.Int

func NewEtherAmount(amount int64) EtherAmount {
	i := big.NewInt(amount)
	return EtherAmount(*i)
}

func EtherToWei(amount float64) EtherAmount {
	amt := big.NewFloat(amount)
	mult := big.NewFloat(numEtherUnits)
	res, _ := big.NewFloat(0).Mul(amt, mult).Int(nil)
	return EtherAmount(*res)
}

func (a EtherAmount) BigInt() *big.Int {
	i := big.Int(a)
	return &i
}

func (a EtherAmount) String() string {
	return a.BigInt().String()
}
