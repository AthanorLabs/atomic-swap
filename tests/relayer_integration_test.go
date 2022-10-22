package tests

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	rcommon "github.com/AthanorLabs/go-relayer/common"
	"github.com/AthanorLabs/go-relayer/impls/gsnforwarder"
	"github.com/AthanorLabs/go-relayer/relayer"
	rrpc "github.com/AthanorLabs/go-relayer/rpc"

	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
)

var (
	defaultRelayerEndpoint = "http://127.0.0.1:7799"
	relayerCommission      = float64(0.01)
)

func (s *IntegrationTestSuite) Test_Success_ClaimRelayer() {
	setupRelayer(s.T())
	s.testSuccessOneSwap(types.EthAssetETH, defaultRelayerEndpoint, relayerCommission)
}

func (s *IntegrationTestSuite) TestERC20_Success_ClaimRelayer() {
	setupRelayer(s.T())
	s.testSuccessOneSwap(
		types.EthAsset(deployERC20Mock(s.T())),
		defaultRelayerEndpoint,
		relayerCommission,
	)
}

func setupRelayer(t *testing.T) {
	relayerSk := GetTestKeyByIndex(t, 1)
	ec, chainID := NewEthClient(t)

	swapContractAddrStr := os.Getenv(contractAddrEnv)
	if swapContractAddrStr == "" {
		panic("CONTRACT_ADDR env var not set")
	}

	swapContractAddr := ethcommon.HexToAddress(swapContractAddrStr)
	contract, err := contracts.NewSwapFactory(swapContractAddr, ec)
	require.NoError(t, err)

	forwarderAddress, err := contract.TrustedForwarder(&bind.CallOpts{})
	require.NoError(t, err)

	// start relayer
	runRelayer(t, ec, forwarderAddress, relayerSk, chainID)
}

func runRelayer(
	t *testing.T,
	ec *ethclient.Client,
	forwarderAddress ethcommon.Address,
	sk *ecdsa.PrivateKey,
	chainID *big.Int,
) {
	iforwarder, err := gsnforwarder.NewIForwarder(forwarderAddress, ec)
	require.NoError(t, err)
	fw := gsnforwarder.NewIForwarderWrapped(iforwarder)

	key := rcommon.NewKeyFromPrivateKey(sk)

	cfg := &relayer.Config{
		Ctx:                   context.Background(),
		EthClient:             ec,
		Forwarder:             fw,
		Key:                   key,
		ChainID:               chainID,
		NewForwardRequestFunc: gsnforwarder.NewIForwarderForwardRequest,
	}

	r, err := relayer.NewRelayer(cfg)
	require.NoError(t, err)

	rpcCfg := &rrpc.Config{
		Port:    7799,
		Relayer: r,
	}
	server, err := rrpc.NewServer(rpcCfg)
	require.NoError(t, err)

	_ = server.Start()
	t.Cleanup(func() {
		// TODO stop server
	})
}
