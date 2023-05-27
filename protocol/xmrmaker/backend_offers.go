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

	if o.EthAsset.IsToken() {
		if useRelayer {
			return nil, errRelayingWithNonEthAsset
		}

		token, err := inst.backend.ETHClient().ERC20Info(inst.backend.Ctx(), o.EthAsset.Address()) //nolint:govet
		if err != nil {
			return nil, err
		}

		// We limit exchange rates to 6 decimals and the min/max XMR amounts to
		// 12 decimals when marshalling the offer. This means we can never
		// exceed ETH's 18 decimals when multiplying min/max values by the
		// exchange rate. Tokens can have far fewer decimals though, so we need
		// additional checks. Calculating the exchange rate will give a good
		// error message if the combined precision of the exchange rate and
		// min/max values would exceed the token's precision.
		_, err = o.ExchangeRate.ToERC20Amount(o.MinAmount, token)
		if err != nil {
			return nil, err
		}

		_, err = o.ExchangeRate.ToERC20Amount(o.MaxAmount, token)
		if err != nil {
			return nil, err
		}

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
