package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
)

// Peers calls net_peers to get the connected peers of a swapd instance.
func (c *Client) Peers() (*rpctypes.PeersResponse, error) {
	const (
		method = "net_peers"
	)

	res := &rpctypes.PeersResponse{}

	if err := c.Post(method, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}
