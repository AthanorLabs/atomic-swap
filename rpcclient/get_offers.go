package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/rpc"
)

// GetOffers calls swap_getOffers.
func (c *Client) GetOffers() (*rpc.GetOffersResponse, error) {
	const (
		method = "swap_getOffers"
	)

	resp := &rpc.GetOffersResponse{}

	if err := c.Post(method, nil, resp); err != nil {
		return nil, err
	}

	return resp, nil
}
