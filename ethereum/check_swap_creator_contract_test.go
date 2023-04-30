// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package contracts

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/tests"
)

// deployContract is a test helper that deploys the SwapCreator contract and returns the
// deployed address
func deployContract(
	t *testing.T,
	ec *ethclient.Client,
	pk *ecdsa.PrivateKey,
) ethcommon.Address {
	ctx := context.Background()
	contractAddr, _, err := DeploySwapCreatorWithKey(ctx, ec, pk)
	require.NoError(t, err)
	return contractAddr
}

// getContractCode is a test helper that deploys the swap creator contract to read back
// and return the finalised byte code post deployment.
func getContractCode(t *testing.T) []byte {
	ec, _ := tests.NewEthClient(t)
	pk := tests.GetMakerTestKey(t)
	contractAddr := deployContract(t, ec, pk)
	code, err := ec.CodeAt(context.Background(), contractAddr, nil)
	require.NoError(t, err)
	return code
}

// This test will fail if the compiled SwapCreator contract is updated, but the
// expectedSwapCreatorBytecodeHex constant is not updated. Use this test to update the
// constant.
func TestExpectedSwapCreatorBytecodeHex(t *testing.T) {
	codeHex := ethcommon.Bytes2Hex(getContractCode(t))
	require.Equal(t, expectedSwapCreatorBytecodeHex, codeHex,
		"update the expectedSwapCreatorBytecodeHex constant with the actual value to fix this test")
}

// Tests that we fail when the wrong contract byte code is found
func TestCheckSwapCreatorContractCode_fail(t *testing.T) {
	auth, ec, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	// Deploy a non-SwapCreator contract and then try to verify it as SwapCreator contract
	erc20Addr, erc20Tx, _, err := DeployTestERC20(auth, ec, "TestERC20", "TEST", 18, addr, big.NewInt(9999))
	require.NoError(t, err)
	_, err = block.WaitForReceipt(context.Background(), ec, erc20Tx.Hash())
	require.NoError(t, err)

	err = CheckSwapCreatorContractCode(context.Background(), ec, erc20Addr)
	require.ErrorIs(t, err, errInvalidSwapCreatorContract)
}

func TestSepoliaContract(t *testing.T) {
	ctx := context.Background()
	ec := tests.NewEthSepoliaClient(t)

	// temporarily place a funded sepolia private key below to deploy the test contract
	const sepoliaKey = ""

	err := CheckSwapCreatorContractCode(ctx, ec, common.StagenetConfig().SwapCreatorAddr)
	if errors.Is(err, errInvalidSwapCreatorContract) && sepoliaKey != "" {
		pk, err := ethcrypto.HexToECDSA(sepoliaKey) //nolint:govet // shadow declaration of err
		require.NoError(t, err)
		swapCreatorAddr, _, err := DeploySwapCreatorWithKey(ctx, ec, pk)
		require.NoError(t, err)
		t.Fatalf("Update common.StagenetConfig()'s SwapCreatorAddr with %s", swapCreatorAddr.Hex())
	}

	require.NoError(t, err)
}

func TestMainnetContract(t *testing.T) {
	t.Skip("needs to be redeployed before merge")
	ctx := context.Background()
	ec := tests.NewEthMainnetClient(t)
	mainnetConf := common.MainnetConfig()
	err := CheckSwapCreatorContractCode(ctx, ec, mainnetConf.SwapCreatorAddr)
	require.NoError(t, err)
}
