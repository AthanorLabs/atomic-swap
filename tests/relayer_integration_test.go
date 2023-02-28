package tests

import (
	"context"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/rpcclient"
)

const (
	defaultRelayerEndpoint = "http://127.0.0.1:7799"
)

var (
	relayerFee = common.DefaultRelayerFee
)

func (s *IntegrationTestSuite) Test_Success_ClaimRelayer() {
	s.testSuccessOneSwap(types.EthAssetETH, defaultRelayerEndpoint, relayerFee)
}

func (s *IntegrationTestSuite) TestERC20_Success_ClaimRelayer() {
	s.testSuccessOneSwap(
		types.EthAsset(deployERC20Mock(s.T())),
		defaultRelayerEndpoint,
		relayerFee,
	)
}

func (s *IntegrationTestSuite) TestXMRMaker_DiscoverRelayer() {
	ctx := context.Background()
	c := rpcclient.NewClient(ctx, defaultXMRMakerSwapdEndpoint)

	// see https://github.com/AthanorLabs/go-relayer/blob/master/net/host.go#L20
	peerIDs, err := c.Discover("isrelayer", defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(peerIDs))
}

func (s *IntegrationTestSuite) Test_Success_ClaimRelayer_P2p() {
	// use fake endpoint, this will cause the node to fallback to the p2p layer
	s.testSuccessOneSwap(types.EthAssetETH, "http://127.0.0.1:9090", relayerFee)
}
