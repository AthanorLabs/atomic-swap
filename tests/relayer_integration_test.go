// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package tests

import (
	"context"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/rpcclient"
)

func (s *IntegrationTestSuite) Test_Success_ClaimRelayer() {
	s.testSuccessOneSwap(types.EthAssetETH, true)
}

func (s *IntegrationTestSuite) TestERC20_Success_ClaimRelayer() {
	s.T().Skip("Claiming ERC20 tokens via relayer is not yet supported")
	s.testSuccessOneSwap(types.EthAsset(deployERC20Mock(s.T())), true)
}

func (s *IntegrationTestSuite) TestXMRMaker_DiscoverRelayer() {
	ctx := context.Background()
	c := rpcclient.NewClient(ctx, defaultXMRMakerSwapdEndpoint)

	// see https://github.com/AthanorLabs/go-relayer/blob/master/net/host.go#L20
	peerIDs, err := c.Discover("relayer", defaultDiscoverTimeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(peerIDs))
}
