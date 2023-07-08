// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package types

import (
	"github.com/cockroachdb/apd/v3"

	"github.com/athanorlabs/atomic-swap/coins"
)

// Pair represents a pair (Such as ETH / XMR)
type Pair struct {
	LiquidityETH *apd.Decimal         `json:"liquidityEth" validate:"required"`
	LiquidityXMR *apd.Decimal         `json:"liquidityXmr" validate:"required"`
	EthAsset     EthAsset             `json:"ethAsset" validate:"required"`
	Token        coins.ERC20TokenInfo `json:"token" validate:"required"`
	Offers       uint64               `json:"offers" validate:"required"`
	Verified     bool                 `json:"verified" valdate:"required"`
}

// NewPair creates and returns a Pair
func NewPair(EthAsset EthAsset) *Pair {
	pair := &Pair{
		LiquidityETH: apd.New(0, 0),
		LiquidityXMR: apd.New(0, 0),
		EthAsset:     EthAsset,

		// Always set to false for now until the verified-list
		// is implemented
		Verified: false,
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

	pair.Offers++

	return nil
}
