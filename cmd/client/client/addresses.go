package client

import (
	"encoding/json"

	"github.com/noot/atomic-swap/rpc"
	"github.com/noot/atomic-swap/rpcclient"
)

func (c *Client) Addresses() ([]string, error) {
	const (
		method = "net_addresses"
	)

	resp, err := rpcclient.PostRPC(c.endpoint, method, "{}")
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	var res *rpc.AddressesResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
		return nil, err
	}

	return res.Addrs, nil
}
