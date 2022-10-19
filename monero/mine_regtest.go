package monero

import (
	"context"
	"sync"
	"time"

	"github.com/MarinX/monerorpc"
	"github.com/MarinX/monerorpc/daemon"
)

const (
	// MonerodRegtestEndpoint is the RPC endpoint used by monerod in the dev environment's regtest mode.
	MonerodRegtestEndpoint = "http://127.0.0.1:18081/json_rpc"

	backgroundMineInterval = 3 * time.Second

	errBlockNotAccepted = "Block not accepted"
)

var mineMu sync.Mutex

// BackgroundMineBlocks starts a background go routine to mine blocks in a monerod instance
// that is in regtest mode. If there is an existing go routine that is already mining from
// a previous call, no new go routine is created.
func BackgroundMineBlocks(ctx context.Context, blockRewardAddress string) {
	var wg sync.WaitGroup
	wg.Add(1)

	defer wg.Wait()

	go func() {
		defer wg.Done()
		if !mineMu.TryLock() {
			return // If there are multiple clients in a test, only let one of them mine.
		}
		defer mineMu.Unlock()

		for {
			time.Sleep(backgroundMineInterval)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backgroundMineInterval):
				// not cancelled, mine another block below
			}

			daemonCli := monerorpc.New(MonerodRegtestEndpoint, nil).Daemon
			_, err := daemonCli.GenerateBlocks(&daemon.GenerateBlocksRequest{
				AmountOfBlocks: 1,
				WalletAddress:  blockRewardAddress,
			})
			if err != nil && err.Error() == errBlockNotAccepted {
				// This probably happens when something else is simultaneously generating
				// blocks, not an error that matters unless it is happening frequently.
				continue
			} else if err != nil {
				log.Warnf("failed to mine block: %s", err)
			}

			log.Debugf("background mined 1 monero block")
		}
	}()
}
