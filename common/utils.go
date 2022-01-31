package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

// Reverse reverses the byte slice and returns it.
func Reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// WaitForReceipt waits for the receipt for the given transaction to be available and returns it.
func WaitForReceipt(ctx context.Context, ethclient *ethclient.Client, txHash ethcommon.Hash) (*ethtypes.Receipt, error) {
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
		return receipt, nil
	}

	return nil, errors.New("failed to get receipt, timed out")
}

// WriteContractAddressToFile writes the contract address to a file in the given basepath
func WriteContractAddressToFile(basepath, addr string) error {
	t := time.Now().Format("2006-Jan-2-15:04:05")
	path := fmt.Sprintf("%s-%s.txt", basepath, t)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}

	type addressFileFormat struct {
		Address string
	}

	bz, err := json.Marshal(addressFileFormat{
		Address: addr,
	})
	if err != nil {
		return err
	}

	_, err = file.Write(bz)
	return err
}
