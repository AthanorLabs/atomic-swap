package tests

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MarinX/monerorpc"
	"github.com/MarinX/monerorpc/daemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/rpcclient"
	"github.com/athanorlabs/atomic-swap/rpcclient/wsclient"
)

const (
	testsEnv          = "TESTS"
	integrationMode   = "integration"
	generateBlocksEnv = "GENERATEBLOCKS"
	falseStr          = "false"

	defaultXMRTakerSwapdEndpoint   = "http://localhost:5001"
	defaultXMRTakerSwapdWSEndpoint = "ws://localhost:5001/ws"
	defaultXMRMakerSwapdEndpoint   = "http://localhost:5002"
	defaultXMRMakerSwapdWSEndpoint = "ws://localhost:5002/ws"
	defaultCharlieSwapdWSEndpoint  = "ws://localhost:5003/ws"

	defaultDiscoverTimeout = 2 // 2 seconds

	xmrmakerProvideAmount = float64(1.0)
	exchangeRate          = float64(0.05)
)

type IntegrationTestSuite struct {
	suite.Suite
}

func TestRunIntegrationTests(t *testing.T) {
	if testing.Short() || os.Getenv(testsEnv) != integrationMode {
		t.Skip()
	}
	monero.BackgroundMineBlocks(t)
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupTest() {
	// Ensure minimum XMR Maker balance before each test is run
	if os.Getenv(generateBlocksEnv) != falseStr {
		// We need slightly more than xmrmakerProvideAmount for transaction fees
		mineMinXMRMakerBalance(s.T(), common.MoneroToPiconero(xmrmakerProvideAmount*2))
	}
}

// mineMinXMRMakerBalance is similar to monero.MineMinXMRBalance(...), but this version
// uses the swapd RPC Balances method to get the wallet address and balance from a
// running swapd instance instead of interacting with a wallet.
func mineMinXMRMakerBalance(t *testing.T, minBalance common.MoneroAmount) {
	daemonCli := monerorpc.New(monero.MonerodRegtestEndpoint, nil).Daemon
	for {
		balances, err := rpcclient.NewClient(defaultXMRMakerSwapdEndpoint).Balances()
		require.NoError(t, err)
		if balances.PiconeroUnlockedBalance >= uint64(minBalance) {
			break
		}
		_, err = daemonCli.GenerateBlocks(&daemon.GenerateBlocksRequest{
			AmountOfBlocks: 32,
			WalletAddress:  balances.MoneroAddress,
		})
		if err != nil && err.Error() == "Block not accepted" {
			continue
		}
		require.NoError(t, err)
	}
}

func (s *IntegrationTestSuite) TestXMRTaker_Discover() {
	bc := rpcclient.NewClient(defaultXMRMakerSwapdEndpoint)
	_, err := bc.MakeOffer(xmrmakerProvideAmount, xmrmakerProvideAmount, exchangeRate, types.EthAssetETH)
	require.NoError(s.T(), err)
	defer func() {
		err = bc.ClearOffers(nil)
		require.NoError(s.T(), err)
	}()

	ac := rpcclient.NewClient(defaultXMRTakerSwapdEndpoint)
	providers, err := ac.Discover(types.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(providers))
	require.GreaterOrEqual(s.T(), len(providers[0]), 2)
}

func (s *IntegrationTestSuite) TestXMRMaker_Discover() {
	c := rpcclient.NewClient(defaultXMRMakerSwapdEndpoint)
	providers, err := c.Discover(types.ProvidesETH, defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 0, len(providers))
}

func (s *IntegrationTestSuite) TestXMRTaker_Query() {
	bc := rpcclient.NewClient(defaultXMRMakerSwapdEndpoint)
	offerID, err := bc.MakeOffer(xmrmakerProvideAmount, xmrmakerProvideAmount, exchangeRate, types.EthAssetETH)
	require.NoError(s.T(), err)
	defer func() {
		err = bc.ClearOffers(nil)
		require.NoError(s.T(), err)
	}()

	c := rpcclient.NewClient(defaultXMRTakerSwapdEndpoint)

	providers, err := c.Discover(types.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(providers))
	require.GreaterOrEqual(s.T(), len(providers[0]), 2)

	resp, err := c.Query(providers[0][0])
	require.NoError(s.T(), err)
	require.GreaterOrEqual(s.T(), len(resp.Offers), 1)
	var respOffer *types.Offer
	for _, offer := range resp.Offers {
		if offerID == offer.GetID().String() {
			respOffer = offer
		}
	}
	require.NotNil(s.T(), respOffer)
	require.Equal(s.T(), xmrmakerProvideAmount, respOffer.MinimumAmount)
	require.Equal(s.T(), xmrmakerProvideAmount, respOffer.MaximumAmount)
	require.Equal(s.T(), exchangeRate, float64(respOffer.ExchangeRate))
}

func (s *IntegrationTestSuite) TestSuccess_OneSwap() {
	const testTimeout = time.Second * 75

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bwsc, err := wsclient.NewWsClient(ctx, defaultXMRMakerSwapdWSEndpoint)
	require.NoError(s.T(), err)

	offerID, statusCh, err := bwsc.MakeOfferAndSubscribe(0.1, xmrmakerProvideAmount,
		types.ExchangeRate(exchangeRate), types.EthAssetETH)
	require.NoError(s.T(), err)

	bc := rpcclient.NewClient(defaultXMRMakerSwapdEndpoint)
	offersBefore, err := bc.GetOffers()
	require.NoError(s.T(), err)
	defer func() {
		err = bc.ClearOffers(nil)
		require.NoError(s.T(), err)
	}()

	errCh := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for {
			select {
			case status := <-statusCh:
				s.T().Log("> XMRMaker got status:", status)
				if status.IsOngoing() {
					continue
				}

				if status != types.CompletedSuccess {
					errCh <- fmt.Errorf("swap did not complete successfully: got %s", status)
				}
				return
			case <-time.After(testTimeout):
				errCh <- errors.New("make offer subscription timed out")
				return
			}
		}
	}()

	ac := rpcclient.NewClient(defaultXMRTakerSwapdEndpoint)
	awsc, err := wsclient.NewWsClient(ctx, defaultXMRTakerSwapdWSEndpoint)
	require.NoError(s.T(), err)

	// TODO: implement discovery over websockets (#97)
	providers, err := ac.Discover(types.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(providers))
	require.GreaterOrEqual(s.T(), len(providers[0]), 2)

	takerStatusCh, err := awsc.TakeOfferAndSubscribe(providers[0][0], offerID, 0.05)
	require.NoError(s.T(), err)

	go func() {
		defer wg.Done()
		for status := range takerStatusCh {
			s.T().Log("> XMRTaker got status:", status)
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
		require.NoError(s.T(), err)
	default:
	}

	offersAfter, err := bc.GetOffers()
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(offersBefore)-len(offersAfter))
}

