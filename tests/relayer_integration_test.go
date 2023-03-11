package tests

import (
	"context"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/relayer"
	"github.com/athanorlabs/atomic-swap/rpcclient"
)

var (
	relayerFee = relayer.MinRelayerFeeEth
)

func (s *IntegrationTestSuite) Test_Success_ClaimRelayer() {
	s.testSuccessOneSwap(types.EthAssetETH, relayerFee)
}

func (s *IntegrationTestSuite) TestERC20_Success_ClaimRelayer() {
	s.T().Skip("Claiming ERC20 tokens via relayer is not yet supported")
	s.testSuccessOneSwap(
		types.EthAsset(deployERC20Mock(s.T())),
		relayerFee,
	)
}

func (s *IntegrationTestSuite) TestXMRMaker_DiscoverRelayer() {
	ctx := context.Background()
	c := rpcclient.NewClient(ctx, defaultXMRMakerSwapdEndpoint)

	// see https://github.com/AthanorLabs/go-relayer/blob/master/net/host.go#L20
	peerIDs, err := c.Discover("relayer", defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(peerIDs))
}
