// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package backend

import (
	"context"
	"math/big"
	"testing"

	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/tests"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestWaitForReceipt(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	privKey := tests.GetTakerTestKey(t)
	ec := extethclient.CreateTestClient(t, privKey)

	gasPrice, err := ec.Raw().SuggestGasPrice(ctx)
	require.NoError(t, err)

	to := ethcommon.Address{}
	txInner := &ethtypes.LegacyTx{
		To:       &to,
		Value:    big.NewInt(99),
		Gas:      21000,
		GasPrice: gasPrice,
	}

	tx, err := ethtypes.SignNewTx(privKey,
		ethtypes.LatestSignerForChainID(ec.ChainID()),
		txInner,
	)
	require.NoError(t, err)

	err = ec.Raw().SendTransaction(ctx, tx)
	require.NoError(t, err)

	b := &backend{
		ethClient: ec,
	}

	receipt, err := b.ETHClient().WaitForReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	require.Equal(t, tx.Hash(), receipt.TxHash)
}
