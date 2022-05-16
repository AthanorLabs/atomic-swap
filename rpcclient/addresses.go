package rpcclient

import (
	"encoding/json"

	"github.com/noot/atomic-swap/common/rpctypes"
	"github.com/noot/atomic-swap/rpc"
)

// Addresses calls net_addresses.
func (c *Client) Addresses() ([]string, error) {
	const (
		method = "net_addresses"
	)

	resp, err := rpctypes.PostRPC(c.endpoint, method, "{}")
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
