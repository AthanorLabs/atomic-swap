package monero

import (
	"context"
	"fmt"
	"time"

	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"

	logging "github.com/ipfs/go-log"
)

const (
	maxRetries = 360
)

var (
	// blockSleepDuration is the duration that we sleep between checks for new blocks. We
	// lower it in dev environments if fast background mining is started.
	blockSleepDuration = time.Second * 10

	log = logging.Logger("monero")
)

// WaitForBlocks waits for `count` new blocks to arrive.
// It returns the height of the chain.
func WaitForBlocks(ctx context.Context, client WalletClient, count int) (uint64, error) {
	prevHeight, err := client.GetHeight()
	if err != nil {
		return 0, fmt.Errorf("failed to get height: %w", err)
	}

	for j := 0; j < count; j++ {
		for i := 0; i < maxRetries; i++ {
			if err := client.Refresh(); err != nil {
				return 0, err
			}

			height, err := client.GetHeight()
			if err != nil {
				continue
			}

			if height > prevHeight {
				return height, nil
			}

			log.Infof("waiting for next block, current height=%d", height)
			if err = common.SleepWithContext(ctx, blockSleepDuration); err != nil {
				return 0, err
			}
		}
	}

	return 0, fmt.Errorf("timed out waiting for blocks")
}

// CreateWallet creates a monero wallet from a private keypair.
func CreateWallet(
	name string,
	env common.Environment,
	client WalletClient,
	kpAB *mcrypto.PrivateKeyPair,
) (mcrypto.Address, error) {
	t := time.Now().Format(common.TimeFmtNSecs)
	walletName := fmt.Sprintf("%s-%s", name, t)
	if err := client.GenerateFromKeys(kpAB, walletName, "", env); err != nil {
		return "", err
	}

	log.Info("created wallet: ", walletName)

	if err := client.Refresh(); err != nil {
		return "", err
	}

	balance, err := client.GetBalance(0)
	if err != nil {
		return "", err
	}

	log.Info("wallet balance: ", balance.Balance)
	return kpAB.Address(env), nil
}
