package protocol

import (
	"context"
	"fmt"
	"math/big"

	"github.com/cockroachdb/apd/v3"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
)

// EthereumAssetAmount represents an amount of an Ethereum asset (ie. ether or an ERC20)
type EthereumAssetAmount interface {
	BigInt() *big.Int
	AsStandard() *apd.Decimal
}

// GetEthereumAssetAmount returns an EthereumAssetAmount (ie WeiAmount or ERC20TokenAmount)
func GetEthereumAssetAmount(
	ctx context.Context,
	ec extethclient.EthClient,
	amt *apd.Decimal,
	asset types.EthAsset,
) (EthereumAssetAmount, error) {
	if asset != types.EthAssetETH {
		_, _, decimals, err := ec.ERC20Info(ctx, asset.Address())
		if err != nil {
			return nil, fmt.Errorf("failed to get ERC20 info: %w", err)
		}

		return coins.NewERC20TokenAmountFromDecimals(amt, decimals), nil
	}

	return coins.EtherToWei(amt), nil
}
