package rpcclient

import (
	"fmt"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/rpc"
)

// GetPastSwapIDs calls swap_getPastIDs
func (c *Client) GetPastSwapIDs() ([]string, error) {
	const (
		method = "swap_getPastIDs"
	)

	res := &rpc.GetPastIDsResponse{}

	if err := c.Post(method, nil, res); err != nil {
		return nil, err
	}

	return res.IDs, nil
}

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

// GetStage calls swap_getStage
func (c *Client) GetStage(id string) (*rpc.GetStageResponse, error) {
	const (
		method = "swap_getStage"
	)

	req := &rpc.GetStageRequest{
		OfferID: id,
	}
	res := &rpc.GetStageResponse{}

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
