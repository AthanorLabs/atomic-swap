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
	falseStr          = "false"

	defaultAliceDaemonEndpoint     = "http://localhost:5001"
	defaultAliceDaemonWSEndpoint   = "ws://localhost:8081"
	defaultBobDaemonEndpoint       = "http://localhost:5002"
	defaultBobDaemonWSEndpoint     = "ws://localhost:8082"
	defaultCharlieDaemonWSEndpoint = "ws://localhost:8083"

	defaultDiscoverTimeout = 2 // 2 seconds

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

	if os.Getenv(generateBlocksEnv) != falseStr {
		generateBlocks(512)
	}

	os.Exit(m.Run())
}

func generateBlocks(num uint) {
	c := monero.NewClient(common.DefaultBobMoneroEndpoint)
	d := monero.NewDaemonClient(common.DefaultMoneroDaemonEndpoint)
	bobAddr, err := c.GetAddress(0)
	if err != nil {
		panic(err)
	}

	fmt.Println("> Generating blocks for test setup...")
	_ = d.GenerateBlocks(bobAddr.Address, num)
	err = c.Refresh()
	if err != nil {
		panic(err)
	}

	fmt.Println("> Completed generating blocks.")
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

func TestSuccess(t *testing.T) {
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

	select {
	case err = <-errCh:
		require.NoError(t, err)
	default:
	}

	bobSwapID := <-bobIDCh
	require.Equal(t, id, bobSwapID)

	offersAfter, err := bc.GetOffers()
	require.NoError(t, err)
	require.Equal(t, 1, len(offersBefore)-len(offersAfter))
}

func TestRefund_AliceCancels(t *testing.T) {
	if os.Getenv(generateBlocksEnv) != falseStr {
		generateBlocks(64)
	}

	const (
		testTimeout = time.Second * 5
		swapTimeout = 5 // 5s
	)

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

			if status != types.CompletedRefund {
				errCh <- fmt.Errorf("swap did not refund successfully for Bob: exit status was %s", status)
			}

			return
		}
	}()

	c := client.NewClient(defaultAliceDaemonEndpoint)
	wsc, err := rpcclient.NewWsClient(ctx, defaultAliceDaemonWSEndpoint)
	require.NoError(t, err)

	err = c.SetSwapTimeout(swapTimeout)
	require.NoError(t, err)

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
			if status != types.ETHLocked {
				continue
			}

			fmt.Println("> Alice cancelled swap!")
			exitStatus, err := c.Cancel() //nolint:govet
			if err != nil {
				t.Log("Alice got error", err)
				errCh <- err
				return
			}

			if exitStatus != types.CompletedRefund {
				errCh <- fmt.Errorf("did not refund successfully for Alice: exit status was %s", exitStatus)
			}

			return
		}
	}()

	wg.Wait()

	select {
	case err = <-errCh:
		require.NoError(t, err)
	default:
	}

	bobSwapID := <-bobIDCh
	require.Equal(t, id, bobSwapID)

	offersAfter, err := bc.GetOffers()
	require.NoError(t, err)
	require.Equal(t, len(offersBefore), len(offersAfter))
}

// TestRefund_BobCancels_untilAfterT1 tests the case where Alice and Bob both lock their funds, but Bob goes offline
// until time t1 in the swap contract passes. This triggers Alice to refund, which Bob will then
// "come online" to see, and he will then refund also.
func TestRefund_BobCancels_untilAfterT1(t *testing.T) {
	testRefundBobCancels(t, 5, types.CompletedRefund)
}

// TestRefund_BobCancels_afterIsReady tests the case where Alice and Bob both lock their funds,
// but Bob goes offline until past isReady==true and t0, but comes online before t1.
//  When Bob comes back online, he should claim the ETH, causing Alice to also claim the XMR.
func TestRefund_BobCancels_afterIsReady(t *testing.T) {
	testRefundBobCancels(t, 30, types.CompletedSuccess)
}

