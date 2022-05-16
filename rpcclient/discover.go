package rpcclient

import (
	"encoding/json"

	"github.com/noot/atomic-swap/common/rpctypes"
	"github.com/noot/atomic-swap/common/types"
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
