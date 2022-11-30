package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
)

// Query calls net_query.
func (c *Client) Query(maddr string) (*rpctypes.QueryPeerResponse, error) {
	const (
		method = "net_queryPeer"
	)

	req := &rpctypes.QueryPeerRequest{
		Multiaddr: maddr,
	}
	res := &rpctypes.QueryPeerResponse{}

	if err := c.Post(method, req, res); err != nil {
		return nil, err
	}

	return res, nil
}