func (s *IntegrationTestSuite) TestRefund_XMRTakerCancels() {
	const (
		testTimeout = time.Second * 60
		swapTimeout = 30
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bwsc, err := wsclient.NewWsClient(ctx, defaultXMRMakerSwapdWSEndpoint)
	require.NoError(s.T(), err)

	offerID, statusCh, err := bwsc.MakeOfferAndSubscribe(0.1, xmrmakerProvideAmount,
		types.ExchangeRate(exchangeRate), types.EthAssetETH)
	require.NoError(s.T(), err)

	bc := rpcclient.NewClient(defaultXMRMakerSwapdEndpoint)
	offersBefore, err := bc.GetOffers()
	require.NoError(s.T(), err)
	defer func() {
		err = bc.ClearOffers(nil)
		require.NoError(s.T(), err)
	}()

	errCh := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for {
			select {
			case status := <-statusCh:
				s.T().Log("> XMRMaker got status:", status)
				if status.IsOngoing() {
					continue
				}
				switch status {
				case types.CompletedRefund:
					// Do nothing, desired outcome
				case types.CompletedSuccess:
					s.T().Log("XMRMaker completed swap before XMRTaker's cancel took affect")
				default:
					errCh <- fmt.Errorf("swap did not succeed or refund for XMRMaker: status=%s", status)
				}
				return
			case <-time.After(testTimeout):
				errCh <- errors.New("make offer subscription timed out")
				return
			}
		}
	}()

	ac := rpcclient.NewClient(defaultXMRTakerSwapdEndpoint)
	awsc, err := wsclient.NewWsClient(ctx, defaultXMRTakerSwapdWSEndpoint)
	require.NoError(s.T(), err)

	err = ac.SetSwapTimeout(swapTimeout)
	require.NoError(s.T(), err)

	providers, err := ac.Discover(types.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(providers))
	require.GreaterOrEqual(s.T(), len(providers[0]), 2)

	takerStatusCh, err := awsc.TakeOfferAndSubscribe(providers[0][0], offerID, 0.05)
	require.NoError(s.T(), err)

	go func() {
		defer wg.Done()
		for status := range takerStatusCh {
			s.T().Log("> XMRTaker got status:", status)
			if status != types.ETHLocked {
				continue
			}

			s.T().Log("> XMRTaker cancelling swap!")
			exitStatus, err := ac.Cancel(offerID) //nolint:govet
			if err != nil {
				s.T().Log("XMRTaker got error", err)
				if !strings.Contains(err.Error(), "revert it's the counterparty's turn, unable to refund") {
					errCh <- err
				}
				return
			}

			switch exitStatus {
			case types.CompletedRefund:
				// the desired outcome, do nothing
			case types.CompletedSuccess:
				s.T().Log("XMRTaker's cancel was beaten out by XMRMaker completing the swap")
			default:
				errCh <- fmt.Errorf("did not refund successfully for XMRTaker: exit status was %s", exitStatus)
			}

			return
		}
	}()

	wg.Wait()

	select {
	case err = <-errCh:
		require.NoError(s.T(), err)
	default:
	}

	offersAfter, err := bc.GetOffers()
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(offersBefore), len(offersAfter))
}

