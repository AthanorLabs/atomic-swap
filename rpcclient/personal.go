package rpcclient

import (
	"encoding/json"

	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/rpc"
)

// SetSwapTimeout calls personal_setSwapTimeout.
func (c *Client) SetSwapTimeout(duration uint64) error {
	const (
		method = "personal_setSwapTimeout"
	)

	req := &rpc.SetSwapTimeoutRequest{
		Timeout: duration,
	}

	params, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := rpctypes.PostRPC(c.endpoint, method, string(params))
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return resp.Error
	}

	return nil
}
