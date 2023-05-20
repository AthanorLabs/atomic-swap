// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/rpc"
)

// GetOffers calls swap_getOffers.
func (c *Client) GetOffers() (*rpc.GetOffersResponse, error) {
	const (
		method = "swap_getOffers"
	)

	resp := &rpc.GetOffersResponse{}

	if err := c.post(method, nil, resp); err != nil {
		return nil, err
	}

	return resp, nil
}
