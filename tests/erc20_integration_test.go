// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package tests

import (
	"context"
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/rpcclient"
)

// deploys ERC20Mock.sol and assigns the whole token balance to the XMRTaker default address.
func deployERC20Mock(t *testing.T) ethcommon.Address {
	ctx := context.Background()
	aliceKey, err := ethcrypto.HexToECDSA(common.DefaultPrivKeyXMRTaker)
	require.NoError(t, err)

	ec := extethclient.CreateTestClient(t, aliceKey)
	txOpts, err := ec.TxOpts(ctx)
	require.NoError(t, err)

	const (
		initialTokenBalance = 1000 // standard units
		decimals            = 18
	)
	tenToDecimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(decimals), nil)
	totalSupply := new(big.Int).Mul(big.NewInt(initialTokenBalance*2), tenToDecimals)
	halfSupply := new(big.Int).Mul(big.NewInt(initialTokenBalance), tenToDecimals)

	erc20Addr, erc20Tx, tokenContract, err := contracts.DeployERC20Mock(
		txOpts,
		ec.Raw(),
		"ERC20Mock",
		"MOCK",
		decimals,
		ec.Address(),
		totalSupply,
	)
	require.NoError(t, err)
	MineTransaction(t, ec.Raw(), erc20Tx)

	// Query Charlie's Ethereum address
	charlieCli := rpcclient.NewClient(ctx, defaultCharlieSwapdEndpoint)
	balResp, err := charlieCli.Balances(nil)
	require.NoError(t, err)
	charlieAddr := balResp.EthAddress

	// Transfer half of the supply to Charlie (using Alice's extended ethereum client)
	txOpts, err = ec.TxOpts(ctx)
	require.NoError(t, err)
	tx, err := tokenContract.Transfer(txOpts, charlieAddr, halfSupply)
	require.NoError(t, err)
	MineTransaction(t, ec.Raw(), tx)

	tokenBalReq := &rpctypes.BalancesRequest{
		TokenAddrs: []ethcommon.Address{erc20Addr},
	}

	// verify that the XMR Taker has exactly 1000 tokens
	aliceCli := rpcclient.NewClient(ctx, defaultXMRTakerSwapdEndpoint)
	balResp, err = aliceCli.Balances(tokenBalReq)
	require.NoError(t, err)
	require.Equal(t, "1000", balResp.TokenBalances[0].AsStandardString())

	// verify that Charlie also has exactly 1000 tokens
	balResp, err = charlieCli.Balances(tokenBalReq)
	require.NoError(t, err)
	require.NoError(t, err)
	require.Equal(t, "1000", balResp.TokenBalances[0].AsStandardString())

	return erc20Addr
}

func (s *IntegrationTestSuite) TestXMRTaker_ERC20_Query() {
	s.testXMRTakerQuery(s.testToken)
}

func (s *IntegrationTestSuite) TestSuccess_ERC20_OneSwap() {
	s.testSuccessOneSwap(s.testToken, false)
}

func (s *IntegrationTestSuite) TestRefund_ERC20_XMRTakerCancels() {
	s.testRefundXMRTakerCancels(s.testToken)
}

func (s *IntegrationTestSuite) TestAbort_ERC20_XMRTakerCancels() {
	s.testAbortXMRTakerCancels(s.testToken)
}

func (s *IntegrationTestSuite) TestAbort_ERC20_XMRMakerCancels() {
	s.testAbortXMRMakerCancels(s.testToken)
}

func (s *IntegrationTestSuite) TestError_ERC20_ShouldOnlyTakeOfferOnce() {
	s.testErrorShouldOnlyTakeOfferOnce(s.testToken)
}

func (s *IntegrationTestSuite) TestSuccess_ERC20_ConcurrentSwaps() {
	s.testSuccessConcurrentSwaps(s.testToken)
}
