package bob

import (
	"errors"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
)

type offerManager struct {
	offers map[types.Hash]*types.Offer
}

func newOfferManager() *offerManager {
	return &offerManager{
		offers: make(map[types.Hash]*types.Offer),
	}
}

func (om *offerManager) putOffer(o *types.Offer) {
	om.offers[o.GetID()] = o
}

func (om *offerManager) getOffer(id types.Hash) *types.Offer {
	return om.offers[id]
}

func (om *offerManager) deleteOffer(id types.Hash) {
	delete(om.offers, id)
}

// MakeOffer makes a new swap offer.
func (b *Instance) MakeOffer(o *types.Offer) error {
	balance, err := b.client.GetBalance(0)
	if err != nil {
		return err
	}

	if common.MoneroAmount(balance.UnlockedBalance) < common.MoneroToPiconero(o.MaximumAmount) {
		return errors.New("unlocked balance is less than maximum offer amount")
	}

	b.offerManager.putOffer(o)
	log.Infof("created new offer: %v", o)
	return nil
}

// GetOffers returns all current offers.
func (b *Instance) GetOffers() []*types.Offer {
	offers := make([]*types.Offer, len(b.offerManager.offers))
	i := 0
	for _, o := range b.offerManager.offers {
		offers[i] = o
		i++
	}
	return offers
}