// TestRefund_XMRMakerCancels_untilAfterT1 tests the case where XMRTaker and XMRMaker
// both lock their funds, but XMRMaker goes offline
// until time t1 in the swap contract passes. This triggers XMRTaker to refund, which XMRMaker will then
// "come online" to see, and he will then refund also.
func (s *IntegrationTestSuite) TestRefund_XMRMakerCancels_untilAfterT1() {
	// Skipping test as it can't guarantee that the refund will happen before the swap completes
	// successfully:  // https://github.com/athanorlabs/atomic-swap/issues/144
	s.T().Skip()
	testRefundXMRMakerCancels(s.T(), 7, types.CompletedRefund)
	time.Sleep(time.Second * 5)
}

// TestRefund_XMRMakerCancels_afterIsReady tests the case where XMRTaker and XMRMaker both lock their
// funds, but XMRMaker goes offline until past isReady==true and t0, but comes online before t1. When
// XMRMaker comes back online, he should claim the ETH, causing XMRTaker to also claim the XMR.
func TestRefund_XMRMakerCancels_afterIsReady(t *testing.T) {
	// Skipping test as it can't guarantee that the refund will happen before the swap completes
	// successfully:  // https://github.com/athanorlabs/atomic-swap/issues/144
	t.Skip()
	testRefundXMRMakerCancels(t, 30, types.CompletedSuccess)
}

func testRefundXMRMakerCancels(t *testing.T, swapTimeout uint64, expectedExitStatus types.Status) { //nolint:unused
	const testTimeout = time.Second * 60

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bc := rpcclient.NewClient(defaultXMRMakerSwapdEndpoint)
	bwsc, err := wsclient.NewWsClient(ctx, defaultXMRMakerSwapdWSEndpoint)
	require.NoError(t, err)
	defer func() {
		err = bc.ClearOffers(nil)
		require.NoError(t, err)
	}()

	offerID, statusCh, err := bwsc.MakeOfferAndSubscribe(0.1, xmrmakerProvideAmount,
		types.ExchangeRate(exchangeRate), types.EthAssetETH)
	require.NoError(t, err)

	offersBefore, err := bc.GetOffers()
	require.NoError(t, err)

	errCh := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for {
			select {
			case status := <-statusCh:
				t.Log("> XMRMaker got status:", status)
				if status != types.XMRLocked {
					continue
				}

				t.Log("> XMRMaker cancelled swap!")
				exitStatus, err := bc.Cancel(offerID) //nolint:govet
				if err != nil {
					errCh <- err
					return
				}

				if exitStatus != expectedExitStatus {
					errCh <- fmt.Errorf("did not get expected exit status for XMRMaker: got %s, expected %s", exitStatus, expectedExitStatus) //nolint:lll
					return
				}

				t.Log("> XMRMaker refunded successfully")
				return
			case <-time.After(testTimeout):
				errCh <- errors.New("make offer subscription timed out")
			}
		}
	}()

	ac := rpcclient.NewClient(defaultXMRTakerSwapdEndpoint)
	awsc, err := wsclient.NewWsClient(ctx, defaultXMRTakerSwapdWSEndpoint)
	require.NoError(t, err)

	err = ac.SetSwapTimeout(swapTimeout)
	require.NoError(t, err)

	providers, err := ac.Discover(types.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.GreaterOrEqual(t, len(providers[0]), 2)
	takerStatusCh, err := awsc.TakeOfferAndSubscribe(providers[0][0], offerID, 0.05)
	require.NoError(t, err)

	go func() {
		defer wg.Done()

		for status := range takerStatusCh {
			t.Log("> XMRTaker got status:", status)
			if status.IsOngoing() {
				continue
			}

			if status != expectedExitStatus {
				errCh <- fmt.Errorf("did not get expected exit status for XMRTaker: got %s, expected %s", status, expectedExitStatus) //nolint:lll
				return
			}

			t.Log("> XMRTaker refunded successfully")
			return
		}
	}()

	wg.Wait()

	select {
	case err = <-errCh:
		require.NoError(t, err)
	default:
	}

	offersAfter, err := bc.GetOffers()
	require.NoError(t, err)
	if expectedExitStatus != types.CompletedSuccess {
		require.Equal(t, len(offersBefore), len(offersAfter))
	} else {
		require.Equal(t, 1, len(offersBefore)-len(offersAfter))
	}
}

