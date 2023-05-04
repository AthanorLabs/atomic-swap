// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package block

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/tests"
)

// Tests the normal, full flow where we subscribe to new headers and quit after finding
// a header with stamp >= ts.
func TestWaitForEthBlockAfterTimestamp_smallWait(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	ts := time.Now().Add(time.Second) // 1 seconds from now
	ctx := context.Background()
	hdr, err := WaitForEthBlockAfterTimestamp(ctx, ec, ts)
	require.NoError(t, err)
	require.GreaterOrEqual(t, hdr.Time, uint64(ts.Unix()))
}

// Tests context cancellation in the sleep before waiting for any new block headers.
func TestWaitForEthBlockAfterTimestamp_cancelledCtxInSleep(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ts := time.Now().Add(24 * time.Hour) // make the test time out if we don't handle the context
	_, err := WaitForEthBlockAfterTimestamp(ctx, ec, ts)
	require.ErrorIs(t, err, context.Canceled)
}

// Tests context cancellation after sleep while waiting for new block headers.
func TestWaitForEthBlockAfterTimestamp_cancelledCtxWaitingForHeaders(t *testing.T) {
	ec, _ := tests.NewEthClient(t)

	// First we set the ts to now and give a short context timeout. We want to pass
	// the initial sleep and test the context handling in the code receiving new block
	// headers
	ts := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := WaitForEthBlockAfterTimestamp(ctx, ec, ts)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

// Tests failure to subscribe to new block headers
func TestWaitForEthBlockAfterTimestamp_failToSubscribe(t *testing.T) {
	// Using websockets connection so we can close to get more error test coverage
	ec, err := ethclient.Dial(common.DefaultGanacheWSEndpoint)
	require.NoError(t, err)
	t.Cleanup(func() {
		ec.Close()
	})

	ts := time.Now()
	ctx := context.Background()
	ec.Close() // make HeaderByNumber return an error
	_, err = WaitForEthBlockAfterTimestamp(ctx, ec, ts)
	require.Error(t, err)
	require.Contains(t, err.Error(), "closed")
}

// Tests that nothing bad happens, other than waiting for an extra block, if the timestamp
// was in the past
func TestWaitForEthBlockAfterTimestamp_alreadyAfter(t *testing.T) {
	ec, _ := tests.NewEthClient(t)

	ts := time.Now().Add(time.Second * -60) // one minute ago
	ctx := context.Background()
	hdr, err := WaitForEthBlockAfterTimestamp(ctx, ec, ts)
	require.NoError(t, err)
	require.Greater(t, hdr.Time, uint64(ts.Unix())) // ts was minute ago, so strictly greater
}
