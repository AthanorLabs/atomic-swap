package tests

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/noot/atomic-swap/cmd/client/client"
	"github.com/noot/atomic-swap/common"

	"github.com/stretchr/testify/require"
)

const (
	defaultAliceMultiaddr      = "/ip4/127.0.0.1/tcp/9933/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2"
	defaultAliceTestLibp2pKey  = "alice.key"
	defaultAliceDaemonEndpoint = "http://localhost:5001"
	defaultBobDaemonEndpoint   = "http://localhost:5002"

	aliceProvideAmount = float64(33.3)
	bobProvideAmount   = float64(44.4)
)

func TestMain(m *testing.M) {
	cmd := exec.Command("../scripts/build.sh")
	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func startSwapDaemon(t *testing.T, done <-chan struct{}, args ...string) *exec.Cmd {
	cmd := exec.Command("../swapd", args...)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	errCh := make(chan error)
	go func() {
		err := cmd.Run()
		if err != nil {
			errCh <- err
		}

		wg.Done()
	}()

	go func() {
		defer wg.Done()

		for {
			select {
			case <-done:
				_ = cmd.Process.Kill()
				_ = cmd.Wait()
				// drain errCh
				<-errCh
				return
			case err := <-errCh:
				select {
				case <-done:
					return
				}
				t.Fatalf("program exited early: %s", err)
			}
		}
	}()

	t.Cleanup(func() {
		wg.Wait()
	})

	time.Sleep(time.Second * 5)
	return cmd
}

func startAlice(t *testing.T, done <-chan struct{}) *exec.Cmd {
	return startSwapDaemon(t, done, "--alice",
		"--max-amount", fmt.Sprintf("%v", aliceProvideAmount),
		"--libp2p-key", defaultAliceTestLibp2pKey,
	)
}

func startBob(t *testing.T, done <-chan struct{}) *exec.Cmd {
	return startSwapDaemon(t, done, "--bob",
		"--max-amount", fmt.Sprintf("%v", bobProvideAmount),
		"--bootnodes", defaultAliceMultiaddr,
		"--wallet-file", "test-wallet",
	)
}

// charlie doesn't provide any coin or participate in any swap.
// he is just a node running the p2p protocol.
func startCharlie(t *testing.T, done <-chan struct{}) *exec.Cmd {
	return startSwapDaemon(t, done,
		"--libp2p-port", "9955",
		"--rpc-port", "5003",
		"--bootnodes", defaultAliceMultiaddr)
}

func startNodes(t *testing.T) {
	done := make(chan struct{})

	_ = startAlice(t, done)
	_ = startBob(t, done)
	_ = startCharlie(t, done)

	t.Cleanup(func() {
		close(done)
	})
}

func TestStartAlice(t *testing.T) {
	done := make(chan struct{})
	_ = startAlice(t, done)
	close(done)
}

func TestStartBob(t *testing.T) {
	done := make(chan struct{})
	_ = startBob(t, done)
	close(done)
}

func TestStartCharlie(t *testing.T) {
	done := make(chan struct{})
	_ = startCharlie(t, done)
	close(done)
}

func TestAlice_Discover(t *testing.T) {
	startNodes(t)
	c := client.NewClient(defaultAliceDaemonEndpoint)
	providers, err := c.Discover(common.ProvidesXMR, 15)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.GreaterOrEqual(t, len(providers[0]), 2)
}

func TestBob_Discover(t *testing.T) {
	startNodes(t)
	c := client.NewClient(defaultBobDaemonEndpoint)
	providers, err := c.Discover(common.ProvidesETH, 15)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.GreaterOrEqual(t, len(providers[0]), 2)
}

func TestAlice_Query(t *testing.T) {
	startNodes(t)
	c := client.NewClient(defaultAliceDaemonEndpoint)

	providers, err := c.Discover(common.ProvidesXMR, 3)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.GreaterOrEqual(t, len(providers[0]), 2)

	resp, err := c.Query(providers[0][0])
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Provides))
	require.Equal(t, common.ProvidesXMR, resp.Provides[0])
	require.Equal(t, 1, len(resp.MaximumAmount))
	require.Equal(t, bobProvideAmount, resp.MaximumAmount[0])
}

func TestBob_Query(t *testing.T) {
	startNodes(t)
	c := client.NewClient(defaultBobDaemonEndpoint)

	providers, err := c.Discover(common.ProvidesETH, 3)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.GreaterOrEqual(t, len(providers[0]), 2)

	resp, err := c.Query(providers[0][0])
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Provides))
	require.Equal(t, common.ProvidesETH, resp.Provides[0])
	require.Equal(t, 1, len(resp.MaximumAmount))
	require.Equal(t, aliceProvideAmount, resp.MaximumAmount[0])
}

func TestAlice_Initiate(t *testing.T) {
	startNodes(t)
	c := client.NewClient(defaultAliceDaemonEndpoint)

	providers, err := c.Discover(common.ProvidesXMR, 3)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.GreaterOrEqual(t, len(providers[0]), 2)

	ok, err := c.Initiate(providers[0][0], common.ProvidesETH, 3, 4)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestBob_Initiate(t *testing.T) {
	startNodes(t)
	c := client.NewClient(defaultBobDaemonEndpoint)

	providers, err := c.Discover(common.ProvidesETH, 3)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.GreaterOrEqual(t, len(providers[0]), 2)

	ok, err := c.Initiate(providers[0][0], common.ProvidesXMR, 3, 1)
	require.NoError(t, err)
	require.True(t, ok)
}
