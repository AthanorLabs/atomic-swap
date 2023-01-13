// Package tests provides integration tests, which exercise fully built swapd instances
// pre-launched by a script. The non *_test.go files are test helper methods for both
// unit and integration tests.
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
	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/athanorlabs/atomic-swap/coins"
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

	defaultXMRTakerSwapdEndpoint   = "http://localhost:5000"
	defaultXMRTakerSwapdWSEndpoint = "ws://localhost:5000/ws"
	defaultXMRMakerSwapdEndpoint   = "http://localhost:5001"
	defaultXMRMakerSwapdWSEndpoint = "ws://localhost:5001/ws"
	defaultCharlieSwapdWSEndpoint  = "ws://localhost:5002/ws"

	defaultDiscoverTimeout = 2 // 2 seconds

	defaultSwapTimeout = 90 // number of seconds that we reset the taker's swap timeout to between tests
)

var (
	xmrmakerProvideAmount = apd.New(1, 0)
	exchangeRate          = coins.StrToExchangeRate("0.05")
)

type IntegrationTestSuite struct {
	suite.Suite
}

func TestRunIntegrationTests(t *testing.T) {
	if testing.Short() || os.Getenv(testsEnv) != integrationMode {
		t.Skip()
	}
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupTest() {
	// Ensure minimum XMR Maker balance before each test is run
	if os.Getenv(generateBlocksEnv) != falseStr {
		// We need slightly more than xmrmakerProvideAmount for transaction fees
		minBal := new(apd.Decimal)
		_, err := coins.DecimalCtx().Mul(minBal, xmrmakerProvideAmount, apd.New(2, 0))
		require.NoError(s.T(), err)
		mineMinXMRMakerBalance(s.T(), coins.MoneroToPiconero(minBal))
	}

	// Reset XMR Maker and Taker between tests, so tests starts in a known state
	ac := rpcclient.NewClient(context.Background(), defaultXMRTakerSwapdEndpoint)
	err := ac.SetSwapTimeout(defaultSwapTimeout)
	require.NoError(s.T(), err)
	bc := rpcclient.NewClient(context.Background(), defaultXMRMakerSwapdEndpoint)
	err = bc.ClearOffers(nil)
	require.NoError(s.T(), err)
}

// mineMinXMRMakerBalance is similar to monero.MineMinXMRBalance(...), but this version
// uses the swapd RPC Balances method to get the wallet address and balance from a
// running swapd instance instead of interacting with a wallet.
func mineMinXMRMakerBalance(t *testing.T, minBalance *coins.PiconeroAmount) {
	daemonCli := monerorpc.New(monero.MonerodRegtestEndpoint, nil).Daemon
	ctx := context.Background()
	for {
		balances, err := rpcclient.NewClient(ctx, defaultXMRMakerSwapdEndpoint).Balances()
		require.NoError(t, err)
		if balances.PiconeroUnlockedBalance.Cmp(minBalance) >= 0 {
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

func (s *IntegrationTestSuite) newSwapdWSClient(ctx context.Context, endpoint string) wsclient.WsClient {
	wsc, err := wsclient.NewWsClient(ctx, endpoint)
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		wsc.Close()
	})
	return wsc
}

func (s *IntegrationTestSuite) TestXMRTaker_Discover() {
	ctx := context.Background()
	bc := rpcclient.NewClient(ctx, defaultXMRMakerSwapdEndpoint)
	_, err := bc.MakeOffer(xmrmakerProvideAmount, xmrmakerProvideAmount, exchangeRate, types.EthAssetETH, "", nil)
	require.NoError(s.T(), err)

	ac := rpcclient.NewClient(ctx, defaultXMRTakerSwapdEndpoint)
	peerIDs, err := ac.Discover(coins.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(peerIDs))
}

func (s *IntegrationTestSuite) TestXMRMaker_Discover() {
	ctx := context.Background()
	c := rpcclient.NewClient(ctx, defaultXMRMakerSwapdEndpoint)
	peerIDs, err := c.Discover(coins.ProvidesETH, defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 0, len(peerIDs))
}

func (s *IntegrationTestSuite) TestXMRTaker_Query() {
	s.testXMRTakerQuery(types.EthAssetETH)
}

func (s *IntegrationTestSuite) testXMRTakerQuery(asset types.EthAsset) {
	ctx := context.Background()
	bc := rpcclient.NewClient(ctx, defaultXMRMakerSwapdEndpoint)
	offerResp, err := bc.MakeOffer(xmrmakerProvideAmount, xmrmakerProvideAmount, exchangeRate, asset, "", nil)
	require.NoError(s.T(), err)

	c := rpcclient.NewClient(ctx, defaultXMRTakerSwapdEndpoint)

	peerIDs, err := c.Discover(coins.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(peerIDs))

	resp, err := c.Query(peerIDs[0])
	require.NoError(s.T(), err)
	require.GreaterOrEqual(s.T(), len(resp.Offers), 1)
	var respOffer *types.Offer
	for _, offer := range resp.Offers {
		if offer.ID == offerResp.OfferID {
			respOffer = offer
		}
	}

	require.NotNil(s.T(), respOffer)
	require.Equal(s.T(), xmrmakerProvideAmount.String(), respOffer.MinAmount.String())
	require.Equal(s.T(), xmrmakerProvideAmount.String(), respOffer.MaxAmount.String())
	require.Equal(s.T(), exchangeRate.String(), respOffer.ExchangeRate.String())
	require.Equal(s.T(), asset, respOffer.EthAsset)
}

func (s *IntegrationTestSuite) TestSuccess_OneSwap() {
	s.testSuccessOneSwap(types.EthAssetETH, "", nil)
}

func (s *IntegrationTestSuite) testSuccessOneSwap(
	asset types.EthAsset,
	relayerEndpoint string,
	relayerCommission *apd.Decimal,
) {
	const testTimeout = time.Second * 90

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	bwsc := s.newSwapdWSClient(ctx, defaultXMRMakerSwapdWSEndpoint)
	min := coins.StrToDecimal("0.1")
	offerResp, statusCh, err := bwsc.MakeOfferAndSubscribe(min, xmrmakerProvideAmount,
		exchangeRate, asset, relayerEndpoint, relayerCommission)
	require.NoError(s.T(), err)

	bc := rpcclient.NewClient(ctx, defaultXMRMakerSwapdEndpoint)
	beforeResp, err := bc.GetOffers()
	require.NoError(s.T(), err)

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
			case <-ctx.Done():
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					errCh <- fmt.Errorf("test timed out")
				} else {
					errCh <- fmt.Errorf("make offer context canceled")
				}
				return
			}
		}
	}()

	ac := rpcclient.NewClient(ctx, defaultXMRTakerSwapdEndpoint)
	awsc := s.newSwapdWSClient(ctx, defaultXMRTakerSwapdWSEndpoint)

	// TODO: implement discovery over websockets (#97)
	peerIDs, err := ac.Discover(coins.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(peerIDs))
	assert.Equal(s.T(), peerIDs[0], offerResp.PeerID)

	providesAmt := coins.StrToDecimal("0.05")
	takerStatusCh, err := awsc.TakeOfferAndSubscribe(offerResp.PeerID, offerResp.OfferID, providesAmt)
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

	afterResp, err := bc.GetOffers()
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(beforeResp.Offers)-len(afterResp.Offers))
}

