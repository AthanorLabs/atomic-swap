package xmrmaker

import (
	"fmt"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
	pcommon "github.com/noot/atomic-swap/protocol"
)

type offerWithExtra struct {
	offer *types.Offer
	extra *types.OfferExtra
}

type offerManager struct {
	offers   map[types.Hash]*offerWithExtra
	basepath string
}

func newOfferManager(basepath string) *offerManager {
	return &offerManager{
		offers:   make(map[types.Hash]*offerWithExtra),
		basepath: basepath,
	}
}

func (om *offerManager) putOffer(o *types.Offer) *types.OfferExtra {
	offer, has := om.offers[o.GetID()]
	if has {
		return offer.extra
	}

	extra := &types.OfferExtra{
		//IDCh:     make(chan uint64, 1),
		StatusCh: make(chan types.Status, 7),
		InfoFile: pcommon.GetSwapInfoFilepath(om.basepath),
	}

	oe := &offerWithExtra{
		offer: o,
		extra: extra,
	}

	om.offers[o.GetID()] = oe
	return extra
}

func (om *offerManager) getAndDeleteOffer(id types.Hash) (*types.Offer, *types.OfferExtra) {
	offer, has := om.offers[id]
	if !has {
		return nil, nil
	}

	delete(om.offers, id)
	return offer.offer, offer.extra
}

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

	extra := b.offerManager.putOffer(o)
	log.Infof("created new offer: %v", o)
	return extra, nil
}

// GetOffers returns all current offers.
func (b *Instance) GetOffers() []*types.Offer {
	// lock entire instance, as if an offer is taken a swap will be deleted
	b.swapMu.Lock()
	defer b.swapMu.Unlock()

	offers := make([]*types.Offer, len(b.offerManager.offers))
	i := 0
	for _, o := range b.offerManager.offers {
		offers[i] = o.offer
		i++
	}
	return offers
}

// ClearOffers clears the given offers.
// If the offer list is empty, it clears all offers.
func (b *Instance) ClearOffers(ids []string) error {
	b.swapMu.Lock()
	defer b.swapMu.Unlock()
	if len(ids) == 0 {
		b.offerManager.offers = make(map[types.Hash]*offerWithExtra)
		return nil
	}

	for _, id := range ids {
		hash, err := types.HexToHash(id)
		if err != nil {
			return fmt.Errorf("invalid id %s: %w", id, err)
		}

		delete(b.offerManager.offers, hash)
	}

	return nil
}
