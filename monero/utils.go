package monero

import (
	"fmt"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero/crypto"

	logging "github.com/ipfs/go-log"
)

const (
	maxRetries         = 360
	blockSleepDuration = time.Second * 10
)

var (
	log = logging.Logger("monero")
)

// WaitForBlocks waits for a new block to arrive.
func WaitForBlocks(client Client) error {
	prevHeight, err := client.GetHeight()
	if err != nil {
		return fmt.Errorf("failed to get height: %w", err)
	}

	for i := 0; i < maxRetries; i++ {
		height, err := client.GetHeight()
		if err != nil {
			continue
		}

		if height > prevHeight {
			return nil
		}

		log.Infof("waiting for next block, current height=%d", height)
		time.Sleep(blockSleepDuration)
	}

	return fmt.Errorf("timed out waiting for next block")
}

// CreateMoneroWallet creates a monero wallet from a private keypair.
func CreateMoneroWallet(name string, env common.Environment, client Client,
	kpAB *crypto.PrivateKeyPair) (crypto.Address, error) {
	t := time.Now().Format("2006-Jan-2-15:04:05")
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
