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

// EventFilterer filters the chain for specific events (logs).
// When it finds a desired log, it puts it into its outbound channel.
type EventFilterer struct {
	ctx         context.Context
	ec          *ethclient.Client
	topic       ethcommon.Hash
	filterQuery eth.FilterQuery
	logCh       chan<- []ethtypes.Log
}

// NewEventFilterer returns a new *EventFilterer.
func NewEventFilterer(
	ctx context.Context,
	ec *ethclient.Client,
	contract ethcommon.Address,
	fromBlock *big.Int,
	topic ethcommon.Hash,
	logCh chan<- []ethtypes.Log,
) *EventFilterer {
	filterQuery := eth.FilterQuery{
		FromBlock: fromBlock,
		Addresses: []ethcommon.Address{contract},
	}

	return &EventFilterer{
		ctx:         ctx,
		ec:          ec,
		topic:       topic,
		filterQuery: filterQuery,
		logCh:       logCh,
	}
}

// Start starts the EventFilterer. It watches the chain for logs.
func (f *EventFilterer) Start() error {
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

			found := []ethtypes.Log{}
			for _, l := range logs {
				if l.Topics[0] != f.topic {
					continue
				}

				found = append(found, l)
			}

			if len(found) != 0 {
				f.logCh <- found
			}

			// the filter inclusive of the latest block when `ToBlock` is nil, so we add 1
			f.filterQuery.FromBlock = big.NewInt(0).Add(currHeader.Number, big.NewInt(1))
			header = currHeader
		}
	}()

	return nil
}
