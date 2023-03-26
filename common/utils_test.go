package common

import (
	"context"
	"io/fs"
	"math"
	"os"
	"path"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
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

func TestEthereumPrivateKeyToAddress(t *testing.T) {
	// Using the 0th deterministic ganache account/key as the test case
	const ethAddressHex = "0x90F8bf6A479f320ead074411a4B0e7944Ea8c9C1"
	const ethKeyHex = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"

	ethKey, err := ethcrypto.HexToECDSA(ethKeyHex)
	require.NoError(t, err)
	addr := EthereumPrivateKeyToAddress(ethKey)
	require.Equal(t, ethAddressHex, addr.String())
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

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	presentFile := path.Join(tmpDir, "file-is-here.txt")
	missingFile := path.Join(tmpDir, "no-file-here.txt")
	noAccessFile := path.Join(tmpDir, "no-access", "any-file.txt")

	// file exists
	require.NoError(t, os.WriteFile(presentFile, nil, 0600))
	exists, err := FileExists(presentFile)
	require.NoError(t, err)
	assert.True(t, exists)

	// file does not exist
	exists, err = FileExists(missingFile)
	require.NoError(t, err)
	assert.False(t, exists)

	// no access to know if the file exists
	require.NoError(t, os.Mkdir(path.Dir(noAccessFile), 0000)) // no access permissions on dir
	_, err = FileExists(noAccessFile)
	require.ErrorIs(t, err, fs.ErrPermission)

	// path present, but it is a directory instead of a file
	_, err = FileExists(tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "directory")
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

func TestGetFreeTCPPort(t *testing.T) {
	port, err := GetFreeTCPPort()
	require.NoError(t, err)
	require.GreaterOrEqual(t, port, uint(1024))
	require.LessOrEqual(t, port, uint(math.MaxUint16))
}