func testRefundBobCancels(t *testing.T, swapTimeout uint64, expectedExitStatus types.Status) {
	if os.Getenv(generateBlocksEnv) != falseStr {
		generateBlocks(64)
	}

	const testTimeout = time.Second * 5

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bcli := client.NewClient(defaultBobDaemonEndpoint)
	bwsc, err := rpcclient.NewWsClient(ctx, defaultBobDaemonWSEndpoint)
	require.NoError(t, err)

	offerID, takenCh, statusCh, err := bwsc.MakeOfferAndSubscribe(0.1, bobProvideAmount,
		types.ExchangeRate(exchangeRate))
	require.NoError(t, err)

	offersBefore, err := bcli.GetOffers()
	require.NoError(t, err)

	bobIDCh := make(chan uint64, 1)
	errCh := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
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
			if status != types.XMRLocked {
				continue
			}

			fmt.Println("> Bob cancelled swap!")
			exitStatus, err := bcli.Cancel() //nolint:govet
			if err != nil {
				errCh <- err
				return
			}

			if exitStatus != expectedExitStatus {
				errCh <- fmt.Errorf("did not get expected exit status for Bob: got %s, expected %s", exitStatus, expectedExitStatus)
				return
			}

			fmt.Println("> Bob refunded successfully")
			return
		}
	}()

	c := client.NewClient(defaultAliceDaemonEndpoint)
	wsc, err := rpcclient.NewWsClient(ctx, defaultAliceDaemonWSEndpoint)
	require.NoError(t, err)

	err = c.SetSwapTimeout(swapTimeout)
	require.NoError(t, err)

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

			if status != expectedExitStatus {
				errCh <- fmt.Errorf("did not get expected exit status for Alice: got %s, expected %s", status, expectedExitStatus)
				return
			}

			fmt.Println("> Alice refunded successfully")
			return
		}
	}()

	wg.Wait()

	select {
	case err = <-errCh:
		require.NoError(t, err)
	default:
	}

	bobSwapID := <-bobIDCh
	require.Equal(t, id, bobSwapID)

	offersAfter, err := bcli.GetOffers()
	require.NoError(t, err)
	if expectedExitStatus != types.CompletedSuccess {
		require.Equal(t, len(offersBefore), len(offersAfter))
	} else {
		require.Equal(t, 1, len(offersBefore)-len(offersAfter))
	}
}

// TestAbort_AliceCancels tests the case where Alice cancels the swap before any funds are locked.
// Both parties should abort the swap successfully.
func TestAbort_AliceCancels(t *testing.T) {
	if os.Getenv(generateBlocksEnv) != falseStr {
		generateBlocks(64)
	}

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

			if status != types.CompletedAbort {
				errCh <- fmt.Errorf("swap did not exit successfully: got %s", status)
			}

			return
		}
	}()

	c := client.NewClient(defaultAliceDaemonEndpoint)
	wsc, err := rpcclient.NewWsClient(ctx, defaultAliceDaemonWSEndpoint)
	require.NoError(t, err)

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
			if status != types.ExpectingKeys {
				continue
			}

			fmt.Println("> Alice cancelled swap!")
			exitStatus, err := c.Cancel() //nolint:govet
			if err != nil {
				errCh <- err
				return
			}

			if exitStatus != types.CompletedAbort {
				errCh <- fmt.Errorf("did not refund exit: exit status was %s", exitStatus)
			}

			return
		}
	}()

	wg.Wait()

	select {
	case err = <-errCh:
		require.NoError(t, err)
	default:
	}

	bobSwapID := <-bobIDCh
	require.Equal(t, id, bobSwapID)

	offersAfter, err := bc.GetOffers()
	require.NoError(t, err)
	require.Equal(t, len(offersBefore), len(offersAfter))
}

