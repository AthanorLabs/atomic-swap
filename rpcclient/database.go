// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package rpcclient provides client libraries for interacting with a local swapd instance using
// the JSON-RPC remote procedure call protocol.
package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/rpc"
)

// GetContractSwapInfo calls database_getContractSwapInfo.
func (c *Client) GetContractSwapInfo(offerID types.Hash) (*rpc.GetContractSwapInfoResponse, error) {
	const (
		method = "database_getContractSwapInfo"
	)

	req := &rpc.GetContractSwapInfoRequest{
		OfferID: offerID,
	}

	res := &rpc.GetContractSwapInfoResponse{}
	if err := c.post(method, req, res); err != nil {
		return nil, err
	}

	return res, nil
}

// GetSwapSecret calls database_getSwapSecret.
func (c *Client) GetSwapSecret(offerID types.Hash) (*rpc.GetSwapSecretResponse, error) {
	const (
		method = "database_getSwapSecret"
	)

	req := &rpc.GetSwapSecretRequest{
		OfferID: offerID,
	}

	res := &rpc.GetSwapSecretResponse{}
	if err := c.post(method, req, res); err != nil {
		return nil, err
	}

	return res, nil
}
