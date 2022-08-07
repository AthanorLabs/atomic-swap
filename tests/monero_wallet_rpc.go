package tests

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// CreateWalletRPCService starts a monero-wallet-rpc listening on a random port for tests. The json_rpc
// URL of the started service is returned.
func CreateWalletRPCService(t *testing.T) string {
	port := getFreePort(t)
	walletRPCBin := getMoneroWalletRPCBin(t)
	walletRPCBinArgs := getWalletRPCFlags(t, port)
	cmd := exec.Command(walletRPCBin, walletRPCBinArgs...)
	outPipe, err := cmd.StdoutPipe()
	require.NoError(t, err)

	err = cmd.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = outPipe.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})
	scanner := bufio.NewScanner(outPipe)
	started := false
	for scanner.Scan() {
		line := scanner.Text()
		//t.Log(line)
		if strings.HasSuffix(line, "Starting wallet RPC server") {
			started = true
			break
		}
		time.Sleep(200 * time.Millisecond) // additional start time
	}
	if !started {
		t.Fatal("failed to start monero-wallet-rpc")
	}

	// drain any additional output
	go func() {
		for scanner.Scan() {
			//t.Log(scanner.Text())
		}
	}()

	require.NoError(t, err)
	return fmt.Sprintf("http://127.0.0.1:%d/json_rpc", port)
}

// getMoneroWalletRPCBin returns the monero-wallet-rpc binary assuming it was
// installed at the top of the repo in a directory named "monero-bin".
func getMoneroWalletRPCBin(t *testing.T) string {
	_, filename, _, ok := runtime.Caller(0) // this test file path
	require.True(t, ok)
	packageDir := path.Dir(filename)
	repoBaseDir := path.Dir(packageDir)
	return path.Join(repoBaseDir, "monero-bin", "monero-wallet-rpc")
}

// getWalletRPCFlags returns the flags used when launching monero-wallet-rpc in a temporary
// test folder.
func getWalletRPCFlags(t *testing.T, port int) []string {
	walletDir := t.TempDir()
	return []string{
		"--rpc-bind-ip=127.0.0.1",
		fmt.Sprintf("--rpc-bind-port=%d", port),
		"--disable-rpc-login",
		fmt.Sprintf("--log-file=%s", path.Join(walletDir, "monero-wallet-rpc.log")),
		fmt.Sprintf("--wallet-dir=%s", t.TempDir()),
	}
}

// getFreePort returns an OS allocated and immediately freed port. There is nothing preventing
// something else on the system from using the port before the caller has a chance, but OS
// allocated ports are randomised to minimise this risk.
func getFreePort(t *testing.T) int {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	require.NoError(t, ln.Close())
	return port
}
