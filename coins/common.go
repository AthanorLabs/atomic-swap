// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package coins

import (
	"github.com/cockroachdb/apd/v3"
	logging "github.com/ipfs/go-log"
)

const (
	// NumEtherDecimals is the number of decimal points needed to represent whole units of Wei in Ether
	NumEtherDecimals = 18
	// NumMoneroDecimals is the number of decimal points needed to represent whole units of piconero in XMR
	NumMoneroDecimals = 12
	// MaxExchangeRateDecimals is the number of decimal points we allow in an exchange rate
	MaxExchangeRateDecimals = 6

	// MaxCoinPrecision is a somewhat arbitrary precision upper bound (2^256 consumes 78 digits)
	MaxCoinPrecision = 100
)

var (
	// decimalCtx is the apd context used for math operations on our coins
	decimalCtx = apd.BaseContext.WithPrecision(MaxCoinPrecision)

	log = logging.Logger("coins")
)

// DecimalCtx clones and returns the apd.Context we use for coin math operations.
func DecimalCtx() *apd.Context {
	// return a clone to prevent external callers from modifying our context
	c := new(apd.Context)
	*c = *decimalCtx
	return c
}
