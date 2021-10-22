package bob

import (
	"github.com/noot/atomic-swap/monero"
)

// Bob contains the functions that will be called by a user who owns XMR
// and wishes to swap for ETH.
type Bob interface {
	GenerateKeys() *monero.PublicKeyPair
}
