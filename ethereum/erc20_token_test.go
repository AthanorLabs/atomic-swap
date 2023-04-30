// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package contracts

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/tests"
)

func TestSwapCreator_NewSwap_ERC20(t *testing.T) {
	pkA := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)
	addr := crypto.PubkeyToAddress(pkA.PublicKey)

	// deploy TestERC20
	erc20Addr, tx, erc20Contract, err :=
		DeployTestERC20(getAuth(t, pkA), ec, "Test of the ERC20 Token", "ERC20Token", 18, addr, big.NewInt(9999))
	require.NoError(t, err)
	receipt := getReceipt(t, ec, tx)
	t.Logf("gas cost to deploy TestERC20.sol: %d (delta %d)",
		receipt.GasUsed, maxTestERC20DeployGas-int(receipt.GasUsed))
	require.GreaterOrEqual(t, maxTestERC20DeployGas, int(receipt.GasUsed))

	testNewSwap(t, types.EthAsset(erc20Addr), erc20Contract)
}

func TestSwapCreator_Claim_ERC20(t *testing.T) {
	pkA := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)
	addr := crypto.PubkeyToAddress(pkA.PublicKey)

	erc20Addr, tx, erc20Contract, err :=
		DeployTestERC20(getAuth(t, pkA), ec, "TestERC20", "TEST", 18, addr, big.NewInt(9999))
	require.NoError(t, err)
	receipt := getReceipt(t, ec, tx)
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
	pkA := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)
	addr := crypto.PubkeyToAddress(pkA.PublicKey)

	erc20Addr, tx, erc20Contract, err :=
		DeployTestERC20(getAuth(t, pkA), ec, "TestERC20", "TEST", 18, addr, big.NewInt(9999))
	require.NoError(t, err)
	receipt := getReceipt(t, ec, tx)
	t.Logf("gas cost to deploy TestERC20.sol: %d (delta %d)",
		receipt.GasUsed, maxTestERC20DeployGas-int(receipt.GasUsed))
	require.GreaterOrEqual(t, maxTestERC20DeployGas, int(receipt.GasUsed))

	testRefundBeforeT0(t, types.EthAsset(erc20Addr), erc20Contract, 2)
}

func TestSwapCreator_RefundAfterT1_ERC20(t *testing.T) {
	pkA := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)
	addr := crypto.PubkeyToAddress(pkA.PublicKey)

	erc20Addr, tx, erc20Contract, err :=
		DeployTestERC20(getAuth(t, pkA), ec, "TestERC20", "TestERC20", 18, addr, big.NewInt(9999))
	require.NoError(t, err)
	receipt := getReceipt(t, ec, tx)
	t.Logf("gas cost to deploy TestERC20.sol: %d (delta %d)",
		receipt.GasUsed, maxTestERC20DeployGas-int(receipt.GasUsed))
	require.GreaterOrEqual(t, maxTestERC20DeployGas, int(receipt.GasUsed))

	testRefundAfterT1(t, types.EthAsset(erc20Addr), erc20Contract, 2)
}