// TestAbort_XMRTakerCancels tests the case where XMRTaker cancels the swap before any funds are locked.
// Both parties should abort the swap successfully.
func (s *IntegrationTestSuite) TestAbort_XMRTakerCancels() {
	const testTimeout = time.Second * 60

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bwsc, err := wsclient.NewWsClient(ctx, defaultXMRMakerSwapdWSEndpoint)
	require.NoError(s.T(), err)

	offerID, statusCh, err := bwsc.MakeOfferAndSubscribe(0.1, xmrmakerProvideAmount,
		types.ExchangeRate(exchangeRate), types.EthAssetETH)
	require.NoError(s.T(), err)

	bc := rpcclient.NewClient(defaultXMRMakerSwapdEndpoint)
	offersBefore, err := bc.GetOffers()
	require.NoError(s.T(), err)
	defer func() {
		err = bc.ClearOffers(nil)
		require.NoError(s.T(), err)
	}()

	errCh := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for {
			select {
			case status := <-statusCh:
				s.T().Log("> XMRMaker got status:", status)
				if status.IsOngoing() {
					continue
				}

				if status != types.CompletedAbort {
					errCh <- fmt.Errorf("swap did not exit successfully for XMRMaker: got %s", status)
				}

				return
			case <-time.After(testTimeout):
				errCh <- errors.New("make offer subscription timed out")
			}
		}
	}()

	ac := rpcclient.NewClient(defaultXMRTakerSwapdEndpoint)
	awsc, err := wsclient.NewWsClient(ctx, defaultXMRTakerSwapdWSEndpoint)
	require.NoError(s.T(), err)

	providers, err := ac.Discover(types.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(providers))
	require.GreaterOrEqual(s.T(), len(providers[0]), 2)

	takerStatusCh, err := awsc.TakeOfferAndSubscribe(providers[0][0], offerID, 0.05)
	require.NoError(s.T(), err)

	go func() {
		defer wg.Done()
		for status := range takerStatusCh {
			s.T().Log("> XMRTaker got status:", status)
			if status != types.ExpectingKeys {
				continue
			}

			s.T().Log("> XMRTaker cancelled swap!")
			exitStatus, err := ac.Cancel(offerID) //nolint:govet
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
		require.NoError(s.T(), err)
	default:
	}

	offersAfter, err := bc.GetOffers()
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(offersBefore), len(offersAfter))
}

