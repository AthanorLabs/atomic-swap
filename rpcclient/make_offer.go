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
) (string, error) {
	const (
		method = "net_makeOffer"
	)

	req := &rpctypes.MakeOfferRequest{
		MinimumAmount:     min,
		MaximumAmount:     max,
		ExchangeRate:      types.ExchangeRate(exchangeRate),
		EthAsset:          ethcommon.Address(ethAsset).Hex(),
		RelayerEndpoint:   relayerEndpoint,
		RelayerCommission: relayerCommission,
	}
	res := &rpctypes.MakeOfferResponse{}

	if err := c.Post(method, req, res); err != nil {
		return "", err
	}

	return res.ID, nil
}
