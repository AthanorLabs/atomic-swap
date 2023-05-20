// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
)

// Addresses calls net_addresses.
func (c *Client) Addresses() (*rpctypes.AddressesResponse, error) {
	const (
		method = "net_addresses"
	)

	res := &rpctypes.AddressesResponse{}

	if err := c.post(method, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}
