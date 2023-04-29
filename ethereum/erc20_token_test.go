// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package contracts

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
)

const (
	maxTestERC20DeployGas = 798226 // internal only const, using long token names will increase this
)

func TestSwapCreator_NewSwap_ERC20(t *testing.T) {
	auth, conn, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	// deploy TestERC20
	erc20Addr, erc20Tx, erc20Contract, err :=
		DeployTestERC20(auth, conn, "Test of the ERC20 Token", "ERC20Token", 18, addr, big.NewInt(9999))
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(context.Background(), conn, erc20Tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy TestERC20.sol: %d (delta %d)",
		receipt.GasUsed, maxTestERC20DeployGas-int(receipt.GasUsed))
	require.GreaterOrEqual(t, maxTestERC20DeployGas, int(receipt.GasUsed))

	testNewSwap(t, types.EthAsset(erc20Addr), erc20Contract)
}

func TestSwapCreator_Claim_ERC20(t *testing.T) {
	auth, conn, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	erc20Addr, erc20Tx, erc20Contract, err := DeployTestERC20(auth, conn, "TestERC20", "TEST", 18, addr, big.NewInt(9999))
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(context.Background(), conn, erc20Tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy TestERC20.sol: %d (delta %d)",
		receipt.GasUsed, maxTestERC20DeployGas-int(receipt.GasUsed))
	require.GreaterOrEqual(t, maxTestERC20DeployGas, int(receipt.GasUsed))

	// 3 logs:
	// Approval
	// Transfer
	// New
	testClaim(t, types.EthAsset(erc20Addr), 2, big.NewInt(99), erc20Contract)
}

func TestSwapCreator_RefundBeforeT0_ERC20(t *testing.T) {
	auth, conn, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	erc20Addr, erc20Tx, erc20Contract, err :=
		DeployTestERC20(auth, conn, "TestERC20", "TEST", 18, addr, big.NewInt(9999))
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(context.Background(), conn, erc20Tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy TestERC20.sol: %d (delta %d)",
		receipt.GasUsed, maxTestERC20DeployGas-int(receipt.GasUsed))
	require.GreaterOrEqual(t, maxTestERC20DeployGas, int(receipt.GasUsed))

	testRefundBeforeT0(t, types.EthAsset(erc20Addr), erc20Contract, 2)
}

func TestSwapCreator_RefundAfterT1_ERC20(t *testing.T) {
	auth, conn, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	erc20Addr, erc20Tx, erc20Contract, err :=
		DeployTestERC20(auth, conn, "TestERC20", "TestERC20", 18, addr, big.NewInt(9999))
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(context.Background(), conn, erc20Tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy TestERC20.sol: %d (delta %d)",
		receipt.GasUsed, maxTestERC20DeployGas-int(receipt.GasUsed))
	require.GreaterOrEqual(t, maxTestERC20DeployGas, int(receipt.GasUsed))

	testRefundAfterT1(t, types.EthAsset(erc20Addr), erc20Contract, 2)
}
