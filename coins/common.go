package coins

import (
	"github.com/cockroachdb/apd/v3"
)

const (
	// NumEtherDecimals is the number of Decimal points needed to represent whole units of Wei in Ether
	NumEtherDecimals = 18
	// NumMoneroDecimals is the number of Decimal points needed to represent whole units of piconero in XMR
	NumMoneroDecimals = 12

	// MaxCoinPrecision is a somewhat arbitrary precision upper bound (2^256 consumes 78 digits)
	MaxCoinPrecision = 100
)

var (
	// DecimalCtx is the apd context used for math operations on our coins
	decimalCtx = apd.BaseContext.WithPrecision(MaxCoinPrecision)
)

// DecimalCtx clones and returns the apd.Context we use for coin math operations.
func DecimalCtx() *apd.Context {
	// return a clone to prevent external callers from modifying our context
	c := new(apd.Context)
	*c = *decimalCtx
	return c
}
