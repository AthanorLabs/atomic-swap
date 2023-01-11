package tests

import (
	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
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
