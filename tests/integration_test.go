package tests

import (
	"context"
	//"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/noot/atomic-swap/cmd/client/client"
	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/rpcclient"
	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/monero"

	"github.com/stretchr/testify/require"
)

const (
	testsEnv          = "TESTS"
	integrationMode   = "integration"
	generateBlocksEnv = "GENERATEBLOCKS"

	defaultAliceTestLibp2pKey    = "alice.key"
	defaultAliceDaemonEndpoint   = "http://localhost:5001"
	defaultAliceDaemonWSEndpoint = "ws://localhost:8081"
	defaultBobDaemonEndpoint     = "http://localhost:5002"
	defaultBobDaemonWSEndpoint   = "ws://localhost:8082" //nolint
	defaultDiscoverTimeout       = 2                     // 2 seconds

	bobProvideAmount = float64(1.0)
	exchangeRate     = float64(0.05)
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(0)
	}

	if os.Getenv(testsEnv) != integrationMode {
		os.Exit(0)
	}

	cmd := exec.Command("../scripts/build.sh")
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("%s\n%s", out, err))
	}

	c := monero.NewClient(common.DefaultBobMoneroEndpoint)
	d := monero.NewDaemonClient(common.DefaultMoneroDaemonEndpoint)
	bobAddr, err := c.GetAddress(0)
	if err != nil {
		panic(err)
	}

	if os.Getenv(generateBlocksEnv) != "false" {
		fmt.Println("> Generating blocks for test setup...")
		_ = d.GenerateBlocks(bobAddr.Address, 512)
		err = c.Refresh()
		if err != nil {
			panic(err)
		}

		fmt.Println("> Completed generating blocks.")
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

	time.Sleep(time.Second * 5)
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
	//const testTimeout = time.Second * 5

	startNodes(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bc := client.NewClient(defaultBobDaemonEndpoint)
	offerID, err := bc.MakeOffer(0.1, 1, 0.05)
	require.NoError(t, err)

	// bwsc, err := rpcclient.NewWsClient(ctx, defaultBobDaemonWSEndpoint)
	// require.NoError(t, err)

	// offerID, takenCh, statusCh, err := bwsc.MakeOfferAndSubscribe(0.1, bobProvideAmount,
	// 	types.ExchangeRate(exchangeRate))
	// require.NoError(t, err)

	errCh := make(chan error, 2)

	// var wg sync.WaitGroup
	// wg.Add(2)

	// go func() {
	// 	defer close(errCh)

	// 	select {
	// 	case taken := <-takenCh:
	// 		require.NotNil(t, taken)
	// 		t.Log("swap ID: ", taken.ID)
	// 	case <-time.After(testTimeout):
	// 		errCh <- errors.New("make offer subscription timed out")
	// 	}

	// 	for status := range statusCh {
	// 		if !status.IsOngoing() {
	// 			continue
	// 		}

	// 		require.Equal(t, types.CompletedSuccess, status)
	// 	}

	// 	wg.Done()
	// }()

	c := client.NewClient(defaultAliceDaemonEndpoint)
	wsc, err := rpcclient.NewWsClient(ctx, defaultAliceDaemonWSEndpoint)
	require.NoError(t, err)

	// TODO: implement discovery over websockets
	providers, err := c.Discover(types.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.GreaterOrEqual(t, len(providers[0]), 2)

	id, takerStatusCh, err := wsc.TakeOfferAndSubscribe(providers[0][0], offerID, 0.05)
	require.NoError(t, err)
	require.Equal(t, uint64(0), id)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for status := range takerStatusCh {
			fmt.Println("> Got status:", status)
			if status.IsOngoing() {
				continue
			}

			if status != types.CompletedSuccess {
				errCh <- fmt.Errorf("swap did not complete successfully: got %s", status)
			}

			return
		}
	}()

	wg.Wait()
	err = <-errCh
	require.NoError(t, err)
}
