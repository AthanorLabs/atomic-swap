package swap

import (
	"github.com/athanorlabs/atomic-swap/common/types"
)

// Database contains the db functions used by the swap manager.
type Database interface {
	PutSwap(*Info) error
	HasSwap(id types.Hash) (bool, error)
	GetSwap(id types.Hash) (*Info, error)
	GetAllSwaps() ([]*Info, error)
}
