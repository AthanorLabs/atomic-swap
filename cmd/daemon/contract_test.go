package main

import (
	"math/big"
	"testing"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/tests"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func TestGetOrDeploySwapFactory(t *testing.T) {
	pk, err := ethcrypto.HexToECDSA(tests.GetTakerTestKey(t))
	require.NoError(t, err)

	ec, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)
	defer ec.Close()

	_, addr, err := getOrDeploySwapFactory(ethcommon.Address{},
		common.Development,
		"/tmp",
		big.NewInt(common.GanacheChainID),
		pk,
		ec,
	)
	require.NoError(t, err)
	t.Log(addr)

	_, addr2, err := getOrDeploySwapFactory(addr,
		common.Development,
		"/tmp",
		big.NewInt(common.GanacheChainID),
		pk,
		ec,
	)
	require.NoError(t, err)
	require.Equal(t, addr, addr2)
}
