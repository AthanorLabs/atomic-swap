package xmrmaker

import (
	"github.com/cockroachdb/apd/v3"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
)

// MakeOffer makes a new swap offer.
func (b *Instance) MakeOffer(
	o *types.Offer,
	relayerEndpoint string,
	relayerCommissionRate *apd.Decimal,
) (*types.OfferExtra, error) {
	b.backend.XMRClient().Lock()
	defer b.backend.XMRClient().Unlock()

	// get monero balance
	balance, err := b.backend.XMRClient().GetBalance(0)
	if err != nil {
		return nil, err
	}

	unlockedBalance := coins.NewPiconeroAmount(balance.UnlockedBalance).AsMonero()
	if unlockedBalance.Cmp(o.MaxAmount) <= 0 {
		return nil, errUnlockedBalanceTooLow{unlockedBalance, o.MaxAmount}
	}

	extra, err := b.offerManager.AddOffer(o, relayerEndpoint, relayerCommissionRate)
	if err != nil {
		return nil, err
	}

	b.net.Advertise()
	log.Infof("created new offer: %v", o)
	return extra, nil
}

// GetOffers returns all current offers.
func (b *Instance) GetOffers() []*types.Offer {
	return b.offerManager.GetOffers()
}

// ClearOffers clears all offers.
// If the offer list is empty, it clears all offers.
func (b *Instance) ClearOffers(offerIDs []types.Hash) error {
	l := len(offerIDs)
	if l == 0 {
		err := b.offerManager.ClearAllOffers()
		if err != nil {
			return err
		}
	}

	b.offerManager.ClearOfferIDs(offerIDs)
	return nil
}
