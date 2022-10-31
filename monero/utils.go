package monero

import (
	"context"
	"fmt"
	"time"

	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"

	logging "github.com/ipfs/go-log"
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
	startHeight, err := client.GetChainHeight()
	if err != nil {
		return 0, fmt.Errorf("failed to get height: %w", err)
	}
	prevHeight := startHeight - 1 // prevHeight is only for logging
	endHeight := startHeight + uint64(count)

	for {
		if err := client.Refresh(); err != nil {
			return 0, err
		}

		height, err := client.GetChainHeight()
		if err != nil {
			return 0, err
		}

		if height >= endHeight {
			return height, nil
		}

		if height > prevHeight {
			log.Debugf("Waiting for next block, current height %d (target height %d)", height, endHeight)
			prevHeight = height
		}

		if err = common.SleepWithContext(ctx, blockSleepDuration); err != nil {
			return 0, err
		}
	}
}

// CreateWallet creates a monero wallet from a private keypair.
func CreateWallet(
	name string,
	env common.Environment,
	client WalletClient,
	kpAB *mcrypto.PrivateKeyPair,
	restoreHeight uint64,
) (mcrypto.Address, error) {
	t := time.Now().Format(common.TimeFmtNSecs)
	walletName := fmt.Sprintf("%s-%s", name, t)
	if err := client.GenerateFromKeys(kpAB, restoreHeight, walletName, "", env); err != nil {
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
