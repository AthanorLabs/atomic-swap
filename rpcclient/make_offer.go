package rpcclient

import (
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/common/types"
)

// MakeOffer calls net_makeOffer.
func (c *Client) MakeOffer(
	min, max, exchangeRate float64,
	ethAsset types.EthAsset,
	relayerEndpoint string,
	relayerCommission float64,
) (*rpctypes.MakeOfferResponse, error) {
	const (
		method = "net_makeOffer"
	)

	req := &rpctypes.MakeOfferRequest{
		MinAmount:         min,
		MaxAmount:         max,
		ExchangeRate:      types.ExchangeRate(exchangeRate),
		EthAsset:          ethcommon.Address(ethAsset).Hex(),
		RelayerEndpoint:   relayerEndpoint,
		RelayerCommission: relayerCommission,
	}
	res := &rpctypes.MakeOfferResponse{}

	if err := c.Post(method, req, res); err != nil {
		return nil, err
	}

	return res, nil
}
