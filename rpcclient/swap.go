package rpcclient

import (
	"fmt"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/rpc"
)

// GetOngoingSwap calls swap_getOngoing
func (c *Client) GetOngoingSwap(id string) (*rpc.GetOngoingResponse, error) {
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
func (c *Client) GetPastSwap(id string) (*rpc.GetPastResponse, error) {
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

// Refund calls swap_refund
func (c *Client) Refund(id string) (*rpc.RefundResponse, error) {
	const (
		method = "swap_refund"
	)

	req := &rpc.RefundRequest{
		OfferID: id,
	}
	res := &rpc.RefundResponse{}

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
