package client

import (
	"encoding/json"

	"github.com/noot/atomic-swap/common/rpcclient"
	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/rpc"
)

// GetOffers calls swap_getOffers.
func (c *Client) GetOffers() ([]*types.Offer, error) {
	const (
		method = "swap_getOffers"
	)

	resp, err := rpcclient.PostRPC(c.endpoint, method, "{}")
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	var res *rpc.GetOffersResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
		return nil, err
	}

	return res.Offers, nil
}
