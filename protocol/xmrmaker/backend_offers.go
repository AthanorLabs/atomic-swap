// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package xmrmaker

import (
	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
)

// MakeOffer makes a new swap offer.
func (inst *Instance) MakeOffer(
	o *types.Offer,
	useRelayer bool,
) (*types.OfferExtra, error) {
	// get monero balance
	balance, err := inst.backend.XMRClient().GetBalance(0)
	if err != nil {
		return nil, err
	}

	unlockedBalance := coins.NewPiconeroAmount(balance.UnlockedBalance).AsMonero()
	if unlockedBalance.Cmp(o.MaxAmount) <= 0 {
		return nil, errUnlockedBalanceTooLow{o.MaxAmount, unlockedBalance}
	}

	// If it is an XMR-for-ETH offer, the min offer amount converted to ETH
	// must be less than relayer fee

	// If it is an XMR-for-TOKEN offer, the maker must have sufficient ETH
	// to claim

	if useRelayer && o.EthAsset.IsToken() {
		return nil, errRelayingWithNonEthAsset
	}

	extra, err := inst.offerManager.AddOffer(o, useRelayer)
	if err != nil {
		return nil, err
	}

	inst.net.Advertise()
	log.Infof("created new offer: %v", o)
	return extra, nil
}

// GetOffers returns all current offers.
func (inst *Instance) GetOffers() []*types.Offer {
	return inst.offerManager.GetOffers()
}

// ClearOffers clears all offers.
func (inst *Instance) ClearOffers(offerIDs []types.Hash) error {
	if len(offerIDs) == 0 {
		return inst.offerManager.ClearAllOffers()
	}
	return inst.offerManager.ClearOfferIDs(offerIDs)
}
