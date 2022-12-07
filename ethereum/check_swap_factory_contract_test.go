package contracts

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/tests"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// deployContract is a test helper that deploys the SwapFactory contract and returns the
// deployed address
func deployContract(
	t *testing.T,
	ec *ethclient.Client,
	pk *ecdsa.PrivateKey,
	trustedForwarder ethcommon.Address,
) ethcommon.Address {
	ctx := context.Background()
	contractAddr, _, err := DeploySwapFactoryWithKey(ctx, ec, pk, trustedForwarder)
	require.NoError(t, err)
	return contractAddr
}

func deployForwarder(t *testing.T, ec *ethclient.Client, pk *ecdsa.PrivateKey) ethcommon.Address {
	addr, err := DeployGSNForwarderWithKey(context.Background(), ec, pk)
	require.NoError(t, err)
	return addr
}

// getContractCode is a test helper that deploys the swap factory contract to read back
// and return the finalised byte code post deployment.
func getContractCode(t *testing.T, trustedForwarder ethcommon.Address) []byte {
	ec, _ := tests.NewEthClient(t)
	pk := tests.GetMakerTestKey(t)
	contractAddr := deployContract(t, ec, pk, trustedForwarder)
	code, err := ec.CodeAt(context.Background(), contractAddr, nil)
	require.NoError(t, err)
	return code
}

func TestCheckForwarderContractCode(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	pk := tests.GetMakerTestKey(t)
	trustedForwarder := deployForwarder(t, ec, pk)
	err := checkForwarderContractCode(context.Background(), ec, trustedForwarder)
	require.NoError(t, err)
}

// This test will fail if the compiled SwapFactory contract is updated, but the
// expectedSwapFactoryBytecodeHex constant is not updated. Use this test to update the
// constant.
func TestExpectedSwapFactoryBytecodeHex(t *testing.T) {
	allZeroTrustedForwarder := ethcommon.Address{}
	codeHex := ethcommon.Bytes2Hex(getContractCode(t, allZeroTrustedForwarder))
	require.Equal(t, codeHex, expectedSwapFactoryBytecodeHex,
		"update the expectedSwapFactoryBytecodeHex constant with the expected value to fix this test")
}

// This test will fail if the compiled SwapFactory contract is updated, but the
// forwarderAddressIndexes slice of trusted forwarder locations is not updated. Use this
// test to update the slice.
func TestForwarderAddressIndexes(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	pk := tests.GetMakerTestKey(t)
	trustedForwarder := deployForwarder(t, ec, pk)
	contactBytes := getContractCode(t, trustedForwarder)

	addressLocations := make([]int, 0) // at the current time, there should always be 2
	for i := 0; i < len(contactBytes)-ethAddrByteLen; i++ {
		if bytes.Equal(contactBytes[i:i+ethAddrByteLen], trustedForwarder[:]) {
			addressLocations = append(addressLocations, i)
			i += ethAddrByteLen - 1 // -1 since the loop will increment by 1
		}
	}

	t.Logf("forwarderAddressIndexes: %v", addressLocations)
	require.EqualValues(t, forwarderAddressIndices, addressLocations,
		"update forwarderAddressIndexes with above logged indexes to fix this test")
}

// Ensure that we correctly verify the SwapFactory contract when initialised with
// different trusted forwarder addresses.
func TestCheckSwapFactoryContractCode(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	pk := tests.GetMakerTestKey(t)
	trustedForwarderAddresses := []string{
		deployForwarder(t, ec, pk).Hex(),
		deployForwarder(t, ec, pk).Hex(),
		deployForwarder(t, ec, pk).Hex(),
	}

	for _, addrHex := range trustedForwarderAddresses {
		tfAddr := ethcommon.HexToAddress(addrHex)
		contractAddr := deployContract(t, ec, pk, tfAddr)
		parsedTFAddr, err := CheckSwapFactoryContractCode(context.Background(), ec, contractAddr)
		require.NoError(t, err)
		require.Equal(t, addrHex, parsedTFAddr.Hex())
	}
}

// Tests that we fail when the wrong contract byte code is found
func TestCheckSwapFactoryContractCode_fail(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	pk := tests.GetMakerTestKey(t)

	// Deploy a forwarder contract and then try to verify it as SwapFactory contract
	contractAddr := deployForwarder(t, ec, pk)
	_, err := CheckSwapFactoryContractCode(context.Background(), ec, contractAddr)
	require.ErrorIs(t, err, errInvalidSwapContract)
}

func TestGoerliContract(t *testing.T) {
	// comment out the next line to test the default goerli contract
	t.Skip("requires access to non-vetted external goerli node")
	const goerliEndpoint = "https://ethereum-goerli-rpc.allthatnode.com"
	// temporarily place a funded goerli private key below to deploy the test contract
	const goerliKey = ""
	ec, err := ethclient.Dial(goerliEndpoint)
	require.NoError(t, err)
	defer ec.Close()

	parsedTFAddr, err := CheckSwapFactoryContractCode(context.Background(), ec, common.StagenetConfig.ContractAddress)
	if errors.Is(err, errInvalidSwapContract) && goerliKey != "" {
		pk, err := ethcrypto.HexToECDSA(goerliKey) //nolint:govet // shadow declaration of err
		require.NoError(t, err)
		forwarderAddr := deployForwarder(t, ec, pk)
		sfAddr, _, err := DeploySwapFactoryWithKey(context.Background(), ec, pk, forwarderAddr)
		require.NoError(t, err)
		t.Logf("New Goerli SwapFactory deployed with TrustedForwarder=%s", forwarderAddr)
		t.Fatalf("Update common.StagenetConfig.ContractAddress with %s", sfAddr.Hex())
	} else {
		require.NoError(t, err)
		t.Logf("Goerli SwapFactory deployed with TrustedForwarder=%s", parsedTFAddr.Hex())
	}
}
