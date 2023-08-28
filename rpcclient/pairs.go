// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
)

// Pairs calls net_pairs to get pairs from all offers.
func (c *Client) Pairs(searchTime uint64) (*rpctypes.PairsResponse, error) {
	const (
		method = "net_pairs"
	)

	req := &rpctypes.PairsRequest{
		SearchTime: searchTime,
	}

	res := &rpctypes.PairsResponse{}

	if err := c.post(method, req, res); err != nil {
		return nil, err
	}

	return res, nil
}
