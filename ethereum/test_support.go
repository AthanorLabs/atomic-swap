// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

//go:build !prod

package contracts

import (
	"context"
	"crypto/ecdsa"
	"sync"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/ethereum/block"
)

//
// FUNCTIONS ONLY FOR UNIT TESTS
//

// these variables should only be accessed by DevDeploySwapCreator
var _swapCreator *SwapCreator
var _swapCreatorAddr *ethcommon.Address
var _swapCreatorAddrMu sync.Mutex

// DevDeploySwapCreator deploys and returns the swapCreator address and contract
// binding for unit tests, returning a cached result if available.
func DevDeploySwapCreator(t *testing.T, ec *ethclient.Client, pk *ecdsa.PrivateKey) (ethcommon.Address, *SwapCreator) {
	ctx := context.Background()
	_swapCreatorAddrMu.Lock()
	defer _swapCreatorAddrMu.Unlock()

	if _swapCreatorAddr == nil {
		txOpts, err := newTXOpts(ctx, ec, pk)
		require.NoError(t, err)

		swapCreatorAddr, tx, swapCreator, err := DeploySwapCreator(txOpts, ec)
		require.NoError(t, err)

		receipt, err := block.WaitForReceipt(ctx, ec, tx.Hash())
		require.NoError(t, err)

		t.Logf("gas cost to deploy SwapCreator.sol: %d (delta %d)",
			receipt.GasUsed, maxSwapCreatorDeployGas-int(receipt.GasUsed))
		require.GreaterOrEqual(t, maxSwapCreatorDeployGas, int(receipt.GasUsed), "deploy SwapCreator")

		_swapCreatorAddr = &swapCreatorAddr
		_swapCreator = swapCreator
	}

	return *_swapCreatorAddr, _swapCreator
}
