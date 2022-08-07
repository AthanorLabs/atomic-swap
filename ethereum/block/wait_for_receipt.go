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
)

const (
	// in total, we will wait up to 1 hour for a transaction to be included
	maxRetries           = 360
	receiptSleepDuration = time.Second * 10
)

var (
	log               = logging.Logger("ethereum/block")
	errReceiptTimeOut = errors.New("failed to get receipt, timed out")
)

// WaitForReceipt waits for the transaction to be mined into a block. If the transaction was reverted when mined,
// we return an error describing why.
func WaitForReceipt(ctx context.Context, ec *ethclient.Client, txHash ethcommon.Hash) (*ethtypes.Receipt, error) {
	for i := 0; i < maxRetries; i++ {
		receipt, err := ec.TransactionReceipt(ctx, txHash)
		if err != nil {
			log.Infof("waiting for transaction to be included in chain: txHash=%s", txHash)
			time.Sleep(receiptSleepDuration)
			continue
		}
		if receipt.Status != ethtypes.ReceiptStatusSuccessful {
			err = fmt.Errorf("transaction failed (gas-lost=%d tx=%s block=%d), %w",
				receipt.GasUsed, txHash, receipt.BlockNumber, errorFromBlock(ctx, ec, receipt))
			return nil, err
		}
		log.Infof("transaction %s included in chain, block hash=%s, block number=%d, gas used=%d",
			txHash,
			receipt.BlockHash,
			receipt.BlockNumber,
			receipt.CumulativeGasUsed,
		)
		return receipt, nil
	}

	return nil, errReceiptTimeOut
}
