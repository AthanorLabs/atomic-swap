package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/rpc"
)

// SetSwapTimeout calls personal_setSwapTimeout.
func (c *Client) SetSwapTimeout(timeoutSeconds uint64) error {
	const (
		method = "personal_setSwapTimeout"
	)

	req := &rpc.SetSwapTimeoutRequest{
		Timeout: timeoutSeconds,
	}

	if err := rpctypes.PostRPC(c.endpoint, method, req, nil); err != nil {
		return err
	}

	return nil
}

// Balances calls personal_balances.
func (c *Client) Balances() (*rpctypes.BalancesResponse, error) {
	const (
		method = "personal_balances"
	)

	balances := &rpctypes.BalancesResponse{}
	if err := rpctypes.PostRPC(c.endpoint, method, nil, balances); err != nil {
		return nil, err
	}

	return balances, nil
}
