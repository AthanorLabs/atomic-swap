package xmrmaker

import (
	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
)

// MakeOffer makes a new swap offer.
func (b *Instance) MakeOffer(
	o *types.Offer,
	useRelayer bool,
) (*types.OfferExtra, error) {
	log.Debugf("attempting to make offer, getting monero balance")

	// get monero balance
	balance, err := b.backend.XMRClient().GetBalance(0)
	if err != nil {
		return nil, err
	}

	log.Debugf("got monero balance")

	unlockedBalance := coins.NewPiconeroAmount(balance.UnlockedBalance).AsMonero()
	if unlockedBalance.Cmp(o.MaxAmount) <= 0 {
		return nil, errUnlockedBalanceTooLow{o.MaxAmount, unlockedBalance}
	}

	if useRelayer && o.EthAsset != types.EthAssetETH {
		return nil, errRelayingWithNonEthAsset
	}

	extra, err := b.offerManager.AddOffer(o, useRelayer)
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
func (b *Instance) ClearOffers(offerIDs []types.Hash) error {
	if len(offerIDs) == 0 {
		return b.offerManager.ClearAllOffers()
	}
	return b.offerManager.ClearOfferIDs(offerIDs)
}
