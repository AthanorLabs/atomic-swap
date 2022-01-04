package client

import (
	"encoding/json"
	"fmt"

	"github.com/noot/atomic-swap/rpc"
	"github.com/noot/atomic-swap/rpcclient"
)

// TakeOffer calls net_takeOffer.
func (c *Client) TakeOffer(maddr string, offerID string, providesAmount float64) (bool, float64, error) {
	const (
		method = "net_takeOffer"
	)

	req := &rpc.TakeOfferRequest{
		Multiaddr:      maddr,
		OfferID:        offerID,
		ProvidesAmount: providesAmount,
	}

	params, err := json.Marshal(req)
	if err != nil {
		return false, 0, err
	}

	resp, err := rpcclient.PostRPC(c.endpoint, method, string(params))
	if err != nil {
		return false, 0, err
	}

	if resp.Error != nil {
		return false, 0, fmt.Errorf("failed to call net_initiate: %w", resp.Error)
	}

	var res *rpc.TakeOfferResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
		return false, 0, err
	}

	return res.Success, res.ReceivedAmount, nil
}
