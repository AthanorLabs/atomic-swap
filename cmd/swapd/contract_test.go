package main

import (
	"context"
	"testing"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/tests"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestGetOrDeploySwapFactory_DeployNoForwarder(t *testing.T) {
	pk := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)
	tmpDir := t.TempDir()

	chainID, err := ec.ChainID(context.Background())
	require.NoError(t, err)
	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, chainID)
	require.NoError(t, err)

	forwarder, err := deployForwarder(context.Background(), ec, txOpts)
	require.NoError(t, err)

	_, _, err = getOrDeploySwapFactory(
		context.Background(),
		ethcommon.Address{},
		common.Development,
		tmpDir,
		pk,
		ec,
		forwarder,
	)
	require.NoError(t, err)
}

func TestGetOrDeploySwapFactory_DeployForwarderAlso(t *testing.T) {
	pk := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)
	tmpDir := t.TempDir()

	_, _, err := getOrDeploySwapFactory(
		context.Background(),
		ethcommon.Address{},
		common.Development,
		tmpDir,
		pk,
		ec,
		ethcommon.Address{},
	)
	require.NoError(t, err)
}

func TestGetOrDeploySwapFactory_Get(t *testing.T) {
	pk := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)
	tmpDir := t.TempDir()

	chainID, err := ec.ChainID(context.Background())
	require.NoError(t, err)
	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, chainID)
	require.NoError(t, err)

	forwarder, err := deployForwarder(context.Background(), ec, txOpts)
	require.NoError(t, err)

	// deploy and get address
	_, address, err := getOrDeploySwapFactory(
		context.Background(),
		ethcommon.Address{},
		common.Development,
		tmpDir,
		pk,
		ec,
		forwarder,
	)
	require.NoError(t, err)

	_, addr2, err := getOrDeploySwapFactory(
		context.Background(),
		address,
		common.Development,
		tmpDir,
		pk,
		ec,
		ethcommon.Address{},
	)
	require.NoError(t, err)
	require.Equal(t, address, addr2)
}
