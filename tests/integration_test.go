package tests

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
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

	defaultAliceDaemonEndpoint   = "http://localhost:5001"
	defaultAliceDaemonWSEndpoint = "ws://localhost:8081"
	defaultBobDaemonEndpoint     = "http://localhost:5002"
	defaultBobDaemonWSEndpoint   = "ws://localhost:8082"
	defaultDiscoverTimeout       = 2 // 2 seconds

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

func TestAlice_Discover(t *testing.T) {
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
	c := client.NewClient(defaultBobDaemonEndpoint)
	providers, err := c.Discover(types.ProvidesETH, defaultDiscoverTimeout)
	require.NoError(t, err)
	require.Equal(t, 0, len(providers))
}

func TestAlice_Query(t *testing.T) {
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
	require.GreaterOrEqual(t, len(resp.Offers), 1)
	require.Equal(t, bobProvideAmount, resp.Offers[0].MinimumAmount)
	require.Equal(t, bobProvideAmount, resp.Offers[0].MaximumAmount)
	require.Equal(t, exchangeRate, float64(resp.Offers[0].ExchangeRate))
}

func TestTakeOffer_HappyPath(t *testing.T) {
	const testTimeout = time.Second * 5

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bwsc, err := rpcclient.NewWsClient(ctx, defaultBobDaemonWSEndpoint)
	require.NoError(t, err)

	offerID, takenCh, statusCh, err := bwsc.MakeOfferAndSubscribe(0.1, bobProvideAmount,
		types.ExchangeRate(exchangeRate))
	require.NoError(t, err)

	bc := client.NewClient(defaultBobDaemonEndpoint)
	offersBefore, err := bc.GetOffers()
	require.NoError(t, err)

	bobIDCh := make(chan uint64, 1)
	errCh := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer close(errCh)
		defer wg.Done()

		select {
		case taken := <-takenCh:
			require.NotNil(t, taken)
			t.Log("swap ID:", taken.ID)
			bobIDCh <- taken.ID
		case <-time.After(testTimeout):
			errCh <- errors.New("make offer subscription timed out")
		}

		for status := range statusCh {
			fmt.Println("> Bob got status:", status)
			if status.IsOngoing() {
				continue
			}

			if status != types.CompletedSuccess {
				errCh <- fmt.Errorf("swap did not complete successfully: got %s", status)
			}

			return
		}
	}()

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

	go func() {
		defer wg.Done()
		for status := range takerStatusCh {
			fmt.Println("> Alice got status:", status)
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
	bobSwapID := <-bobIDCh
	require.Equal(t, id, bobSwapID)

	offersAfter, err := bc.GetOffers()
	require.NoError(t, err)
	require.Equal(t, 1, len(offersBefore)-len(offersAfter))
}