// This test simulates the case where neither XMRTaker and XMRMaker have
// locked funds yet, and XMRMaker cancels the swap.
// The swap should abort on XMRMaker's side, but might abort *or* refund on XMRTaker's side, in case she ended up
// locking ETH before she was notified that XMRMaker disconnected.
func (s *IntegrationTestSuite) TestAbort_XMRMakerCancels() {
	const testTimeout = time.Second * 60

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bcli := rpcclient.NewClient(defaultXMRMakerSwapdEndpoint)
	bwsc, err := wsclient.NewWsClient(ctx, defaultXMRMakerSwapdWSEndpoint)
	require.NoError(s.T(), err)

	offerID, statusCh, err := bwsc.MakeOfferAndSubscribe(0.1, xmrmakerProvideAmount,
		types.ExchangeRate(exchangeRate), types.EthAssetETH)
	require.NoError(s.T(), err)

	bc := rpcclient.NewClient(defaultXMRMakerSwapdEndpoint)
	offersBefore, err := bc.GetOffers()
	require.NoError(s.T(), err)
	defer func() {
		err = bc.ClearOffers(nil)
		require.NoError(s.T(), err)
	}()

	errCh := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for {
			select {
			case status := <-statusCh:
				s.T().Log("> XMRMaker got status:", status)
				if status != types.KeysExchanged {
					continue
				}

				s.T().Log("> XMRMaker cancelled swap!")
				exitStatus, err := bcli.Cancel(offerID) //nolint:govet
				if err != nil {
					errCh <- err
					return
				}

				if exitStatus != types.CompletedAbort {
					errCh <- fmt.Errorf("did not abort successfully: exit status was %s", exitStatus)
					return
				}

				s.T().Log("> XMRMaker exited successfully")
				return
			case <-time.After(testTimeout):
				errCh <- errors.New("make offer subscription timed out")
			}
		}
	}()

	c := rpcclient.NewClient(defaultXMRTakerSwapdEndpoint)
	wsc, err := wsclient.NewWsClient(ctx, defaultXMRTakerSwapdWSEndpoint)
	require.NoError(s.T(), err)

	providers, err := c.Discover(types.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(providers))
	require.GreaterOrEqual(s.T(), len(providers[0]), 2)

	takerStatusCh, err := wsc.TakeOfferAndSubscribe(providers[0][0], offerID, 0.05)
	require.NoError(s.T(), err)

	go func() {
		defer wg.Done()

		for status := range takerStatusCh {
			s.T().Log("> XMRTaker got status:", status)
			if status.IsOngoing() {
				continue
			}

			if status != types.CompletedAbort && status != types.CompletedRefund {
				errCh <- fmt.Errorf("swap did not exit successfully: got %s", status)
				return
			}

			s.T().Log("> XMRTaker exited successfully")
			return
		}
	}()

	wg.Wait()

	select {
	case err = <-errCh:
		require.NoError(s.T(), err)
	default:
	}

	offersAfter, err := bc.GetOffers()
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(offersBefore), len(offersAfter))
}

// TestError_ShouldOnlyTakeOfferOnce tests the case where two takers try to take the same offer concurrently.
// Only one should succeed, the other should return an error or Abort status.
func (s *IntegrationTestSuite) TestError_ShouldOnlyTakeOfferOnce() {
	const testTimeout = time.Second * 60
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bc := rpcclient.NewClient(defaultXMRMakerSwapdEndpoint)
	offerID, err := bc.MakeOffer(xmrmakerProvideAmount, xmrmakerProvideAmount, exchangeRate, types.EthAssetETH)
	require.NoError(s.T(), err)
	defer func() {
		err = bc.ClearOffers(nil)
		require.NoError(s.T(), err)
	}()

	ac := rpcclient.NewClient(defaultXMRTakerSwapdEndpoint)

	providers, err := ac.Discover(types.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(providers))
	require.GreaterOrEqual(s.T(), len(providers[0]), 2)

	errCh := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		wsc, err := wsclient.NewWsClient(ctx, defaultXMRTakerSwapdWSEndpoint)
		require.NoError(s.T(), err)

		takerStatusCh, err := wsc.TakeOfferAndSubscribe(providers[0][0], offerID, 0.05)
		if err != nil {
			errCh <- err
			return
		}

		for status := range takerStatusCh {
			s.T().Log("> XMRTaker[0] got status:", status)
			if status.IsOngoing() {
				continue
			}

			if status != types.CompletedSuccess {
				errCh <- fmt.Errorf("0th swap did not exit successfully: got %s", status)
				cancel()
				return
			}

			s.T().Log("> XMRTaker[0] exited successfully")
			return
		}
	}()

	go func() {
		defer wg.Done()
		wsc, err := wsclient.NewWsClient(ctx, defaultCharlieSwapdWSEndpoint)
		require.NoError(s.T(), err)

		takerStatusCh, err := wsc.TakeOfferAndSubscribe(providers[0][0], offerID, 0.05)
		if err != nil {
			errCh <- err
			return
		}

		for status := range takerStatusCh {
			s.T().Log("> XMRTaker[1] got status:", status)
			if status.IsOngoing() {
				continue
			}

			if status != types.CompletedSuccess {
				errCh <- fmt.Errorf("1st swap did not exit successfully: got %s", status)
				return
			}

			s.T().Log("> XMRTaker[1] exited successfully")
			return
		}
	}()

	wg.Wait()

	select {
	case err := <-errCh:
		require.NotNil(s.T(), err)
		s.T().Log("got expected error:", err)
	case <-time.After(testTimeout):
		s.T().Fatalf("did not get error from XMRTaker or Charlie")
	}

	select {
	case err := <-errCh:
		s.T().Fatalf("should only have one error! also got %s", err)
	default:
	}
}

