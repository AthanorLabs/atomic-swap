// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package contracts

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/common/vjson"
)

// swap is the same as the auto-generated SwapCreatorSwap type, but with some type
// adjustments and annotations for JSON marshalling.
type swap struct {
	Owner            common.Address `json:"owner" validate:"required"`
	Claimer          common.Address `json:"claimer" validate:"required"`
	ClaimCommitment  types.Hash     `json:"claimCommitment" validate:"required"`
	RefundCommitment types.Hash     `json:"refundCommitment" validate:"required"`
	Timeout1         *big.Int       `json:"timeout1" validate:"required"`
	Timeout2         *big.Int       `json:"timeout2" validate:"required"`
	Asset            common.Address `json:"asset"`
	Value            *big.Int       `json:"value" validate:"required"`
	Nonce            *big.Int       `json:"nonce" validate:"required"`
}

// MarshalJSON provides JSON marshalling for SwapCreatorSwap
func (sfs *SwapCreatorSwap) MarshalJSON() ([]byte, error) {
	return vjson.MarshalStruct(&swap{
		Owner:            sfs.Owner,
		Claimer:          sfs.Claimer,
		ClaimCommitment:  sfs.ClaimCommitment,
		RefundCommitment: sfs.RefundCommitment,
		Timeout1:         sfs.Timeout1,
		Timeout2:         sfs.Timeout2,
		Asset:            sfs.Asset,
		Value:            sfs.Value,
		Nonce:            sfs.Nonce,
	})
}

// UnmarshalJSON provides JSON unmarshalling for SwapCreatorSwap
func (sfs *SwapCreatorSwap) UnmarshalJSON(data []byte) error {
	s := &swap{}
	if err := vjson.UnmarshalStruct(data, s); err != nil {
		return err
	}
	*sfs = SwapCreatorSwap{
		Owner:            s.Owner,
		Claimer:          s.Claimer,
		ClaimCommitment:  s.ClaimCommitment,
		RefundCommitment: s.RefundCommitment,
		Timeout1:         s.Timeout1,
		Timeout2:         s.Timeout2,
		Asset:            s.Asset,
		Value:            s.Value,
		Nonce:            s.Nonce,
	}
	return nil
}
