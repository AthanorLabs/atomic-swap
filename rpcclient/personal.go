package rpcclient

import (
	"encoding/json"
	"errors"

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

// Balances calls personal_balances.
func (c *Client) Balances() (*rpctypes.BalancesResponse, error) {
	const (
		method = "personal_balances"
	)

	resp, err := rpctypes.PostRPC(c.endpoint, method, "{}")
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	var balances rpctypes.BalancesResponse
	if err = json.Unmarshal(resp.Result, &balances); err != nil {
		return nil, err
	}
	if balances.WeiBalance == nil {
		return nil, errors.New("required field wei_balance missing")
	}

	return &balances, nil
}
