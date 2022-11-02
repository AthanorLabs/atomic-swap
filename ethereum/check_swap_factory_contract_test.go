package contracts

import (
	"bytes"
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/tests"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// deployContract is a test helper that deploys the SwapFactory contract and returns the
// deployed address
func deployContract(t *testing.T, ec *ethclient.Client, trustedForwarder ethcommon.Address) ethcommon.Address {
	pk := tests.GetMakerTestKey(t)
	ctx := context.Background()
	chainID, err := ec.ChainID(ctx)
	require.NoError(t, err)
	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, chainID)
	require.NoError(t, err)
	_, tx, _, err := DeploySwapFactory(txOpts, ec, trustedForwarder)
	require.NoError(t, err)
	contractAddr, err := bind.WaitDeployed(ctx, ec, tx)
	require.NoError(t, err)
	return contractAddr
}

// getContractCode is a test helper that returns the contract bytecode at a given ethereum
// address.
func getContractCode(t *testing.T, trustedForwarder ethcommon.Address) []byte {
	ec, _ := tests.NewEthClient(t)
	contractAddr := deployContract(t, ec, trustedForwarder)
	code, err := ec.CodeAt(context.Background(), contractAddr, nil)
	require.NoError(t, err)
	return code
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
	// arbitrary sentinel address that we search for in the contract byte code
	trustedForwarder := ethcommon.HexToAddress("0x64e902cD8A29bBAefb9D4e2e3A24d8250C606ee7")
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
	trustedForwarderAddresses := []string{
		"0xaAaAaAaaAaAaAaaAaAAAAAAAAaaaAaAaAaaAaaAa",
		"0xbBbBBBBbbBBBbbbBbbBbbbbBBbBbbbbBbBbbBBbB",
		"0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC",
	}
	ec, _ := tests.NewEthClient(t)
	for _, addrHex := range trustedForwarderAddresses {
		tfAddr := ethcommon.HexToAddress(addrHex)
		contractAddr := deployContract(t, ec, tfAddr)
		parsedTFAddr, err := CheckSwapFactoryContractCode(context.Background(), ec, contractAddr)
		require.NoError(t, err)
		require.Equal(t, addrHex, parsedTFAddr.Hex())
	}
}

// Tests that we fail when the wrong contract byte code is found
func TestCheckSwapFactoryContractCode_fail(t *testing.T) {
	ec, chainID := tests.NewEthClient(t)
	pk := tests.GetMakerTestKey(t)
	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, chainID)
	require.NoError(t, err)

	// Deploying an arbitrary contract that won't match the swap factory contract
	contractAddr, tx, _, err :=
		DeployERC20Mock(txOpts, ec, "ERC20Mock", "MOCK", ethcommon.Address{0x1}, big.NewInt(9999))
	require.NoError(t, err)
	_, err = block.WaitForReceipt(context.Background(), ec, tx.Hash())
	require.NoError(t, err)

	_, err = CheckSwapFactoryContractCode(context.Background(), ec, contractAddr)
	require.ErrorIs(t, err, errInvalidSwapContract)
}
