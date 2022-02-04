package tests

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/noot/atomic-swap/cmd/client/client"
	"github.com/noot/atomic-swap/common/types"

	"github.com/stretchr/testify/require"
)

const (
	defaultAliceTestLibp2pKey  = "alice.key"
	defaultAliceDaemonEndpoint = "http://localhost:5001"
	defaultBobDaemonEndpoint   = "http://localhost:5002"
	defaultDiscoverTimeout     = 2 // 2 seconds

	bobProvideAmount = float64(44.4)
	exchangeRate     = float64(0.05)
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(0)
	}

	cmd := exec.Command("../scripts/build.sh")
	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func startSwapDaemon(t *testing.T, done <-chan struct{}, args ...string) {
	cmd := exec.Command("../swapd", args...)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	type errOut struct {
		err error
		out string
	}

	errCh := make(chan *errOut)
	go func() {
		out, err := cmd.CombinedOutput()
		if err != nil {
			errCh <- &errOut{
				err: err,
				out: string(out),
			}
		}

		wg.Done()
	}()

	go func() {
		defer wg.Done()

		select {
		case <-done:
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			// drain errCh
			<-errCh
			return
		case err := <-errCh:
			fmt.Println("program exited early: ", err.err)
			fmt.Println("output: ", err.out)
		}
	}()

	t.Cleanup(func() {
		wg.Wait()
	})

	time.Sleep(time.Second * 2)
}

func startAlice(t *testing.T, done <-chan struct{}) []string {
	startSwapDaemon(t, done, "--dev-alice",
		"--libp2p-key", defaultAliceTestLibp2pKey,
	)
	c := client.NewClient(defaultAliceDaemonEndpoint)
	addrs, err := c.Addresses()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(addrs), 1)
	return addrs
}

func startBob(t *testing.T, done <-chan struct{}, aliceMultiaddr string) {
	startSwapDaemon(t, done, "--dev-bob",
		"--bootnodes", aliceMultiaddr,
		"--wallet-file", "test-wallet",
	)
}

// charlie doesn't provide any coin or participate in any swap.
// he is just a node running the p2p protocol.
func startCharlie(t *testing.T, done <-chan struct{}, aliceMultiaddr string) {
	startSwapDaemon(t, done,
		"--libp2p-port", "9955",
		"--rpc-port", "5003",
		"--bootnodes", aliceMultiaddr)
}

func startNodes(t *testing.T) {
	done := make(chan struct{})

	addrs := startAlice(t, done)
	startBob(t, done, addrs[0])
	startCharlie(t, done, addrs[0])

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
	addrs := startAlice(t, done)
	startBob(t, done, addrs[0])
	close(done)
}

func TestStartCharlie(t *testing.T) {
	done := make(chan struct{})
	addrs := startAlice(t, done)
	startCharlie(t, done, addrs[0])
	close(done)
}

func TestAlice_Discover(t *testing.T) {
	startNodes(t)
	bc := client.NewClient(defaultBobDaemonEndpoint)
	_, err := bc.MakeOffer(bobProvideAmount, bobProvideAmount, exchangeRate)
	require.NoError(t, err)

	c := client.NewClient(defaultAliceDaemonEndpoint)
	providers, err := c.Discover(types.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.GreaterOrEqual(t, len(providers[0]), 2)
}

func TestBob_Discover(t *testing.T) {
	startNodes(t)
	c := client.NewClient(defaultBobDaemonEndpoint)
	providers, err := c.Discover(types.ProvidesETH, defaultDiscoverTimeout)
	require.NoError(t, err)
	require.Equal(t, 0, len(providers))
}

func TestAlice_Query(t *testing.T) {
	startNodes(t)
	bc := client.NewClient(defaultBobDaemonEndpoint)
	_, err := bc.MakeOffer(bobProvideAmount, bobProvideAmount, exchangeRate)
	require.NoError(t, err)

	c := client.NewClient(defaultAliceDaemonEndpoint)

	providers, err := c.Discover(types.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.GreaterOrEqual(t, len(providers[0]), 2)

	resp, err := c.Query(providers[0][0])
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Offers))
	require.Equal(t, bobProvideAmount, resp.Offers[0].MinimumAmount)
	require.Equal(t, bobProvideAmount, resp.Offers[0].MaximumAmount)
	require.Equal(t, exchangeRate, float64(resp.Offers[0].ExchangeRate))
}

func TestAlice_TakeOffer(t *testing.T) {
	startNodes(t)

	bc := client.NewClient(defaultBobDaemonEndpoint)
	offerID, err := bc.MakeOffer(0.1, bobProvideAmount, exchangeRate)
	require.NoError(t, err)

	c := client.NewClient(defaultAliceDaemonEndpoint)

	providers, err := c.Discover(types.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.GreaterOrEqual(t, len(providers[0]), 2)

	id, err := c.TakeOffer(providers[0][0], offerID, 0.1)
	require.NoError(t, err)
	require.Equal(t, uint64(0), id)
}
