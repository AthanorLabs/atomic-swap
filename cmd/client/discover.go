package main

import (
	"encoding/json"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/rpc"
	"github.com/noot/atomic-swap/rpcclient"
)

func (c *Client) discover(provides common.ProvidesCoin, searchTime uint64) ([][]string, error) {
	const (
		method = "net_discover"
	)

	req := &rpc.DiscoverRequest{
		Provides:   provides,
		SearchTime: searchTime,
	}

	params, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := rpcclient.PostRPC(c.endpoint, method, string(params))
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	var res *rpc.DiscoverResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
		return nil, err
	}

	return res.Peers, nil
}