// This test simulates the case where neither Alice and Bob have locked funds yet, and Bob cancels the swap.
// The swap should abort on Bob's side, but might abort *or* refund on Alice's side, in case she ended up
// locking ETH before she was notified that Bob disconnected.
func TestAbort_BobCancels(t *testing.T) {
	if os.Getenv(generateBlocksEnv) != falseStr {
		generateBlocks(64)
	}

	const testTimeout = time.Second * 5

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bcli := client.NewClient(defaultBobDaemonEndpoint)
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
			if status != types.KeysExchanged {
				continue
			}

			fmt.Println("> Bob cancelled swap!")
			exitStatus, err := bcli.Cancel() //nolint:govet
			if err != nil {
				errCh <- err
				return
			}

			if exitStatus != types.CompletedAbort {
				errCh <- fmt.Errorf("did not abort successfully: exit status was %s", exitStatus)
				return
			}

			fmt.Println("> Bob exited successfully")
			return
		}
	}()

	c := client.NewClient(defaultAliceDaemonEndpoint)
	wsc, err := rpcclient.NewWsClient(ctx, defaultAliceDaemonWSEndpoint)
	require.NoError(t, err)

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

			if status != types.CompletedAbort && status != types.CompletedRefund {
				errCh <- fmt.Errorf("swap did not exit successfully: got %s", status)
				return
			}

			fmt.Println("> Alice exited successfully")
			return
		}
	}()

	wg.Wait()

	select {
	case err = <-errCh:
		require.NoError(t, err)
	default:
	}

	bobSwapID := <-bobIDCh
	require.Equal(t, id, bobSwapID)

	offersAfter, err := bc.GetOffers()
	require.NoError(t, err)
	require.Equal(t, len(offersBefore), len(offersAfter))
}

// TestError_ShouldOnlyTakeOfferOnce tests the case where two takers try to take the same offer concurrently.
// Only one should succeed, the other should return an error or Abort status.
func TestError_ShouldOnlyTakeOfferOnce(t *testing.T) {
	const testTimeout = time.Second * 5

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bc := client.NewClient(defaultBobDaemonEndpoint)
	offerID, err := bc.MakeOffer(bobProvideAmount, bobProvideAmount, exchangeRate)
	require.NoError(t, err)

	ac := client.NewClient(defaultAliceDaemonEndpoint)

	providers, err := ac.Discover(types.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.GreaterOrEqual(t, len(providers[0]), 2)

	errCh := make(chan error)

	go func() {
		wsc, err := rpcclient.NewWsClient(ctx, defaultAliceDaemonWSEndpoint)
		require.NoError(t, err)

		_, takerStatusCh, err := wsc.TakeOfferAndSubscribe(providers[0][0], offerID, 0.05)
		if err != nil {
			errCh <- err
			return
		}

		for status := range takerStatusCh {
			fmt.Println("> Alice got status:", status)
			if status.IsOngoing() {
				continue
			}

			if status != types.CompletedSuccess {
				errCh <- fmt.Errorf("swap did not exit successfully: got %s", status)
				return
			}

			fmt.Println("> Alice exited successfully")
			return
		}
	}()

	go func() {
		wsc, err := rpcclient.NewWsClient(ctx, defaultCharlieDaemonWSEndpoint)
		require.NoError(t, err)

		_, takerStatusCh, err := wsc.TakeOfferAndSubscribe(providers[0][0], offerID, 0.05)
		if err != nil {
			errCh <- err
			return
		}

		for status := range takerStatusCh {
			fmt.Println("> Alice got status:", status)
			if status.IsOngoing() {
				continue
			}

			if status != types.CompletedSuccess {
				errCh <- fmt.Errorf("swap did not exit successfully: got %s", status)
				return
			}

			fmt.Println("> Alice exited successfully")
			return
		}
	}()

	select {
	case err := <-errCh:
		require.NotNil(t, err)
		t.Log("got expected error", err)
	case <-time.After(testTimeout):
		t.Fatalf("did not get error from Alice or Charlie")
	}

	select {
	case err := <-errCh:
		t.Fatalf("should only have one error! also got %s", err)
	default:
	}
}