func (s *IntegrationTestSuite) TestRefund_XMRTakerCancels() {
	s.testRefundXMRTakerCancels(types.EthAssetETH)
}

func (s *IntegrationTestSuite) testRefundXMRTakerCancels(asset types.EthAsset) {
	const (
		testTimeout = time.Second * 60
		swapTimeout = 30
	)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	bwsc := s.newSwapdWSClient(ctx, defaultXMRMakerSwapdWSEndpoint)
	offerResp, statusCh, err := bwsc.MakeOfferAndSubscribe(xmrmakerProvideAmount, xmrmakerProvideAmount,
		exchangeRate, asset, "", nil)
	require.NoError(s.T(), err)

	bc := rpcclient.NewClient(ctx, defaultXMRMakerSwapdEndpoint)
	beforeResp, err := bc.GetOffers()
	require.NoError(s.T(), err)

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
			case <-ctx.Done():
				errCh <- fmt.Errorf("make offer context canceled: %w", ctx.Err())
				return
			}
		}
	}()

	ac := rpcclient.NewClient(ctx, defaultXMRTakerSwapdEndpoint)
	awsc := s.newSwapdWSClient(ctx, defaultXMRTakerSwapdWSEndpoint)

	err = ac.SetSwapTimeout(swapTimeout)
	require.NoError(s.T(), err)

	peerIDs, err := ac.Discover(coins.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(peerIDs))
	assert.Equal(s.T(), offerResp.PeerID, peerIDs[0])

	providesAmt := coins.StrToDecimal("0.05")
	takerStatusCh, err := awsc.TakeOfferAndSubscribe(offerResp.PeerID, offerResp.OfferID, providesAmt)
	require.NoError(s.T(), err)

	go func() {
		defer wg.Done()
		for status := range takerStatusCh {
			s.T().Log("> XMRTaker got status:", status)
			if status != types.ETHLocked {
				continue
			}

			s.T().Log("> XMRTaker cancelling swap!")
			exitStatus, err := ac.Cancel(offerResp.OfferID) //nolint:govet
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

	// wait for offer to be re-added
	time.Sleep(time.Second * 2)
	afterResp, err := bc.GetOffers()
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(beforeResp.Offers), len(afterResp.Offers))
}

