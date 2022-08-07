package block

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/noot/atomic-swap/tests"
)

// Checks normal, non-cancelled operation
func TestSleepWithContext_fullSleep(t *testing.T) {
	ctx := context.Background()
	sleepWithContext(ctx, -1*time.Hour) // negative duration doesn't sleep or panic
	sleepWithContext(ctx, 10*time.Millisecond)
}

// Checks that we handle context cancellation and break out of the sleep
func TestSleepWithContext_canceled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	sleepWithContext(ctx, 24*time.Hour) // time out the test if we fail
}

// Tests the normal, full flow where we subscribe to new headers and quit after finding
// a header with stamp >= ts.
func TestWaitForEthBlockAfterTimestamp_smallWait(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	ts := time.Now().Unix() + 1 // 1 seconds from now
	ctx := context.Background()
	hdr, err := WaitForEthBlockAfterTimestamp(ctx, ec, ts)
	require.NoError(t, err)
	require.GreaterOrEqual(t, hdr.Time, uint64(ts))
}

// Tests context cancellation in the sleep before waiting for any new block headers.
func TestWaitForEthBlockAfterTimestamp_cancelledCtxInSleep(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ts := time.Now().Add(24 * time.Hour).Unix() // make the test time out if we don't handle the context
	_, err := WaitForEthBlockAfterTimestamp(ctx, ec, ts)
	require.ErrorIs(t, err, context.Canceled)
}

// Tests context cancellation after sleep while waiting for new block headers.
func TestWaitForEthBlockAfterTimestamp_cancelledCtxWaitingForHeaders(t *testing.T) {
	ec, _ := tests.NewEthClient(t)

	// First we set the ts to now and give a short context timeout. We want to pass
	// the initial sleep and test the context handling in the code receiving new block
	// headers
	ts := time.Now().Unix()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := WaitForEthBlockAfterTimestamp(ctx, ec, ts)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

// Tests failure to subscribe to new block headers
func TestWaitForEthBlockAfterTimestamp_failToSubscribe(t *testing.T) {
	ec, _ := tests.NewEthClient(t)

	ts := time.Now().Unix()
	ctx := context.Background()
	ec.Close() // make SubscribeNewHead return an error
	_, err := WaitForEthBlockAfterTimestamp(ctx, ec, ts)
	require.Error(t, err)
	require.Contains(t, err.Error(), "closed")
}

// Tests that nothing bad happens, other than waiting for an extra block, if the timestamp
// was in the past
func TestWaitForEthBlockAfterTimestamp_alreadyAfter(t *testing.T) {
	ec, _ := tests.NewEthClient(t)

	ts := time.Now().Unix() - 60 // one minute ago
	ctx := context.Background()
	hdr, err := WaitForEthBlockAfterTimestamp(ctx, ec, ts)
	require.NoError(t, err)
	require.Greater(t, hdr.Time, uint64(ts)) // ts was minute ago, so strictly greater
}
