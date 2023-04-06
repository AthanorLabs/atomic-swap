// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package block

import (
	"context"
	"errors"
	"fmt"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log"

	"github.com/athanorlabs/atomic-swap/common"
)

const (
	// in total, we will wait up to 1 hour for a transaction to be included
	maxRetries           = 360
	receiptSleepDuration = time.Second * 2
)

var (
	log               = logging.Logger("ethereum/block")
	errReceiptTimeOut = errors.New("failed to get receipt, timed out")
)

// WaitForReceipt waits for the transaction to be mined into a block. If the transaction was reverted when mined,
// we return an error describing why.
func WaitForReceipt(ctx context.Context, ec *ethclient.Client, txHash ethcommon.Hash) (*ethtypes.Receipt, error) {
	for i := 0; i < maxRetries; i++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		receipt, err := ec.TransactionReceipt(ctx, txHash)
		if err != nil {
			log.Infof("waiting for transaction to be included in chain: txHash=%s", txHash)
			if err = common.SleepWithContext(ctx, receiptSleepDuration); err != nil {
				return nil, err
			}
			continue
		}
		if receipt.Status != ethtypes.ReceiptStatusSuccessful {
			err = fmt.Errorf("failed transaction included in block (%s): %w",
				common.ReceiptInfo(receipt), ErrorFromBlock(ctx, ec, receipt))
			return nil, err
		}
		log.Debugf("transaction included in chain %s", common.ReceiptInfo(receipt))
		return receipt, nil
	}

	return nil, errReceiptTimeOut
}
