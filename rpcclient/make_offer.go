package rpcclient

import (
	"encoding/json"
	"fmt"

	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/common/types"
)

// MakeOffer calls net_makeOffer.
func (c *Client) MakeOffer(min, max, exchangeRate float64) (string, error) {
	const (
		method = "net_makeOffer"
	)

	req := &rpctypes.MakeOfferRequest{
		MinimumAmount: min,
		MaximumAmount: max,
		ExchangeRate:  types.ExchangeRate(exchangeRate),
	}

	params, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	resp, err := rpctypes.PostRPC(c.endpoint, method, string(params))
	if err != nil {
		return "", err
	}

	if resp.Error != nil {
		return "", fmt.Errorf("failed to call %s: %w", method, resp.Error)
	}

	var res *rpctypes.MakeOfferResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
		return "", err
	}

	return res.ID, nil
}
