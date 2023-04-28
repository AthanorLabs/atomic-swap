// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package rpcclient

import (
	"fmt"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/rpc"
)

// GetOngoingSwap calls swap_getOngoing
func (c *Client) GetOngoingSwap(id *types.Hash) (*rpc.GetOngoingResponse, error) {
	const (
		method = "swap_getOngoing"
	)

	req := &rpc.GetOngoingRequest{
		OfferID: id,
	}

	res := &rpc.GetOngoingResponse{}

	if err := c.Post(method, req, res); err != nil {
		return nil, err
	}
	return res, nil
}

// GetPastSwap calls swap_getPast
func (c *Client) GetPastSwap(id *types.Hash) (*rpc.GetPastResponse, error) {
	const (
		method = "swap_getPast"
	)

	req := &rpc.GetPastRequest{
		OfferID: id,
	}

	res := &rpc.GetPastResponse{}

	if err := c.Post(method, req, res); err != nil {
		return nil, err
	}

	return res, nil
}

// GetStatus calls swap_getStatus
func (c *Client) GetStatus(id types.Hash) (*rpc.GetStatusResponse, error) {
	const (
		method = "swap_getStatus"
	)

	req := &rpc.GetStatusRequest{
		ID: id,
	}
	res := &rpc.GetStatusResponse{}

	if err := c.Post(method, req, res); err != nil {
		return nil, err
	}

	return res, nil
}

// ClearOffers calls swap_clearOffers
func (c *Client) ClearOffers(offerIDs []types.Hash) error {
	const (
		method = "swap_clearOffers"
	)

	req := &rpc.ClearOffersRequest{
		OfferIDs: offerIDs,
	}

	if err := c.Post(method, req, nil); err != nil {
		return fmt.Errorf("failed to call %s: %w", method, err)
	}

	return nil
}

// Claim calls swap_claim
func (c *Client) Claim(offerID types.Hash) (*rpc.ManualTransactionResponse, error) {
	const (
		method = "swap_claim"
	)

	req := &rpc.ManualTransactionRequest{
		OfferID: offerID,
	}

	res := &rpc.ManualTransactionResponse{}

	if err := c.Post(method, req, res); err != nil {
		return nil, err
	}

	return res, nil
}

// Refund calls swap_refund
func (c *Client) Refund(offerID types.Hash) (*rpc.ManualTransactionResponse, error) {
	const (
		method = "swap_refund"
	)

	req := &rpc.ManualTransactionRequest{
		OfferID: offerID,
	}

	res := &rpc.ManualTransactionResponse{}

	if err := c.Post(method, req, res); err != nil {
		return nil, err
	}

	return res, nil
}

// SuggestedExchangeRate calls swap_suggestedExchangeRate
func (c *Client) SuggestedExchangeRate() (*rpc.SuggestedExchangeRateResponse, error) {
	const (
		method = "swap_suggestedExchangeRate"
	)

	res := &rpc.SuggestedExchangeRateResponse{}
	if err := c.Post(method, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}
