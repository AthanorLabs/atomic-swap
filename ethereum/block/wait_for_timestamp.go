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
func WaitForEthBlockAfterTimestamp(ctx context.Context, ec *ethclient.Client, ts int64) (*ethtypes.Header, error) {
	timeDelta := time.Duration(ts-time.Now().Unix()) * time.Second

	// The sleep is safe even if timeDelta is negative. We only optimise for timestamps in the future, but if
	// the timestamp had already passed for some reason, nothing bad happens.
	if err := common.SleepWithContext(ctx, timeDelta); err != nil {
		return nil, err
	}

	// subscribe to new block headers
	headers := make(chan *ethtypes.Header)
	defer close(headers)
	sub, err := ec.SubscribeNewHead(ctx, headers)
	if err != nil {
		return nil, err
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-sub.Err():
			return nil, err
		case header := <-headers:
			if header.Time >= uint64(ts) {
				return header, nil
			}
		}
	}
}
