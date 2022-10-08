package rpcclient

import (
	"encoding/json"

	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/common/types"
)

// Discover calls net_discover.
func (c *Client) Discover(provides types.ProvidesCoin, searchTime uint64) ([][]string, error) {
	const (
		method = "net_discover"
	)

	req := &rpctypes.DiscoverRequest{
		Provides:   provides,
		SearchTime: searchTime,
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

	var res *rpctypes.DiscoverResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
		return nil, err
	}

	return res.Peers, nil
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

	var res rpctypes.QueryAllResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
		return nil, err
	}

	return res.PeersWithOffers, nil
}