// TestRefund_XMRMakerCancels_untilAfterT1 tests the case where XMRTaker and XMRMaker
// both lock their funds, but XMRMaker goes offline
// until time t1 in the swap contract passes. This triggers XMRTaker to refund, which XMRMaker will then
// "come online" to see, and he will then refund also.
func (s *IntegrationTestSuite) TestRefund_XMRMakerCancels_untilAfterT1() {
	// Skipping test as it can't guarantee that the refund will happen before the swap completes
	// successfully:  // https://github.com/athanorlabs/atomic-swap/issues/144
	s.T().Skip()
	s.testRefundXMRMakerCancels(7, types.CompletedRefund)
	time.Sleep(time.Second * 5)
}

// TestRefund_XMRMakerCancels_afterIsReady tests the case where XMRTaker and XMRMaker both lock their
// funds, but XMRMaker goes offline until past isReady==true and t0, but comes online before t1. When
// XMRMaker comes back online, he should claim the ETH, causing XMRTaker to also claim the XMR.
func (s *IntegrationTestSuite) TestRefund_XMRMakerCancels_afterIsReady() {
	// Skipping test as it can't guarantee that the refund will happen before the swap completes
	// successfully:  // https://github.com/athanorlabs/atomic-swap/issues/144
	s.T().Skip()
	s.testRefundXMRMakerCancels(30, types.CompletedSuccess)
}

func (s *IntegrationTestSuite) testRefundXMRMakerCancels( //nolint:unused
	swapTimeout uint64,
	expectedExitStatus types.Status,
) {
	const testTimeout = time.Second * 60

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	bc := rpcclient.NewClient(ctx, defaultXMRMakerSwapdEndpoint)
	bwsc := s.newSwapdWSClient(ctx, defaultXMRMakerSwapdWSEndpoint)

	offerResp, statusCh, err := bwsc.MakeOfferAndSubscribe(xmrmakerProvideAmount, xmrmakerProvideAmount,
		exchangeRate, types.EthAssetETH, "", nil)
	require.NoError(s.T(), err)

	beforeResp, err := bc.GetOffers()
	require.NoError(s.T(), err)

	errCh := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for {
			select {
			case status := <-statusCh:
				s.T().Log("> XMRMaker got status:", status)
				if status != types.XMRLocked {
					continue
				}

				s.T().Log("> XMRMaker cancelled swap!")
				exitStatus, err := bc.Cancel(offerResp.OfferID) //nolint:govet
				if err != nil {
					errCh <- err
					return
				}

				if exitStatus != expectedExitStatus {
					errCh <- fmt.Errorf("did not get expected exit status for XMRMaker: got %s, expected %s", exitStatus, expectedExitStatus) //nolint:lll
					return
				}

				s.T().Log("> XMRMaker refunded successfully")
				return
			case <-ctx.Done():
				errCh <- fmt.Errorf("make offer context canceled: %w", ctx.Err())
				return
			}
		}
	}()

	ac := rpcclient.NewClient(ctx, defaultXMRTakerSwapdEndpoint)
	awsc := s.newSwapdWSClient(ctx, defaultXMRTakerSwapdWSEndpoint)

	err = ac.SetSwapTimeout(swapTimeout)
	require.NoError(s.T(), err)

	peerIDs, err := ac.Discover(coins.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(peerIDs))
	providesAmt := coins.StrToDecimal("0.05")
	takerStatusCh, err := awsc.TakeOfferAndSubscribe(offerResp.PeerID, offerResp.OfferID, providesAmt)
	require.NoError(s.T(), err)

	go func() {
		defer wg.Done()

		for status := range takerStatusCh {
			s.T().Log("> XMRTaker got status:", status)
			if status.IsOngoing() {
				continue
			}

			if status != expectedExitStatus {
				errCh <- fmt.Errorf("did not get expected exit status for XMRTaker: got %s, expected %s", status, expectedExitStatus) //nolint:lll
				return
			}

			s.T().Log("> XMRTaker refunded successfully")
			return
		}
	}()

	wg.Wait()

	select {
	case err = <-errCh:
		require.NoError(s.T(), err)
	default:
	}

	afterResp, err := bc.GetOffers()
	require.NoError(s.T(), err)
	if expectedExitStatus != types.CompletedSuccess {
		require.Equal(s.T(), len(beforeResp.Offers), len(afterResp.Offers))
	} else {
		require.Equal(s.T(), 1, len(beforeResp.Offers)-len(afterResp.Offers))
	}
}

