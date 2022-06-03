package rpcclient

import (
	"encoding/json"

	"github.com/noot/atomic-swap/common/rpctypes"
	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/rpc"
)

// Cancel calls swap_cancel.
func (c *Client) Cancel() (types.Status, error) {
	const (
		method = "swap_cancel"
	)

	resp, err := rpctypes.PostRPC(c.endpoint, method, "{}")
	if err != nil {
		return 0, err
	}

	if resp.Error != nil {
		return 0, resp.Error
	}

	var res *rpc.CancelResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
		return 0, err
	}

	return res.Status, nil
}
