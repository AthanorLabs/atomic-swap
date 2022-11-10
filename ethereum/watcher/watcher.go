package watcher

import (
	"context"
	"math/big"
	"time"

	eth "github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var checkForBlocksTimeout = time.Second

// EventFilter filters the chain for specific events (logs).
// When it finds a desired log, it puts it into its outbound channel.
type EventFilter struct {
	ctx         context.Context
	ec          *ethclient.Client
	topic       ethcommon.Hash
	filterQuery eth.FilterQuery
	logCh       chan<- ethtypes.Log
}

// NewEventFilter returns a new *EventFilter.
func NewEventFilter(
	ctx context.Context,
	ec *ethclient.Client,
	contract ethcommon.Address,
	fromBlock *big.Int,
	topic ethcommon.Hash,
	logCh chan<- ethtypes.Log,
) *EventFilter {
	filterQuery := eth.FilterQuery{
		FromBlock: fromBlock,
		Addresses: []ethcommon.Address{contract},
	}

	return &EventFilter{
		ctx:         ctx,
		ec:          ec,
		topic:       topic,
		filterQuery: filterQuery,
		logCh:       logCh,
	}
}

// Start starts the EventFilter. It watches the chain for logs.
func (f *EventFilter) Start() error {
	header, err := f.ec.HeaderByNumber(f.ctx, nil)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-f.ctx.Done():
				return
			case <-time.After(checkForBlocksTimeout):
			}

			currHeader, err := f.ec.HeaderByNumber(f.ctx, nil)
			if err != nil {
				continue
			}

			if currHeader.Number.Cmp(header.Number) <= 0 {
				// no new blocks, don't do anything
				continue
			}

			// let's see if we have logs
			logs, err := f.ec.FilterLogs(f.ctx, f.filterQuery)
			if err != nil {
				continue
			}

			for _, l := range logs {
				if l.Topics[0] != f.topic {
					continue
				}

				if l.Removed {
					continue
				}

				f.logCh <- l
			}

			// the filter inclusive of the latest block when `ToBlock` is nil, so we add 1
			f.filterQuery.FromBlock = big.NewInt(0).Add(currHeader.Number, big.NewInt(1))
			header = currHeader
		}
	}()

	return nil
}
