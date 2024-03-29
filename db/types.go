// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package db

import (
	"math/big"

	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

// EthereumSwapInfo represents information required on the Ethereum side in case of recovery
type EthereumSwapInfo struct {
	// StartNumber the block number of the `newSwap` transaction. The same for
	// both maker/taker.
	StartNumber *big.Int `json:"startNumber" validate:"required"`

	// SwapID is the swap ID used by the swap contract; not the same as the
	// swap/offer ID used by swapd. It's the hash of the ABI encoded
	// `contracts.SwapCreatorSwap` struct.
	SwapID types.Hash `json:"swapID" validate:"required"`

	// Swap is the `Swap` structure inside SwapCreator.sol.
	Swap *contracts.SwapCreatorSwap `json:"swap" validate:"required"`

	// SwapCreatorAddr is the address of the contract on which the swap was created.
	SwapCreatorAddr ethcommon.Address `json:"swapCreatorAddr" validate:"required"`
}
