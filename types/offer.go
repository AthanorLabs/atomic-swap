package types

import (
	"encoding/json"

	"github.com/noot/atomic-swap/common"

	"golang.org/x/crypto/sha3"
)

type Hash [32]byte

type Offer struct {
	hash          Hash
	Provides      common.ProvidesCoin
	MinimumAmount float64
	MaximumAmount float64
	ExchangeRate  common.ExchangeRate
}

func (o *Offer) Hash() Hash {
	if o.hash != [32]byte{} {
		return o.hash
	}

	b, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}

	o.hash = sha3.Sum256(b)
	return o.hash
}
