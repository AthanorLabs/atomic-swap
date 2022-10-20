package common

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestReverse(t *testing.T) {
	in := []byte{0xa, 0xb, 0xc}
	expected := []byte{0xc, 0xb, 0xa}
	require.Equal(t, expected, Reverse(in))
	require.Equal(t, []byte{0xa, 0xb, 0xc}, in) // backing array of original slice is unmodified

	in2 := [3]byte{0xa, 0xb, 0xc}
	require.Equal(t, expected, Reverse(in2[:]))
	require.Equal(t, in2, [3]byte{0xa, 0xb, 0xc}) // input array is unmodified
}

func TestGetTopic(t *testing.T) {
	refundedTopic := ethcommon.HexToHash("0x007c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f")
	require.Equal(t, GetTopic(RefundedEventSignature), refundedTopic)
}

func TestMakeDir(t *testing.T) {
	path := path.Join(t.TempDir(), "mainnet")
	require.NoError(t, MakeDir(path))
	assert.NoError(t, MakeDir(path)) // No error if the dir already exists
	fileStats, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, "drwx------", fileStats.Mode().String()) // only user has access
}

// Checks normal, non-cancelled operation
func TestSleepWithContext_fullSleep(t *testing.T) {
	ctx := context.Background()
	err := SleepWithContext(ctx, -1*time.Hour) // negative duration doesn't sleep or panic
	assert.NoError(t, err)
	err = SleepWithContext(ctx, 10*time.Millisecond)
	assert.NoError(t, err)
}

// Checks that we handle context cancellation and break out of the sleep
func TestSleepWithContext_canceled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := SleepWithContext(ctx, 24*time.Hour) // time out the test if we fail
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
