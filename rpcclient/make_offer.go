// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package rpcclient

import (
	"github.com/cockroachdb/apd/v3"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/common/types"
)

// MakeOffer calls net_makeOffer.
func (c *Client) MakeOffer(
	min, max *apd.Decimal,
	exchangeRate *coins.ExchangeRate,
	ethAsset types.EthAsset,
	useRelayer bool,
) (*rpctypes.MakeOfferResponse, error) {
	const (
		method = "net_makeOffer"
	)

	req := &rpctypes.MakeOfferRequest{
		MinAmount:    min,
		MaxAmount:    max,
		ExchangeRate: exchangeRate,
		EthAsset:     ethAsset,
		UseRelayer:   useRelayer,
	}
	res := &rpctypes.MakeOfferResponse{}

	if err := c.post(method, req, res); err != nil {
		return nil, err
	}

	return res, nil
}
