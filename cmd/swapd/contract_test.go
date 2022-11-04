package main

import (
	"context"
	"testing"

	"github.com/athanorlabs/atomic-swap/common"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/tests"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestGetOrDeploySwapFactory_DeployNoForwarder(t *testing.T) {
	pk := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)
	tmpDir := t.TempDir()

	forwarder, err := contracts.DeployGSNForwarderWithKey(context.Background(), ec, pk)
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

	forwarder, err := contracts.DeployGSNForwarderWithKey(context.Background(), ec, pk)
	require.NoError(t, err)
	t.Log(forwarder)

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
