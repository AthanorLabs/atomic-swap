// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/rpc"
)

// SetSwapTimeout calls personal_setSwapTimeout.
func (c *Client) SetSwapTimeout(timeoutSeconds uint64) error {
	const (
		method = "personal_setSwapTimeout"
	)

	req := &rpc.SetSwapTimeoutRequest{
		Timeout: timeoutSeconds,
	}

	if err := c.Post(method, req, nil); err != nil {
		return err
	}

	return nil
}

// GetSwapTimeout calls personal_getSwapTimeout.
func (c *Client) GetSwapTimeout() (*rpc.GetSwapTimeoutResponse, error) {
	const (
		method = "personal_getSwapTimeout"
	)

	swapTimeout := &rpc.GetSwapTimeoutResponse{}
	if err := c.Post(method, nil, swapTimeout); err != nil {
		return nil, err
	}

	return swapTimeout, nil
}

// Balances calls personal_balances.
func (c *Client) Balances() (*rpctypes.BalancesResponse, error) {
	const (
		method = "personal_balances"
	)

	balances := &rpctypes.BalancesResponse{}
	if err := c.Post(method, nil, balances); err != nil {
		return nil, err
	}

	return balances, nil
}
