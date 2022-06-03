package common

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
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

	errReceiptTimeOut = errors.New("failed to get receipt, timed out")
)

// Reverse reverses the byte slice and returns it.
func Reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// WaitForReceipt waits for the receipt for the given transaction to be available and returns it.
func WaitForReceipt(ctx context.Context, ethclient *ethclient.Client, txHash ethcommon.Hash) (*ethtypes.Receipt, error) { //nolint:lll
	for i := 0; i < maxRetries; i++ {
		receipt, err := ethclient.TransactionReceipt(ctx, txHash)
		if err != nil {
			log.Infof("waiting for transaction to be included in chain: txHash=%s", txHash)
			time.Sleep(receiptSleepDuration)
			continue
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

// EthereumPrivateKeyToAddress returns the address associated with a private key
func EthereumPrivateKeyToAddress(privkey *ecdsa.PrivateKey) ethcommon.Address {
	pub := privkey.Public().(*ecdsa.PublicKey)
	return ethcrypto.PubkeyToAddress(*pub)
}