func (s *IntegrationTestSuite) TestSuccess_ConcurrentSwaps() {
	const testTimeout = time.Minute * 6
	const numConcurrentSwaps = 10

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type makerTest struct {
		offerID  string
		statusCh <-chan types.Status
		errCh    chan error
	}

	makerTests := make([]*makerTest, numConcurrentSwaps)

	for i := 0; i < numConcurrentSwaps; i++ {
		bwsc, err := wsclient.NewWsClient(ctx, defaultXMRMakerSwapdWSEndpoint)
		require.NoError(s.T(), err)

		offerID, statusCh, err := bwsc.MakeOfferAndSubscribe(0.1, xmrmakerProvideAmount,
			types.ExchangeRate(exchangeRate), types.EthAssetETH)
		require.NoError(s.T(), err)

		s.T().Log("maker made offer ", offerID)

		makerTests[i] = &makerTest{
			offerID:  offerID,
			statusCh: statusCh,
			errCh:    make(chan error, 2),
		}
	}

	bc := rpcclient.NewClient(defaultXMRMakerSwapdEndpoint)
	offersBefore, err := bc.GetOffers()
	require.NoError(s.T(), err)
	defer func() {
		err = bc.ClearOffers(nil)
		require.NoError(s.T(), err)
	}()

	var wg sync.WaitGroup
	wg.Add(2 * numConcurrentSwaps)

	for _, tc := range makerTests {
		go func(tc *makerTest) {
			defer wg.Done()

			for {
				select {
				case status := <-tc.statusCh:
					s.T().Log("> XMRMaker got status:", status)
					if status.IsOngoing() {
						continue
					}

					if status != types.CompletedSuccess {
						tc.errCh <- fmt.Errorf("swap did not complete successfully: got %s", status)
					}

					return
				case <-time.After(testTimeout):
					tc.errCh <- errors.New("make offer subscription timed out")
				}
			}
		}(tc)
	}

	type takerTest struct {
		statusCh <-chan types.Status
		errCh    chan error
	}

	takerTests := make([]*takerTest, numConcurrentSwaps)

	for i := 0; i < numConcurrentSwaps; i++ {
		ac := rpcclient.NewClient(defaultXMRTakerSwapdEndpoint)
		awsc, err := wsclient.NewWsClient(ctx, defaultXMRTakerSwapdWSEndpoint) //nolint:govet
		require.NoError(s.T(), err)

		// TODO: implement discovery over websockets (#97)
		providers, err := ac.Discover(types.ProvidesXMR, defaultDiscoverTimeout)
		require.NoError(s.T(), err)
		require.Equal(s.T(), 1, len(providers))
		require.GreaterOrEqual(s.T(), len(providers[0]), 2)

		offerID := makerTests[i].offerID
		takerStatusCh, err := awsc.TakeOfferAndSubscribe(providers[0][0], offerID, 0.05)
		require.NoError(s.T(), err)

		s.T().Log("taker took offer ", offerID)

		takerTests[i] = &takerTest{
			statusCh: takerStatusCh,
			errCh:    make(chan error, 2),
		}
	}

	for _, tc := range takerTests {
		go func(tc *takerTest) {
			defer wg.Done()
			for status := range tc.statusCh {
				s.T().Log("> XMRTaker got status:", status)
				if status.IsOngoing() {
					continue
				}

				if status != types.CompletedSuccess {
					tc.errCh <- fmt.Errorf("swap did not complete successfully: got %s", status)
				}

				return
			}
		}(tc)
	}

	wg.Wait()

	for _, tc := range makerTests {
		select {
		case err = <-tc.errCh:
			assert.NoError(s.T(), err)
		default:
		}
	}

	for _, tc := range takerTests {
		select {
		case err = <-tc.errCh:
			assert.NoError(s.T(), err)
		default:
		}
	}

	offersAfter, err := bc.GetOffers()
	require.NoError(s.T(), err)
	require.Equal(s.T(), numConcurrentSwaps, len(offersBefore)-len(offersAfter))
}
