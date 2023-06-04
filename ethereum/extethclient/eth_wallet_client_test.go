// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package extethclient

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/tests"
)

func init() {
	cliutil.SetLogLevels("debug")
}

func Test_ethClient_Transfer(t *testing.T) {
	ctx := context.Background()

	senderKey := tests.GetTestKeyByIndex(t, 0)
	receiverKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	senderEC := CreateTestClient(t, senderKey)
	receiverEC := CreateTestClient(t, receiverKey)

	transferAmt := coins.EtherToWei(coins.StrToDecimal("0.123456789012345678"))

	receipt, err := senderEC.Transfer(ctx, receiverEC.Address(), transferAmt)
	require.NoError(t, err)
	require.Equal(t, receipt.GasUsed, uint64(TransferGas))

	// balance is exactly equal to the transferred amount
	receiverBal, err := receiverEC.Balance(ctx)
	require.NoError(t, err)
	require.Equal(t, receiverBal.AsEtherString(), transferAmt.AsEtherString())
}

func Test_ethClient_Sweep(t *testing.T) {
	ctx := context.Background()
	srcBal := coins.EtherToWei(coins.StrToDecimal("0.5"))

	// We don't want to completely drain a ganache key, so we need to generate a
	// new key for the sweep sender and then fund the account.
	testFunder := tests.GetTestKeyByIndex(t, 0)
	sweepSrcKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	sweepDestKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	funderEC := CreateTestClient(t, testFunder)
	sourceEC := CreateTestClient(t, sweepSrcKey)
	destEC := CreateTestClient(t, sweepDestKey)

	// fund the sweep source account with 0.5 ETH
	_, err = funderEC.Transfer(ctx, sourceEC.Address(), srcBal)
	require.NoError(t, err)

	receipt, err := sourceEC.Sweep(ctx, destEC.Address())
	require.NoError(t, err)
	require.Equal(t, receipt.GasUsed, uint64(TransferGas))

	fees := new(big.Int).Mul(receipt.EffectiveGasPrice, big.NewInt(int64(receipt.GasUsed)))
	expectedDestBal := coins.NewWeiAmount(new(big.Int).Sub(srcBal.BigInt(), fees))

	destBal, err := destEC.Balance(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedDestBal.AsEtherString(), destBal.AsEtherString())
}

// Unfortunately, ganache does not have a mempool, so we can't do a meaningful
// test that does actual cancellation. We just test it as sending a transaction
// that doesn't cancel a nonce in the mempool.
func Test_ethClient_CancelTxWithNonce(t *testing.T) {
	ctx := context.Background()
	pk := tests.GetTestKeyByIndex(t, 0)
	ec := CreateTestClient(t, pk)

	nonce, err := ec.Raw().NonceAt(ctx, ec.Address(), nil)
	require.NoError(t, err)

	gasPrice, err := ec.SuggestGasPrice(ctx)
	require.NoError(t, err)

	receipt, err := ec.CancelTxWithNonce(ctx, nonce, gasPrice)
	require.NoError(t, err)

	require.Equal(t, receipt.EffectiveGasPrice.String(), gasPrice.String())
}

func Test_validateChainID_devSuccess(t *testing.T) {
	err := validateChainID(common.Development, big.NewInt(common.GanacheChainID))
	require.NoError(t, err)
}

func Test_validateChainID_mismatchedEnv(t *testing.T) {
	err := validateChainID(common.Mainnet, big.NewInt(common.GanacheChainID))
	require.Error(t, err)
	assert.ErrorContains(t, err, "expected Mainnet chain ID (1), but found 1337")

	err = validateChainID(common.Stagenet, big.NewInt(common.GanacheChainID))
	require.Error(t, err)
	assert.ErrorContains(t, err, "expected Sepolia chain ID (11155111), but found 1337")
}
