package protocol

import (
	"context"
	"fmt"
	"math/big"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
)

// EthereumAssetAmount represents an amount of an Ethereum asset (ie. ether or an ERC20)
type EthereumAssetAmount interface {
	BigInt() *big.Int
	AsStandard() float64
}

// GetEthereumAssetAmount returns an EthereumAssetAmount (ie EtherAmount or ERC20TokenAmount)
func GetEthereumAssetAmount(
	ctx context.Context,
	ec extethclient.EthClient,
	amt float64,
	asset types.EthAsset,
) (EthereumAssetAmount, error) {
	if asset != types.EthAssetETH {
		_, _, decimals, err := ec.ERC20Info(ctx, asset.Address())
		if err != nil {
			return nil, fmt.Errorf("failed to get ERC20 info: %w", err)
		}

		return common.NewERC20TokenAmountFromDecimals(amt, float64(decimals)), nil
	}

	return common.EtherToWei(amt), nil
}
