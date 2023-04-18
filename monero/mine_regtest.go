// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package monero

import (
	"context"
	"sync"
	"time"

	"github.com/MarinX/monerorpc"
	"github.com/MarinX/monerorpc/daemon"

	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
)

const (
	// MonerodRegtestEndpoint is the RPC endpoint used by monerod in the dev environment's regtest mode.
	MonerodRegtestEndpoint = "http://127.0.0.1:18081/json_rpc"

	backgroundMineInterval = 1 * time.Second

	errBlockNotAccepted = "Block not accepted"
)

var mineMu sync.Mutex

// BackgroundMineBlocks starts a background go routine to mine blocks in a monerod instance
// that is in regtest mode. If there is an existing go routine that is already mining from
// a previous call, no new go routine is created.
func BackgroundMineBlocks(ctx context.Context, blockRewardAddress *mcrypto.Address) {
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()

	// Lower the sleep duration used by WaitForBlock
	blockSleepDuration = backgroundMineInterval / 3
	go func() {
		defer wg.Done()
		if !mineMu.TryLock() {
			return // If there are multiple clients in a test, only let one of them mine.
		}
		defer mineMu.Unlock()

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(backgroundMineInterval):
				// not cancelled, mine another block below
			}

			daemonCli := monerorpc.New(MonerodRegtestEndpoint, nil).Daemon
			resp, err := daemonCli.GenerateBlocks(&daemon.GenerateBlocksRequest{
				AmountOfBlocks: 1,
				WalletAddress:  blockRewardAddress.String(),
			})
			if err != nil && err.Error() == errBlockNotAccepted {
				// This probably happens when something else is simultaneously generating
				// blocks, not an error that matters unless it is happening frequently.
				continue
			} else if err != nil {
				log.Warnf("Failed to mine block: %s", err)
			}
			if false { // change to true if debugging and you want to see when new blocks are generated
				log.Debugf("Background mined 1 monero block at height=%d", resp.Height)
			}
		}
	}()
}
