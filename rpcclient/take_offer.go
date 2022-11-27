package rpcclient

import (
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

	if err := rpctypes.PostRPC(c.endpoint, method, req, nil); err != nil {
		return err
	}

	return nil
}
