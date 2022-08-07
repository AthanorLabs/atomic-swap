package backend

import (
	"errors"
)

var (
	errMustProvideDaemonEndpoint = errors.New("environment is development, must provide monero daemon endpoint")
	errNilSwapContractOrAddress  = errors.New("must provide swap contract and address")
	errNoXMRDepositAddress       = errors.New("no xmr deposit address for given id")
)
