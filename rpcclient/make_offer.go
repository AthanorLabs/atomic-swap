package rpcclient

import (
	"github.com/cockroachdb/apd/v3"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/common/types"
)

// MakeOffer calls net_makeOffer.
func (c *Client) MakeOffer(
	min, max *apd.Decimal,
	exchangeRate *coins.ExchangeRate,
	ethAsset types.EthAsset,
	relayerEndpoint string,
	relayerFee *apd.Decimal,
) (*rpctypes.MakeOfferResponse, error) {
	const (
		method = "net_makeOffer"
	)

	req := &rpctypes.MakeOfferRequest{
		MinAmount:       min,
		MaxAmount:       max,
		ExchangeRate:    exchangeRate,
		EthAsset:        ethAsset,
		RelayerEndpoint: relayerEndpoint,
		RelayerFee:      relayerFee,
	}
	res := &rpctypes.MakeOfferResponse{}

	if err := c.Post(method, req, res); err != nil {
		return nil, err
	}

	return res, nil
}
