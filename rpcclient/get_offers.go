package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/rpc"
)

// GetOffers calls swap_getOffers.
func (c *Client) GetOffers() ([]*types.Offer, error) {
	const (
		method = "swap_getOffers"
	)

	res := &rpc.GetOffersResponse{}

	if err := c.Post(method, nil, res); err != nil {
		return nil, err
	}

	return res.Offers, nil
}
