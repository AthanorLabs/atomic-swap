// Copyright 2023 Athanor Labs (ON)
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
	Owner        common.Address `json:"owner" validate:"required"`
	Claimer      common.Address `json:"claimer" validate:"required"`
	PubKeyClaim  types.Hash     `json:"pubKeyClaim" validate:"required"`
	PubKeyRefund types.Hash     `json:"pubKeyRefund" validate:"required"`
	Timeout0     *big.Int       `json:"timeout0" validate:"required"`
	Timeout1     *big.Int       `json:"timeout1" validate:"required"`
	Asset        common.Address `json:"asset"`
	Value        *big.Int       `json:"value" validate:"required"`
	Nonce        *big.Int       `json:"nonce" validate:"required"`
}

// MarshalJSON provides JSON marshalling for SwapCreatorSwap
func (sfs *SwapCreatorSwap) MarshalJSON() ([]byte, error) {
	return vjson.MarshalStruct(&swap{
		Owner:        sfs.Owner,
		Claimer:      sfs.Claimer,
		PubKeyClaim:  sfs.PubKeyClaim,
		PubKeyRefund: sfs.PubKeyRefund,
		Timeout0:     sfs.Timeout0,
		Timeout1:     sfs.Timeout1,
		Asset:        sfs.Asset,
		Value:        sfs.Value,
		Nonce:        sfs.Nonce,
	})
}

// UnmarshalJSON provides JSON unmarshalling for SwapCreatorSwap
func (sfs *SwapCreatorSwap) UnmarshalJSON(data []byte) error {
	s := &swap{}
	if err := vjson.UnmarshalStruct(data, s); err != nil {
		return err
	}
	*sfs = SwapCreatorSwap{
		Owner:        s.Owner,
		Claimer:      s.Claimer,
		PubKeyClaim:  s.PubKeyClaim,
		PubKeyRefund: s.PubKeyRefund,
		Timeout0:     s.Timeout0,
		Timeout1:     s.Timeout1,
		Asset:        s.Asset,
		Value:        s.Value,
		Nonce:        s.Nonce,
	}
	return nil
}
