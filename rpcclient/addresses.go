package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/rpc"
)

// Addresses calls net_addresses.
func (c *Client) Addresses() ([]string, error) {
	const (
		method = "net_addresses"
	)

	res := &rpc.AddressesResponse{}

	if err := c.Post(method, nil, res); err != nil {
		return nil, err
	}

	return res.Addrs, nil
}
