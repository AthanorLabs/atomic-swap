// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package contracts

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/athanorlabs/go-relayer/impls/gsnforwarder"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/tests"
)

// getContractCode is a test helper that deploys the swap creator contract to read back
// and return the finalised byte code post deployment.
func getContractCode(t *testing.T, forwarderAddr ethcommon.Address) []byte {
	ec, _ := tests.NewEthClient(t)
	pk := tests.GetMakerTestKey(t)
	contractAddr, _ := deploySwapCreatorWithForwarder(t, ec, pk, forwarderAddr)
	code, err := ec.CodeAt(context.Background(), contractAddr, nil)
	require.NoError(t, err)
	return code
}

func TestCheckForwarderContractCode(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	pk := tests.GetMakerTestKey(t)
	forwarderAddr := deployForwarder(t, ec, pk)
	err := CheckForwarderContractCode(context.Background(), ec, forwarderAddr)
	require.NoError(t, err)
}

// This test will fail if the compiled SwapCreator contract is updated, but the
// expectedSwapCreatorBytecodeHex constant is not updated. Use this test to update the
// constant.
func TestExpectedSwapCreatorBytecodeHex(t *testing.T) {
	allZeroTrustedForwarder := ethcommon.Address{}
	codeHex := ethcommon.Bytes2Hex(getContractCode(t, allZeroTrustedForwarder))
	require.Equal(t, expectedSwapCreatorBytecodeHex, codeHex,
		"update the expectedSwapCreatorBytecodeHex constant with the actual value to fix this test")
}

// This test will fail if the compiled SwapCreator contract is updated, but the
// forwarderAddrIndexes slice of trusted forwarder locations is not updated. Use
// this test to update the slice.
func TestForwarderAddrIndexes(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	pk := tests.GetMakerTestKey(t)
	forwarderAddr := deployForwarder(t, ec, pk)
	contactBytes := getContractCode(t, forwarderAddr)

	addressLocations := make([]int, 0) // at the current time, there should always be 2
	for i := 0; i < len(contactBytes)-ethAddrByteLen; i++ {
		if bytes.Equal(contactBytes[i:i+ethAddrByteLen], forwarderAddr[:]) {
			addressLocations = append(addressLocations, i)
			i += ethAddrByteLen - 1 // -1 since the loop will increment by 1
		}
	}

	t.Logf("forwarderAddrIndexes: %v", addressLocations)
	require.EqualValues(t, forwarderAddrIndices, addressLocations,
		"update forwarderAddrIndexes with above logged indexes to fix this test")
}

// Ensure that we correctly verify the SwapCreator contract when initialised with
// different trusted forwarder addresses.
func TestCheckSwapCreatorContractCode(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	pk := tests.GetMakerTestKey(t)
	forwarderAddrs := []string{
		deployForwarder(t, ec, pk).Hex(),
		deployForwarder(t, ec, pk).Hex(),
		deployForwarder(t, ec, pk).Hex(),
	}

	for _, addrHex := range forwarderAddrs {
		tfAddr := ethcommon.HexToAddress(addrHex)
		contractAddr, _ := deploySwapCreatorWithForwarder(t, ec, pk, tfAddr)
		parsedTFAddr, err := CheckSwapCreatorContractCode(context.Background(), ec, contractAddr)
		require.NoError(t, err)
		require.Equal(t, addrHex, parsedTFAddr.Hex())
	}
}

// Tests that we fail when the wrong contract byte code is found
func TestCheckSwapCreatorContractCode_fail(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	pk := tests.GetMakerTestKey(t)

	// Deploy a forwarder contract and then try to verify it as SwapCreator contract
	contractAddr := deployForwarder(t, ec, pk)
	_, err := CheckSwapCreatorContractCode(context.Background(), ec, contractAddr)
	require.ErrorIs(t, err, errInvalidSwapCreatorContract)
}

func TestSepoliaContract(t *testing.T) {
	ctx := context.Background()
	ec := tests.NewEthSepoliaClient(t)

	// temporarily place a funded sepolia private key below to deploy the test contract
	const sepoliaKey = ""

	parsedTFAddr, err := CheckSwapCreatorContractCode(ctx, ec, common.StagenetConfig().SwapCreatorAddr)
	if errors.Is(err, errInvalidSwapCreatorContract) && sepoliaKey != "" {
		pk, err := ethcrypto.HexToECDSA(sepoliaKey) //nolint:govet // shadow declaration of err
		require.NoError(t, err)
		forwarderAddr := ethcommon.HexToAddress(gsnforwarder.SepoliaForwarderAddrHex)
		swapCreatorAddr, _, err := DeploySwapCreatorWithKey(ctx, ec, pk, forwarderAddr)
		require.NoError(t, err)
		t.Fatalf("Update common.StagenetConfig()'s SwapCreatorAddr with %s", swapCreatorAddr.Hex())
	}

	require.NoError(t, err)
	require.Equal(t, gsnforwarder.SepoliaForwarderAddrHex, parsedTFAddr.Hex())
}

func TestMainnetContract(t *testing.T) {
	ctx := context.Background()
	ec := tests.NewEthMainnetClient(t)
	mainnetConf := common.MainnetConfig()
	parsedTFAddr, err := CheckSwapCreatorContractCode(ctx, ec, mainnetConf.SwapCreatorAddr)
	require.NoError(t, err)
	require.Equal(t, gsnforwarder.MainnetForwarderAddrHex, parsedTFAddr.Hex())
}
