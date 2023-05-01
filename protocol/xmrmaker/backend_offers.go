// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package xmrmaker

import (
	"github.com/athanorlabs/atomic-swap/common/types"
)

// MakeOffer makes a new swap offer.
func (inst *Instance) MakeOffer(
	o *types.Offer,
	useRelayer bool,
) (*types.OfferExtra, error) {
	err := validateMinBalance(
		inst.backend.Ctx(),
		inst.backend.XMRClient(),
		inst.backend.ETHClient(),
		o.MaxAmount,
		o.EthAsset,
	)
	if err != nil {
		return nil, err
	}

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
