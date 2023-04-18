// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package protocol

import (
	"context"
	"fmt"

	"github.com/cockroachdb/apd/v3"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
)

// GetEthAssetAmount converts the passed asset amt (in standard units) to
// EthAssetAmount (ie WeiAmount or ERC20TokenAmount)
func GetEthAssetAmount(
	ctx context.Context,
	ec extethclient.EthClient,
	amt *apd.Decimal, // in standard units
	asset types.EthAsset,
) (coins.EthAssetAmount, error) {
	if asset != types.EthAssetETH {
		tokenInfo, err := ec.ERC20Info(ctx, asset.Address())
		if err != nil {
			return nil, fmt.Errorf("failed to get ERC20 info: %w", err)
		}

		return coins.NewERC20TokenAmountFromDecimals(amt, tokenInfo), nil
	}

	return coins.EtherToWei(amt), nil
}