// TestAbort_XMRTakerCancels tests the case where XMRTaker cancels the swap before any funds are locked.
// Both parties should abort the swap successfully.
func (s *IntegrationTestSuite) TestAbort_XMRTakerCancels() {
	s.testAbortXMRTakerCancels(types.EthAssetETH)
}

func (s *IntegrationTestSuite) testAbortXMRTakerCancels(asset types.EthAsset) {
	const testTimeout = time.Second * 60

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	bwsc := s.newSwapdWSClient(ctx, defaultXMRMakerSwapdWSEndpoint)

	min := coins.StrToDecimal("0.1")
	offerResp, statusCh, err := bwsc.MakeOfferAndSubscribe(min, xmrmakerProvideAmount,
		exchangeRate, asset, "", nil)
	require.NoError(s.T(), err)

	bc := rpcclient.NewClient(ctx, defaultXMRMakerSwapdEndpoint)
	beforeResp, err := bc.GetOffers()
	require.NoError(s.T(), err)

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
			case <-ctx.Done():
				errCh <- fmt.Errorf("make offer context canceled: %w", ctx.Err())
				return
			}
		}
	}()

	ac := rpcclient.NewClient(ctx, defaultXMRTakerSwapdEndpoint)
	awsc := s.newSwapdWSClient(ctx, defaultXMRTakerSwapdWSEndpoint)

	peerIDs, err := ac.Discover(coins.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(peerIDs))
	assert.Equal(s.T(), offerResp.PeerID, peerIDs[0])

	amount := coins.StrToDecimal("0.05")
	takerStatusCh, err := awsc.TakeOfferAndSubscribe(offerResp.PeerID, offerResp.OfferID, amount)
	require.NoError(s.T(), err)

	go func() {
		defer wg.Done()
		for status := range takerStatusCh {
			s.T().Log("> XMRTaker got status:", status)
			if status != types.ExpectingKeys {
				continue
			}

			s.T().Log("> XMRTaker cancelled swap!")
			exitStatus, err := ac.Cancel(offerResp.OfferID) //nolint:govet
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

	// wait for offer to be re-added
	time.Sleep(time.Second)
	afterResp, err := bc.GetOffers()
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(beforeResp.Offers), len(afterResp.Offers))
}

// This test simulates the case where neither XMRTaker and XMRMaker have
// locked funds yet, and XMRMaker cancels the swap.
// The swap should abort on XMRMaker's side, but might abort *or* refund on XMRTaker's side, in case she ended up
// locking ETH before she was notified that XMRMaker disconnected.
func (s *IntegrationTestSuite) TestAbort_XMRMakerCancels() {
	s.testAbortXMRMakerCancels(types.EthAssetETH)
}

