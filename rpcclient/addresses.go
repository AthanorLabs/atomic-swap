package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
)

// Addresses calls net_addresses.
func (c *Client) Addresses() (*rpctypes.AddressesResponse, error) {
	const (
		method = "net_addresses"
	)

	res := &rpctypes.AddressesResponse{}

	if err := c.Post(method, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}
