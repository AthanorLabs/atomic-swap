package rpcclient

import (
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/rpc"
)

// Cancel calls swap_cancel.
func (c *Client) Cancel(id string) (types.Status, error) {
	const (
		method = "swap_cancel"
	)

	req := &rpc.CancelRequest{
		OfferID: id,
	}
	res := &rpc.CancelResponse{}

	if err := c.Post(method, req, res); err != nil {
		return 0, err
	}

	return res.Status, nil
}