func (s *IntegrationTestSuite) testAbortXMRMakerCancels(asset types.EthAsset) {
	const testTimeout = time.Second * 60

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	bcli := rpcclient.NewClient(ctx, defaultXMRMakerSwapdEndpoint)
	bwsc := s.newSwapdWSClient(ctx, defaultXMRMakerSwapdWSEndpoint)

	offerResp, statusCh, err := bwsc.MakeOfferAndSubscribe(xmrmakerProvideAmount, xmrmakerProvideAmount,
		exchangeRate, asset, "", nil)
	require.NoError(s.T(), err)

	bc := rpcclient.NewClient(ctx, defaultXMRMakerSwapdEndpoint)
	beforeResp, err := bc.GetOffers()
	require.NoError(s.T(), err)

	errCh := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for {
			select {
			case status := <-statusCh:
				s.T().Log("> XMRMaker got status:", status)
				s.T().Log("> XMRMaker cancelled swap!")
				exitStatus, err := bcli.Cancel(offerResp.OfferID) //nolint:govet
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
			case <-ctx.Done():
				errCh <- fmt.Errorf("make offer context canceled: %w", ctx.Err())
				return
			}
		}
	}()

	c := rpcclient.NewClient(ctx, defaultXMRTakerSwapdEndpoint)
	wsc := s.newSwapdWSClient(ctx, defaultXMRTakerSwapdWSEndpoint)

	peerIDs, err := c.Discover(coins.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(peerIDs))

	providesAmount := coins.StrToDecimal("0.05")
	takerStatusCh, err := wsc.TakeOfferAndSubscribe(offerResp.PeerID, offerResp.OfferID, providesAmount)
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

	afterResp, err := bc.GetOffers()
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(beforeResp.Offers), len(afterResp.Offers))
}

// TestError_ShouldOnlyTakeOfferOnce tests the case where two takers try to take the same offer concurrently.
// Only one should succeed, the other should return an error or Abort status.
func (s *IntegrationTestSuite) TestError_ShouldOnlyTakeOfferOnce() {
	s.testErrorShouldOnlyTakeOfferOnce(types.EthAssetETH)
}

func (s *IntegrationTestSuite) testErrorShouldOnlyTakeOfferOnce(asset types.EthAsset) {
	const testTimeout = time.Second * 60
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	bc := rpcclient.NewClient(ctx, defaultXMRMakerSwapdEndpoint)
	offerResp, err := bc.MakeOffer(xmrmakerProvideAmount, xmrmakerProvideAmount, exchangeRate, asset, "", nil)
	require.NoError(s.T(), err)

	ac := rpcclient.NewClient(ctx, defaultXMRTakerSwapdEndpoint)

	peerIDs, err := ac.Discover(coins.ProvidesXMR, defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(peerIDs))
	assert.Equal(s.T(), offerResp.PeerID, peerIDs[0])

	errCh := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		wsc := s.newSwapdWSClient(ctx, defaultXMRTakerSwapdWSEndpoint)

		providesAmount := coins.StrToDecimal("0.05")
		takerStatusCh, err := wsc.TakeOfferAndSubscribe(offerResp.PeerID, offerResp.OfferID, providesAmount) //nolint:govet
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
				return
			}

			s.T().Log("> XMRTaker[0] exited successfully")
			return
		}
	}()

	go func() {
		defer wg.Done()
		wsc := s.newSwapdWSClient(ctx, defaultCharlieSwapdWSEndpoint)

		providesAmount := coins.StrToDecimal("0.05")
		takerStatusCh, err := wsc.TakeOfferAndSubscribe(offerResp.PeerID, offerResp.OfferID, providesAmount) //nolint:govet
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
	case err = <-errCh:
		require.NotNil(s.T(), err)
		s.T().Log("got expected error:", err)
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			s.Fail("Test timed out")
		} else {
			s.Fail("Did not get error from XMRTaker or Charlie")
		}
	}

	select {
	case err = <-errCh:
		s.Failf("Should only have one error!", "Second error: %s", err)
	default:
	}
}

func (s *IntegrationTestSuite) TestSuccess_ConcurrentSwaps() {
	s.testSuccessConcurrentSwaps(types.EthAssetETH)
}

