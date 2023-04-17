package main

import (
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
)

func (s *swapCLITestSuite) Test_lookupToken() {
	c := s.rpcEndpoint()

	// First call triggers a lookup (assuming not cached yet)
	token1, err := lookupToken(c, s.mockDaiAddr())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), token1)

	// Second call hits the cache
	token2, err := lookupToken(c, s.mockDaiAddr())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), token1)

	// same address, since it came from the cache
	require.True(s.T(), token1 == token2)

	invalidAddr := ethcommon.Address{0x1}
	_, err = lookupToken(c, invalidAddr)
	require.ErrorContains(s.T(), err, "no contract code at given address")
}

func (s *swapCLITestSuite) Test_ethAssetSymbol() {
	c := s.rpcEndpoint()
	symbol, err := ethAssetSymbol(c, types.EthAssetETH)
	require.NoError(s.T(), err)
	require.Equal(s.T(), symbol, "ETH")

	symbol, err = ethAssetSymbol(c, types.EthAsset(s.mockTetherAddr()))
	require.NoError(s.T(), err)
	require.Equal(s.T(), symbol, `"USDT"`) // quoted at the current time
}

func (s *swapCLITestSuite) Test_providedAndReceivedSymbols() {
	c := s.rpcEndpoint()

	// 2nd parameter says we are the maker
	providedSym, receivedSym, err := providedAndReceivedSymbols(c, coins.ProvidesXMR, types.EthAssetETH)
	require.NoError(s.T(), err)
	require.Equal(s.T(), providedSym, "XMR")
	require.Equal(s.T(), receivedSym, "ETH")

	// 2nd parameter says we are the taker, but not necessarily that the ETH asset is ETH
	ethAsset := types.EthAsset(s.mockTetherAddr())
	providedSym, receivedSym, err = providedAndReceivedSymbols(c, coins.ProvidesETH, ethAsset)
	require.NoError(s.T(), err)
	require.Equal(s.T(), providedSym, `"USDT"`)
	require.Equal(s.T(), receivedSym, "XMR")
}

func (s *swapCLITestSuite) Test_printOffer() {
	c := s.rpcEndpoint()

	o := types.NewOffer(
		coins.ProvidesXMR,
		coins.StrToDecimal("1.5"),      // maker min
		coins.StrToDecimal("2.5"),      // maker max
		coins.StrToExchangeRate("200"), // 250 USDT per 1 XMR
		types.EthAsset(s.mockTetherAddr()),
	)

	err := printOffer(c, o, 0, "")
	require.NoError(s.T(), err)
}
