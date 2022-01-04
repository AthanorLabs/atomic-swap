package bob

import (
	"github.com/noot/atomic-swap/types"
)

type offerManager struct {
	offers map[types.Hash]*types.Offer
}

func (b *bob) MakeOffer(o *types.Offer) error {
	return nil
}
