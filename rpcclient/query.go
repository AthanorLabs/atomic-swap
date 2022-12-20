package rpcclient

import (
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/common/rpctypes"
)

// Query calls net_query.
func (c *Client) Query(who peer.ID) (*rpctypes.QueryPeerResponse, error) {
	const (
		method = "net_queryPeer"
	)

	req := &rpctypes.QueryPeerRequest{
		PeerID: who,
	}
	res := &rpctypes.QueryPeerResponse{}

	if err := c.Post(method, req, res); err != nil {
		return nil, err
	}

	return res, nil
}
