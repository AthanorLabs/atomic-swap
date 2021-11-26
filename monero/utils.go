package monero

import (
	"fmt"
	"time"

	logging "github.com/ipfs/go-log"
)

const (
	maxRetries         = 360
	blockSleepDuration = time.Second * 10
)

var (
	log = logging.Logger("common")
)

func WaitForBlocks(client Client) error {
	prevHeight, err := client.GetHeight()
	if err != nil {
		return fmt.Errorf("failed to get height: %w", err)
	}

	for i := 0; i < maxRetries; i++ {
		height, err := client.GetHeight()
		if err != nil {
			return fmt.Errorf("failed to get height: %w", err)
		}

		if height > prevHeight {
			return nil
		}

		log.Infof("waiting for next block, current height=%d", height)
		time.Sleep(blockSleepDuration)
	}

	return fmt.Errorf("timed out waiting for next block")
}
