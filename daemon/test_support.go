//go:build !prod

package daemon

import (
	"fmt"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// This file is only for test support. Use the build tag "prod" to prevent
// symbols in this file from consuming space in production binaries.

// WaitForSwapdStart takes the rpcPort of a swapd instance and waits for it to
// be in a listening state. Fails the test if the server isn't listening after a
// little over 60 seconds.
func WaitForSwapdStart(t *testing.T, rpcPort uint16) {
	const maxSeconds = 60
	addr := fmt.Sprintf("127.0.0.1:%d", rpcPort)

	startTime := time.Now()

	for i := 0; i < maxSeconds; i++ {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			startupTime := time.Since(startTime).Round(time.Second)
			t.Logf("daemon on rpc port %d started after %s", rpcPort, startupTime)
			require.NoError(t, conn.Close())
			return
		}
		// DialTimeout doesn't do retries. If the connection was refused, it happened
		// almost immediately, so we still need to sleep.
		require.ErrorIs(t, err, syscall.ECONNREFUSED)
		time.Sleep(time.Second)
	}
	t.Fatalf("giving up, swapd RPC port %d is not listening after %d seconds", rpcPort, maxSeconds)
}
