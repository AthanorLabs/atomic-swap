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

	if err := c.Post(method, req, nil); err != nil {
		return err
	}

	return nil
}

func (c *Client) GetSwapTimeout() (uint64, error) {
	const (
		method = "personal_getSwapTimeout"
	)

	var swapTimeout uint64
	if err := c.Post(method, nil, swapTimeout); err != nil {
		return 0, err
	}

	return swapTimeout, nil
}

// Balances calls personal_balances.
func (c *Client) Balances() (*rpctypes.BalancesResponse, error) {
	const (
		method = "personal_balances"
	)

	balances := &rpctypes.BalancesResponse{}
	if err := c.Post(method, nil, balances); err != nil {
		return nil, err
	}

	return balances, nil
}
