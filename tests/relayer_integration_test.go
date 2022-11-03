package tests

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	rcommon "github.com/athanorlabs/go-relayer/common"
	"github.com/athanorlabs/go-relayer/impls/gsnforwarder"
	"github.com/athanorlabs/go-relayer/relayer"
	rrpc "github.com/athanorlabs/go-relayer/rpc"

	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
)

const (
	relayerBindAddr        = "127.0.0.1:7799"
	defaultRelayerEndpoint = "http://" + relayerBindAddr
	relayerCommission      = float64(0.01)
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

func setupRelayer(t *testing.T) {
	relayerSk := GetTestKeyByIndex(t, 1)
	ec, chainID := NewEthClient(t)

	swapContractAddrStr := os.Getenv(contractAddrEnv)
	require.NotEmptyf(t, swapContractAddrStr, "CONTRACT_ADDR environment variable not set")

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
	ctx := context.Background()

	iforwarder, err := gsnforwarder.NewIForwarder(forwarderAddress, ec)
	require.NoError(t, err)
	fw := gsnforwarder.NewIForwarderWrapped(iforwarder)

	key := rcommon.NewKeyFromPrivateKey(sk)

	cfg := &relayer.Config{
		Ctx:                   ctx,
		EthClient:             ec,
		Forwarder:             fw,
		Key:                   key,
		ChainID:               chainID,
		NewForwardRequestFunc: gsnforwarder.NewIForwarderForwardRequest,
	}

	r, err := relayer.NewRelayer(cfg)
	require.NoError(t, err)

	server, err := rrpc.NewServer(&rrpc.Config{
		Ctx:     context.Background(),
		Address: relayerBindAddr,
		Relayer: r,
	})
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := server.Start()
		require.ErrorIs(t, err, http.ErrServerClosed)
	}()
	t.Cleanup(func() {
		require.NoError(t, server.Stop())
		wg.Wait()
	})
}
