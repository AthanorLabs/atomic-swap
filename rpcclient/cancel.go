// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/rpc"
)

// Cancel calls swap_cancel.
func (c *Client) Cancel(offerID types.Hash) (types.Status, error) {
	const (
		method = "swap_cancel"
	)

	req := &rpc.CancelRequest{
		OfferID: offerID,
	}
	res := &rpc.CancelResponse{}

	if err := c.Post(method, req, res); err != nil {
		return 0, err
	}

	return res.Status, nil
}
