// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package contracts

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/tests"
)

func deployERC20Token(
	t *testing.T,
	ec *ethclient.Client,
	pk *ecdsa.PrivateKey, // token owner (and pays for deployment)
	name string,
	symbol string,
	decimals uint8,
	supplyStdUnits int64,
) (ethcommon.Address, *TestERC20) {
	addr := crypto.PubkeyToAddress(pk.PublicKey)
	supply := new(big.Int).Mul(big.NewInt(supplyStdUnits),
		new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil),
	)

	tokenAddr, tx, tokenContract, err :=
		DeployTestERC20(getAuth(t, pk), ec, name, symbol, decimals, addr, supply)
	require.NoError(t, err)
	receipt := getReceipt(t, ec, tx)

	t.Logf("gas cost to deploy TestERC20.sol: %d (delta %d)",
		receipt.GasUsed, maxTestERC20DeployGas-int(receipt.GasUsed))
	require.GreaterOrEqual(t, maxTestERC20DeployGas, int(receipt.GasUsed))

	return tokenAddr, tokenContract
}

func TestSwapCreator_NewSwap_ERC20(t *testing.T) {
	pkA := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)

	tokenAddr, tokenContract := deployERC20Token(
		t,
		ec,
		pkA,
		"Test of the ERC20 Token",
		"ERC20Token",
		18,
		9999,
	)

	testNewSwap(t, types.EthAsset(tokenAddr), tokenContract)
}

func TestSwapCreator_Claim_ERC20(t *testing.T) {
	pkA := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)

	tokenAddr, tokenContract := deployERC20Token(
		t,
		ec,
		pkA,
		"TestERC20",
		"TEST",
		18,
		9999,
	)

	// 3 logs:
	// Approval
	// Transfer
	// New
	testClaim(t, types.EthAsset(tokenAddr), 2, big.NewInt(99), tokenContract)
}

func TestSwapCreator_RefundBeforeT0_ERC20(t *testing.T) {
	pkA := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)

	tokenAddr, tokenContract := deployERC20Token(
		t,
		ec,
		pkA,
		"TestERC20",
		"TEST",
		18,
		9999,
	)

	testRefundBeforeT0(t, types.EthAsset(tokenAddr), tokenContract, 2)
}

func TestSwapCreator_RefundAfterT1_ERC20(t *testing.T) {
	pkA := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)

	tokenAddr, tokenContract := deployERC20Token(
		t,
		ec,
		pkA,
		"TestERC20",
		"TEST",
		18,
		9999,
	)

	testRefundAfterT1(t, types.EthAsset(tokenAddr), tokenContract, 2)
}
