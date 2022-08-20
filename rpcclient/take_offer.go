package rpcclient

import (
	"encoding/json"
	"fmt"

	"github.com/athanorlabs/atomic-swap/common/rpctypes"
)

// TakeOffer calls net_takeOffer.
func (c *Client) TakeOffer(maddr string, offerID string, providesAmount float64) error {
	const (
		method = "net_takeOffer"
	)

	req := &rpctypes.TakeOfferRequest{
		Multiaddr:      maddr,
		OfferID:        offerID,
		ProvidesAmount: providesAmount,
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
		return fmt.Errorf("failed to call %s: %w", method, resp.Error)
	}

	var res *rpctypes.TakeOfferResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
		return err
	}

	return nil
}
