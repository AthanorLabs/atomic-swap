package xmrmaker

import (
	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
)

// MakeOffer makes a new swap offer.
func (b *Instance) MakeOffer(o *types.Offer) (*types.OfferExtra, error) {
	b.backend.LockClient()
	defer b.backend.UnlockClient()

	balance, err := b.backend.GetBalance(0)
	if err != nil {
		return nil, err
	}

	if common.MoneroAmount(balance.UnlockedBalance) < common.MoneroToPiconero(o.MaximumAmount) {
		return nil, errUnlockedBalanceTooLow
	}

	extra := b.offerManager.AddOffer(o)
	log.Infof("created new offer: %v", o)
	return extra, nil
}

// GetOffers returns all current offers.
func (b *Instance) GetOffers() []*types.Offer {
	return b.offerManager.GetOffers()
}

// ClearOffers clears all offers.
func (b *Instance) ClearOffers() {
	b.offerManager.ClearOffers()
}
