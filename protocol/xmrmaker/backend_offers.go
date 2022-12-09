package xmrmaker

import (
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
)

// MakeOffer makes a new swap offer.
func (b *Instance) MakeOffer(
	o *types.Offer,
	relayerEndpoint string,
	relayerCommission float64,
) (*types.OfferExtra, error) {
	b.backend.XMRClient().Lock()
	defer b.backend.XMRClient().Unlock()

	// get monero balance
	balance, err := b.backend.XMRClient().GetBalance(0)
	if err != nil {
		return nil, err
	}

	unlockedBalance := common.PiconeroAmount(balance.UnlockedBalance)
	if unlockedBalance < common.MoneroToPiconero(o.MaximumAmount) {
		return nil, errUnlockedBalanceTooLow{unlockedBalance.AsMonero(), o.MaximumAmount}
	}

	extra, err := b.offerManager.AddOffer(o, relayerEndpoint, relayerCommission)
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
