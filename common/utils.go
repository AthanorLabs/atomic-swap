package common

import (
	"context"
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
	log = logging.Logger("common")
)

func Reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func WaitForReceipt(ctx context.Context, ethclient *ethclient.Client, txHash ethcommon.Hash) (*ethtypes.Receipt, bool) {
	for i := 0; i < maxRetries; i++ {
		receipt, err := ethclient.TransactionReceipt(ctx, txHash)
		if err != nil {
			log.Infof("waiting for transaction to be included in chain: txHash=%s", txHash)
			time.Sleep(receiptSleepDuration)
			continue
		}

		log.Debugf("transaction %s included in chain, block hash=%s, block number=%d, gas used=%d",
			txHash,
			receipt.BlockHash,
			receipt.BlockNumber,
			receipt.CumulativeGasUsed,
		)
		return receipt, true
	}

	return nil, false
}
