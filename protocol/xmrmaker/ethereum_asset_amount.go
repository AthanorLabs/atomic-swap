// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package xmrmaker

import (
	"context"
	"fmt"

	"github.com/cockroachdb/apd/v3"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
)

// getEthAssetAmount converts the passed asset amt (in standard units) to
// EthAssetAmount (ie WeiAmount or ERC20TokenAmount)
func getEthAssetAmount(
	ctx context.Context,
	ec extethclient.EthClient,
	amt *apd.Decimal, // in standard units
	asset types.EthAsset,
) (coins.EthAssetAmount, error) {
	if asset.IsToken() {
		token, err := ec.ERC20Info(ctx, asset.Address())
		if err != nil {
			return nil, fmt.Errorf("failed to get ERC20 info: %w", err)
		}

		if coins.ExceedsDecimals(amt, token.NumDecimals) {
			return nil, fmt.Errorf("value can not be represented in the token's %d decimals", token.NumDecimals)
		}

		return coins.NewTokenAmountFromDecimals(amt, token), nil
	}

	if coins.ExceedsDecimals(amt, coins.NumEtherDecimals) {
		return nil, fmt.Errorf("value can not be represented in ETH's %d decimals", coins.NumEtherDecimals)
	}

	return coins.EtherToWei(amt), nil
}