func (s *IntegrationTestSuite) testSuccessConcurrentSwaps(asset types.EthAsset) {
	const numConcurrentSwaps = 10
	const swapTimeout = 30 * numConcurrentSwaps
	const testTimeout = (swapTimeout * 2 * time.Second) + (1 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	ac := rpcclient.NewClient(ctx, defaultXMRTakerSwapdEndpoint)
	err := ac.SetSwapTimeout(swapTimeout)
	require.NoError(s.T(), err)

	type makerTest struct {
		offerID  types.Hash
		statusCh <-chan types.Status
		errCh    chan error
		index    int
	}

	// Create the XMRMaker offers synchronously
	makerTests := make([]*makerTest, numConcurrentSwaps)
	for i := 0; i < numConcurrentSwaps; i++ {
		bwsc := s.newSwapdWSClient(ctx, defaultXMRMakerSwapdWSEndpoint)
		offerResp, statusCh, err := bwsc.MakeOfferAndSubscribe(xmrmakerProvideAmount, xmrmakerProvideAmount, //nolint:govet
			exchangeRate, asset, "", nil)
		require.NoError(s.T(), err)

		s.T().Logf("XMRMaker[%d] made offer %s", i, offerResp.OfferID)

		makerTests[i] = &makerTest{
			offerID:  offerResp.OfferID,
			statusCh: statusCh,
			errCh:    make(chan error, numConcurrentSwaps),
			index:    i,
		}
	}

	bc := rpcclient.NewClient(ctx, defaultXMRMakerSwapdEndpoint)
	beforeResp, err := bc.GetOffers()
	require.NoError(s.T(), err)

	var wg sync.WaitGroup
	wg.Add(2 * numConcurrentSwaps)

	// Track each XMRMaker's status asynchronously
	for _, mkrTest := range makerTests {
		go func(mt *makerTest) {
			defer wg.Done()

			for {
				select {
				case status := <-mt.statusCh:
					s.T().Logf("> XMRMaker[%d] got status: %s", mt.index, status)
					if status.IsOngoing() {
						continue
					}
					if status != types.CompletedSuccess {
						mt.errCh <- fmt.Errorf("XMRMaker[%d] swap did not succeed: %s", mt.index, status)
					}
					return
				case <-ctx.Done():
					mt.errCh <- fmt.Errorf("XMRMaker[%d] context canceled: %w", mt.index, ctx.Err())
					return
				}
			}
		}(mkrTest)
	}

	type takerTest struct {
		statusCh <-chan types.Status
		errCh    chan error
		index    int
	}

	// Create the XMRTakers synchronously
	takerTests := make([]*takerTest, numConcurrentSwaps)
	for i := 0; i < numConcurrentSwaps; i++ {
		awsc := s.newSwapdWSClient(ctx, defaultXMRTakerSwapdWSEndpoint)

		// TODO: implement discovery over websockets (#97)
		peerIDs, err := ac.Discover(coins.ProvidesXMR, defaultDiscoverTimeout) //nolint:govet
		require.NoError(s.T(), err)
		require.Equal(s.T(), 1, len(peerIDs))

		offerID := makerTests[i].offerID
		providesAmount := coins.StrToDecimal("0.05")
		takerStatusCh, err := awsc.TakeOfferAndSubscribe(peerIDs[0], offerID, providesAmount)
		require.NoError(s.T(), err)

		s.T().Logf("XMRTaker[%d] took offer %s", i, offerID)

		takerTests[i] = &takerTest{
			statusCh: takerStatusCh,
			errCh:    make(chan error, numConcurrentSwaps),
			index:    i,
		}
	}

	// Track each XMRTaker's status asynchronously
	for _, tkrTest := range takerTests {
		tkrTest := tkrTest
		go func(tt *takerTest) {
			defer wg.Done()
			for {
				select {
				case status := <-tt.statusCh:
					s.T().Logf("> XMRTaker[%d] got status: %s", tt.index, status)
					if status.IsOngoing() {
						continue
					}
					if status != types.CompletedSuccess {
						tt.errCh <- fmt.Errorf("XMRTaker[%d] did not succeed: %s", tt.index, status)
					}
					return
				case <-ctx.Done():
					tkrTest.errCh <- fmt.Errorf("XMRTaker[%d] context ended: %w", tt.index, ctx.Err())
					return
				}
			}
		}(tkrTest)
	}

	wg.Wait()
	s.T().Logf("All %d XMR makers and takers completed", numConcurrentSwaps)

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

	afterResp, err := bc.GetOffers()
	require.NoError(s.T(), err)
	require.Equal(s.T(), numConcurrentSwaps, len(beforeResp.Offers)-len(afterResp.Offers))
}
