// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package contracts

import (
	"context"
	"errors"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/tests"
)

// getContractCode is a test helper that deploys the swap creator contract to read back
// and return the finalised byte code post deployment.
func getContractCode(t *testing.T) []byte {
	ec, _ := tests.NewEthClient(t)
	pk := tests.GetMakerTestKey(t)
	contractAddr, _ := deploySwapCreator(t, ec, pk)
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

// Ensure that we correctly verify the SwapCreator contract when initialised with
// different trusted forwarder addresses.
func TestCheckSwapCreatorContractCode(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	pk := tests.GetMakerTestKey(t)

	contractAddr, _ := deploySwapCreator(t, ec, pk)
	err := CheckSwapCreatorContractCode(context.Background(), ec, contractAddr)
	require.NoError(t, err)
}

// Tests that we fail when the wrong contract byte code is found
func TestCheckSwapCreatorContractCode_fail(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	pk := tests.GetMakerTestKey(t)

	// Deploy a token contract and then try to verify it as SwapCreator contract
	contractAddr, _ := deployERC20Token(t, ec, pk, "name", "symbol", 10, 100)
	err := CheckSwapCreatorContractCode(context.Background(), ec, contractAddr)
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
