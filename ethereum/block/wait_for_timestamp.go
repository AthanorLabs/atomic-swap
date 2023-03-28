// Package block contains ethereum helper methods that center around blocks, like waiting
// for a certain block timestamp, waiting for a transaction to be mined in a block, and
// extracting an error for a transaction from the block that mined it.
package block

import (
	"context"
	"time"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/athanorlabs/atomic-swap/common"
)

// WaitForEthBlockAfterTimestamp returns the header of the first block whose timestamp is >= ts.
func WaitForEthBlockAfterTimestamp(ctx context.Context, ec *ethclient.Client, ts time.Time) (*ethtypes.Header, error) {
	timeDelta := time.Until(ts)
	if timeDelta < 0 {
		timeDelta = 0
	}

	// The sleep is safe even if timeDelta is negative. We only optimise for timestamps in the future, but if
	// the timestamp had already passed for some reason, nothing bad happens.
	if err := common.SleepWithContext(ctx, timeDelta); err != nil {
		return nil, err
	}

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			header, err := ec.HeaderByNumber(ctx, nil)
			if err != nil {
				return nil, err
			}

			if header.Time >= uint64(ts.Unix()) {
				return header, nil
			}
		}
	}
}
