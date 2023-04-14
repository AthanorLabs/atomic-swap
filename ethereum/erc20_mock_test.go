// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package contracts

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/ethereum/block"
)

func TestSwapCreator_NewSwap_ERC20(t *testing.T) {
	auth, conn, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	// deploy ERC20Mock
	erc20Addr, erc20Tx, erc20Contract, err := DeployERC20Mock(auth, conn, "ERC20Mock", "MOCK", addr, big.NewInt(9999))
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(context.Background(), conn, erc20Tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy ERC20Mock.sol: %d", receipt.GasUsed)

	testNewSwap(t, erc20Addr, erc20Contract)
}

func TestSwapCreator_Claim_ERC20(t *testing.T) {
	auth, conn, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	erc20Addr, erc20Tx, erc20Contract, err := DeployERC20Mock(auth, conn, "ERC20Mock", "MOCK", addr, big.NewInt(9999))
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(context.Background(), conn, erc20Tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy ERC20Mock.sol: %d", receipt.GasUsed)

	// 3 logs:
	// Approval
	// Transfer
	// New
	testClaim(t, erc20Addr, 2, big.NewInt(99), erc20Contract)
}

func TestSwapCreator_RefundBeforeT0_ERC20(t *testing.T) {
	auth, conn, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	erc20Addr, erc20Tx, erc20Contract, err := DeployERC20Mock(auth, conn, "ERC20Mock", "MOCK", addr, big.NewInt(9999))
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(context.Background(), conn, erc20Tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy ERC20Mock.sol: %d", receipt.GasUsed)

	testRefundBeforeT0(t, erc20Addr, erc20Contract, 2)
}

func TestSwapCreator_RefundAfterT1_ERC20(t *testing.T) {
	auth, conn, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	erc20Addr, erc20Tx, erc20Contract, err := DeployERC20Mock(auth, conn, "ERC20Mock", "MOCK", addr, big.NewInt(9999))
	require.NoError(t, err)
	receipt, err := block.WaitForReceipt(context.Background(), conn, erc20Tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy ERC20Mock.sol: %d", receipt.GasUsed)

	testRefundAfterT1(t, erc20Addr, erc20Contract, 2)
}
