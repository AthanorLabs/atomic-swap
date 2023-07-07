// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package types

import (
	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

type Pair struct {
	LiquidityETH *apd.Decimal      `json:"liquidityETH" validate:"required"`
	LiquidityXMR *apd.Decimal      `json:"liquidityXMR" validate:"required"`
	Asset        EthAsset          `json:"asset"`
	Address      ethcommon.Address `json:"address"`
	Offers       uint64            `json:"offers" validate:"required"`
	Verified     bool              `json:"verified" valdate:"required"`
}

func NewPair() *Pair {
	pair := &Pair{
		LiquidityETH: apd.New(0, 0),
		LiquidityXMR: apd.New(0, 0),
	}
	return pair
}

// AddOffer adds an offer to a pair
func (pair *Pair) AddOffer(o *Offer) error {
	_, err := coins.DecimalCtx().Add(pair.LiquidityXMR, pair.LiquidityXMR, o.MaxAmount)
	if err != nil {
		return err
	}

	// Max Amount converted in ETH/Token
	MaxAmountETH, err := o.ExchangeRate.ToETH(o.MaxAmount)
	if err != nil {
		return err
	}

	_, err = coins.DecimalCtx().Add(pair.LiquidityETH, pair.LiquidityETH, MaxAmountETH)
	if err != nil {
		return err
	}

	pair.Offers += 1
	pair.Address = o.EthAsset.Address()

	// Always set to false for now until the verified-list
	// is implemented
	pair.Verified = false

	return nil
}
