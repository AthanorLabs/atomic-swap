package tests

import (
	"context"
	"time"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/rpcclient"
	"github.com/stretchr/testify/require"
)

const (
	defaultRelayerEndpoint = "http://127.0.0.1:7799"
)

var (
	relayerCommission = coins.StrToDecimal("0.01")
)

func (s *IntegrationTestSuite) Test_Success_ClaimRelayer() {
	s.testSuccessOneSwap(types.EthAssetETH, defaultRelayerEndpoint, relayerCommission)
}

func (s *IntegrationTestSuite) TestERC20_Success_ClaimRelayer() {
	s.testSuccessOneSwap(
		types.EthAsset(deployERC20Mock(s.T())),
		defaultRelayerEndpoint,
		relayerCommission,
	)
}

func (s *IntegrationTestSuite) TestXMRMaker_DiscoverRelayer() {
	ctx := context.Background()
	c := rpcclient.NewClient(ctx, defaultXMRMakerSwapdEndpoint)
	// the sleep for 30 seconds here is because the integration tests start the relayer first,
	// but since it has no peers, it's unable to advertise in the DHT.
	// the next time it'll advertise is after 30 seconds, so this is needed
	// when running this test alone so that the relayer has enough time
	// to try advertising again.
	time.Sleep(time.Second * 30)

	// see https://github.com/AthanorLabs/go-relayer/blob/master/net/host.go#L20
	peerIDs, err := c.Discover("", defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	peerIDs, err = c.Discover("isrelayer", defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(peerIDs))
}

func (s *IntegrationTestSuite) Test_Success_ClaimRelayer_P2p() {
	// use fake endpoint, this will cause the node to fallback to the p2p layer
	s.testSuccessOneSwap(types.EthAssetETH, "http://127.0.0.1:9090", relayerCommission)
}
