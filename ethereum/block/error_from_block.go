// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package block

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ErrorFromBlock returns the error for why a transaction was reverted when mined.
// Normally these errors are detected when creating the transaction, as the contract
// call is simulated to estimate gas, but the state is different in the mined block
// and the transaction can fail (losing gas) after it already went out to the network.
// In this case, we simulate the call using the mined block to extract the error.
func ErrorFromBlock(ctx context.Context, ec *ethclient.Client, receipt *ethtypes.Receipt) error {
	tx, err := ec.TransactionInBlock(ctx, receipt.BlockHash, receipt.TransactionIndex)
	if err != nil {
		return fmt.Errorf("unable to determine error in mined block, %w", err)
	}
	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("unable to determine error in mined block, %w", err)
	}
	txMessage, err := core.TransactionToMessage(tx, ethtypes.LatestSignerForChainID(chainID), nil)
	if err != nil {
		return fmt.Errorf("unable to determine error in mined block, %w", err)
	}

	callMessage := ethereum.CallMsg{
		From:       txMessage.From,
		To:         txMessage.To,
		Gas:        txMessage.GasLimit,
		GasPrice:   txMessage.GasPrice,
		GasFeeCap:  txMessage.GasFeeCap,
		GasTipCap:  txMessage.GasTipCap,
		Value:      txMessage.Value,
		Data:       txMessage.Data,
		AccessList: txMessage.AccessList,
	}
	_, err = ec.CallContract(context.Background(), callMessage, receipt.BlockNumber)
	if err == nil {
		return fmt.Errorf("unable to determine error in mined block")
	}
	return err
}
