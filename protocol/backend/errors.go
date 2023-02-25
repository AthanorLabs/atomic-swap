package backend

import (
	"errors"
)

var (
	errNilSwapContractOrAddress = errors.New("must provide swap contract and address")
)
