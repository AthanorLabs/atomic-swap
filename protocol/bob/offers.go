package bob

import (
	"errors"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
)

type offerManager struct {
	offers map[types.Hash]*types.Offer

	swapIDChs map[types.Hash]chan uint64

	// this channel is later used as the swap's statusCh when the offer is taken
	takenStatusChs map[types.Hash]chan types.Status
}

func newOfferManager() *offerManager {
	return &offerManager{
		offers:         make(map[types.Hash]*types.Offer),
		swapIDChs:      make(map[types.Hash]chan uint64),
		takenStatusChs: make(map[types.Hash]chan types.Status),
	}
}

func (om *offerManager) putOffer(o *types.Offer) {
	om.offers[o.GetID()] = o
	om.swapIDChs[o.GetID()] = make(chan uint64)
	om.takenStatusChs[o.GetID()] = make(chan types.Status, 7)
}

func (om *offerManager) getAndDeleteOffer(id types.Hash) (*types.Offer, chan uint64, chan types.Status) {
	offer, has := om.offers[id]
	if !has {
		return nil, nil, nil
	}

	statusCh := om.takenStatusChs[id]
	swapIDCh := om.swapIDChs[id]
	delete(om.offers, id)
	delete(om.takenStatusChs, id)
	return offer, swapIDCh, statusCh
}

// MakeOffer makes a new swap offer.
func (b *Instance) MakeOffer(o *types.Offer) (<-chan uint64, <-chan types.Status, error) {
	balance, err := b.client.GetBalance(0)
	if err != nil {
		return nil, nil, err
	}

	if common.MoneroAmount(balance.UnlockedBalance) < common.MoneroToPiconero(o.MaximumAmount) {
		return nil, nil, errors.New("unlocked balance is less than maximum offer amount")
	}

	b.offerManager.putOffer(o)
	log.Infof("created new offer: %v", o)
	return b.offerManager.swapIDChs[o.GetID()], b.offerManager.takenStatusChs[o.GetID()], nil
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
