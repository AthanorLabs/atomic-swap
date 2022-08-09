package rpcclient

import (
	"encoding/json"
	"fmt"

	"github.com/noot/atomic-swap/common/rpctypes"
	"github.com/noot/atomic-swap/rpc"
)

// GetPastSwapIDs calls swap_getPastIDs
func (c *Client) GetPastSwapIDs() ([]string, error) {
	const (
		method = "swap_getPastIDs"
	)

	resp, err := rpctypes.PostRPC(c.endpoint, method, "{}")
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("failed to call %s: %w", method, resp.Error)
	}

	var res *rpc.GetPastIDsResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
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

	params, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := rpctypes.PostRPC(c.endpoint, method, string(params))
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("failed to call %s: %w", method, resp.Error)
	}

	var res *rpc.GetOngoingResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
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

	params, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := rpctypes.PostRPC(c.endpoint, method, string(params))
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	var res *rpc.GetPastResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
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

	params, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := rpctypes.PostRPC(c.endpoint, method, string(params))
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("failed to call %s: %w", method, resp.Error)
	}

	var res *rpc.RefundResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
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

	params, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := rpctypes.PostRPC(c.endpoint, method, string(params))
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("failed to call %s: %w", method, resp.Error)
	}

	var res *rpc.GetStageResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// ClearOffers calls swap_clearOffers
func (c *Client) ClearOffers(ids []string) error {
	const (
		method = "swap_clearOffers"
	)

	req := &rpc.ClearOffersRequest{
		IDs: ids,
	}

	params, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := rpctypes.PostRPC(c.endpoint, method, string(params))
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("failed to call %s: %w", method, resp.Error)
	}

	return nil
}
