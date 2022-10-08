package backend

import (
	"errors"
)

var (
	errNilSwapContractOrAddress = errors.New("must provide swap contract and address")
	errNoXMRDepositAddress      = errors.New("no xmr deposit address for given id")
)
