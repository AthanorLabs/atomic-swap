// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

//go:build !prod

package contracts

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
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

// variables should only be accessed by GetMockTether
var _mockTether *coins.ERC20TokenInfo
var _mockTetherMu sync.Mutex

// GetMockTether returns the ERC20TokenInfo of a dev token configured with
// similar parameters to Tether.
func GetMockTether(t *testing.T, ec *ethclient.Client, pk *ecdsa.PrivateKey) *coins.ERC20TokenInfo {
	const (
		name        = "Tether USD"
		symbol      = "USDT"
		numDecimals = 6
	)

	_mockTetherMu.Lock()
	defer _mockTetherMu.Unlock()

	// checking the length instead of nil in case a previous run failed
	if _mockTether != nil {
		mintTokens(t, ec, pk, _mockTether)
		return _mockTether
	}

	ownerAddress := common.EthereumPrivateKeyToAddress(pk)

	ctx := context.Background()
	txOpts, err := newTXOpts(ctx, ec, pk)
	require.NoError(t, err)

	supply := calcTokenUnits(1000, numDecimals)
	addr, tx, _, err := DeployTestERC20(txOpts, ec, name, symbol, numDecimals, ownerAddress, supply)
	require.NoError(t, err)
	_, err = block.WaitForReceipt(context.Background(), ec, tx.Hash())
	require.NoError(t, err)

	_mockTether = &coins.ERC20TokenInfo{
		Address:     addr,
		NumDecimals: numDecimals,
		Name:        name,
		Symbol:      symbol,
	}

	return _mockTether
}

// variables should only be accessed by GetMockDAI
var _mockDAI *coins.ERC20TokenInfo
var _mockDAIMu sync.Mutex

// GetMockDAI returns the ERC20TokenInfo of a dev token configured with
// similar parameters to the DAI stablecoin.
func GetMockDAI(t *testing.T, ec *ethclient.Client, pk *ecdsa.PrivateKey) *coins.ERC20TokenInfo {
	const (
		name        = "Dai Stablecoin"
		symbol      = "DAI"
		numDecimals = 18
	)

	_mockDAIMu.Lock()
	defer _mockDAIMu.Unlock()

	// checking the length instead of nil in case a previous run failed
	if _mockDAI != nil {
		mintTokens(t, ec, pk, _mockDAI)
		return _mockDAI
	}

	ownerAddress := common.EthereumPrivateKeyToAddress(pk)

	ctx := context.Background()
	txOpts, err := newTXOpts(ctx, ec, pk)
	require.NoError(t, err)

	supply := calcTokenUnits(1000, numDecimals)
	addr, tx, _, err := DeployTestERC20(txOpts, ec, name, symbol, numDecimals, ownerAddress, supply)
	require.NoError(t, err)
	_, err = block.WaitForReceipt(context.Background(), ec, tx.Hash())
	require.NoError(t, err)

	_mockDAI = &coins.ERC20TokenInfo{
		Address:     addr,
		NumDecimals: numDecimals,
		Name:        name,
		Symbol:      symbol,
	}

	return _mockDAI
}

// calcTokenUnits converts the token's standard units into its internal,
// smallest non-divisible units.
func calcTokenUnits(numStdUnits int64, decimals uint8) *big.Int {
	powerOf10 := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	return new(big.Int).Mul(big.NewInt(numStdUnits), powerOf10)
}

// mintTokens ensures that the account associated with `pk` has at least 1000 tokens
func mintTokens(t *testing.T, ec *ethclient.Client, pk *ecdsa.PrivateKey, token *coins.ERC20TokenInfo) {
	ctx := context.Background()
	bindOpts := &bind.CallOpts{Context: ctx}

	tokenContract, err := NewTestERC20(token.Address, ec)
	require.NoError(t, err)

	decimals, err := tokenContract.Decimals(bindOpts)
	require.NoError(t, err)

	symbol, err := tokenContract.Symbol(bindOpts)
	require.NoError(t, err)
	require.Equal(t, token.Symbol, symbol)

	ownerAddress := common.EthereumPrivateKeyToAddress(pk)

	desiredAmt := calcTokenUnits(1000, decimals)
	currentAmt, err := tokenContract.BalanceOf(bindOpts, ownerAddress)
	require.NoError(t, err)

	if currentAmt.Cmp(desiredAmt) < 0 {
		txOpts, err := newTXOpts(context.Background(), ec, pk)
		require.NoError(t, err)
		mintAmt := new(big.Int).Sub(desiredAmt, currentAmt)
		tx, err := tokenContract.Mint(txOpts, ownerAddress, mintAmt)
		require.NoError(t, err)
		_, err = block.WaitForReceipt(ctx, ec, tx.Hash())
		require.NoError(t, err)
	}
}
