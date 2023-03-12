package xmrmaker

import (
	"github.com/cockroachdb/apd/v3"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/relayer"
)

// MakeOffer makes a new swap offer.
func (b *Instance) MakeOffer(
	o *types.Offer,
	relayerFee *apd.Decimal,
) (*types.OfferExtra, error) {
	// get monero balance
	balance, err := b.backend.XMRClient().GetBalance(0)
	if err != nil {
		return nil, err
	}

	unlockedBalance := coins.NewPiconeroAmount(balance.UnlockedBalance).AsMonero()
	if unlockedBalance.Cmp(o.MaxAmount) <= 0 {
		return nil, errUnlockedBalanceTooLow{unlockedBalance, o.MaxAmount}
	}

	if relayerFee != nil {
		if o.EthAsset != types.EthAssetETH {
			return nil, errRelayingWithNonEthAsset
		}
		if relayerFee.Cmp(relayer.MinRelayerFeeEth) < 0 {
			return nil, errRelayerFeeTooLow{relayer.MinRelayerFeeEth, relayerFee}
		}
	}

	extra, err := b.offerManager.AddOffer(o, relayerFee)
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
