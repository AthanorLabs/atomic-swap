// Package watcher provides tools to track events emitted from ethereum contracts.
package watcher

import (
	"context"
	"math/big"
	"time"

	eth "github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log"
)

var (
	log                   = logging.Logger("ethereum/watcher")
	checkForBlocksTimeout = time.Second
)

// EventFilter filters the chain for specific events (logs).
// When it finds a desired log, it puts it into its outbound channel.
type EventFilter struct {
	ctx         context.Context
	cancel      context.CancelFunc
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

	ctx, cancel := context.WithCancel(ctx)
	return &EventFilter{
		ctx:         ctx,
		cancel:      cancel,
		ec:          ec,
		topic:       topic,
		filterQuery: filterQuery,
		logCh:       logCh,
	}
}

// Start starts the EventFilter. It watches the chain for logs.
func (f *EventFilter) Start() error {
	go func() {
		for {
			select {
			case <-f.ctx.Done():
				return
			case <-time.After(checkForBlocksTimeout):
			}

			currHeader, err := f.ec.HeaderByNumber(f.ctx, nil)
			if err != nil {
				log.Errorf("failed to get header in event watcher: %s", err)
				continue
			}

			if currHeader.Number.Cmp(f.filterQuery.FromBlock) <= 0 {
				// no new blocks, don't do anything
				log.Debugf("EventFilter continue, currHeader %d filterQuery.FromBlock %d",
					currHeader.Number,
					f.filterQuery.FromBlock,
				)
				continue
			}

			// let's see if we have logs
			logs, err := f.ec.FilterLogs(f.ctx, f.filterQuery)
			if err != nil {
				log.Errorf("failed to filter logs for topic %s: %s", f.topic, err)
				continue
			}

			log.Debugf("filtered for logs from block %s to block %s", f.filterQuery.FromBlock, currHeader.Number)

			for _, l := range logs {
				if l.Topics[0] != f.topic {
					continue
				}

				if l.Removed {
					log.Debugf("found removed log: tx hash %s", l.TxHash)
					continue
				}

				log.Debugf("watcher for topic %s found log in block %d", f.topic, l.BlockNumber)
				f.logCh <- l
			}

			f.filterQuery.FromBlock = currHeader.Number
		}
	}()

	return nil
}

// Stop stops the EventFilter.
func (f *EventFilter) Stop() {
	f.cancel()
}
