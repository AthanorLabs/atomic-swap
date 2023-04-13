// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package tests

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
)

func setupXMRTakerAuth(t *testing.T) (*bind.TransactOpts, *ethclient.Client, *ecdsa.PrivateKey) {
	conn, chainID := NewEthClient(t)
	pk, err := ethcrypto.HexToECDSA(common.DefaultPrivKeyXMRTaker)
	require.NoError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(pk, chainID)
	require.NoError(t, err)
	return auth, conn, pk
}

// deploys ERC20Mock.sol and assigns the whole token balance to the XMRTaker default address.
func deployERC20Mock(t *testing.T) ethcommon.Address {
	auth, conn, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := ethcrypto.PubkeyToAddress(*pub)

	decimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	balance := new(big.Int).Mul(big.NewInt(9999999), decimals)
	erc20Addr, erc20Tx, _, err := contracts.DeployERC20Mock(auth, conn, "ERC20Mock", "MOCK", 18, addr, balance)
	require.NoError(t, err)
	_, err = block.WaitForReceipt(context.Background(), conn, erc20Tx.Hash())
	require.NoError(t, err)
	return erc20Addr
}

func (s *IntegrationTestSuite) TestXMRTaker_ERC20_Query() {
	s.testXMRTakerQuery(types.EthAsset(deployERC20Mock(s.T())))
}

func (s *IntegrationTestSuite) TestSuccess_ERC20_OneSwap() {
	s.testSuccessOneSwap(types.EthAsset(deployERC20Mock(s.T())), false)
}

func (s *IntegrationTestSuite) TestRefund_ERC20_XMRTakerCancels() {
	s.testRefundXMRTakerCancels(types.EthAsset(deployERC20Mock(s.T())))
}

func (s *IntegrationTestSuite) TestAbort_ERC20_XMRTakerCancels() {
	s.testAbortXMRTakerCancels(types.EthAsset(deployERC20Mock(s.T())))
}

func (s *IntegrationTestSuite) TestAbort_ERC20_XMRMakerCancels() {
	s.testAbortXMRMakerCancels(types.EthAsset(deployERC20Mock(s.T())))
}

func (s *IntegrationTestSuite) TestError_ERC20_ShouldOnlyTakeOfferOnce() {
	s.testErrorShouldOnlyTakeOfferOnce(types.EthAsset(deployERC20Mock(s.T())))
}

func (s *IntegrationTestSuite) TestSuccess_ERC20_ConcurrentSwaps() {
	s.testSuccessConcurrentSwaps(types.EthAsset(deployERC20Mock(s.T())))
}
