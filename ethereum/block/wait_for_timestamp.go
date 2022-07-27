package block

import (
	"context"
	"time"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// sleepWithContext is the same as time.Sleep(...) but with preemption if the context is complete.
func sleepWithContext(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
	case <-timer.C:
	}
}

// WaitForEthBlockAfterTimestamp returns the header of the first block whose timestamp is >= ts.
func WaitForEthBlockAfterTimestamp(ctx context.Context, ec *ethclient.Client, ts int64) (*ethtypes.Header, error) {
	timeDelta := time.Duration(ts-time.Now().Unix()) * time.Second

	// The sleep is safe even if timeDelta is negative. We only optimise for timestamps in the future, but if
	// the timestamp had already passed for some reason, nothing bad happens.
	sleepWithContext(ctx, timeDelta)

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
