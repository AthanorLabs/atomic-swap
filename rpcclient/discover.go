package rpcclient

import (
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/common/types"
)

// Discover calls net_discover.
func (c *Client) Discover(provides types.ProvidesCoin, searchTime uint64) ([]peer.ID, error) {
	const (
		method = "net_discover"
	)

	req := &rpctypes.DiscoverRequest{
		Provides:   provides,
		SearchTime: searchTime,
	}
	res := &rpctypes.DiscoverResponse{}

	if err := c.Post(method, req, res); err != nil {
		return nil, err
	}

	return res.PeerIDs, nil
}

// QueryAll calls net_queryAll.
func (c *Client) QueryAll(provides types.ProvidesCoin, searchTime uint64) ([]*rpctypes.PeerWithOffers, error) {
	const (
		method = "net_queryAll"
	)

	req := &rpctypes.DiscoverRequest{
		Provides:   provides,
		SearchTime: searchTime,
	}
	res := &rpctypes.QueryAllResponse{}

	if err := c.Post(method, req, res); err != nil {
		return nil, err
	}

	return res.PeersWithOffers, nil
}
