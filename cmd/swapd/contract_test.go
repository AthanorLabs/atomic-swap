package main

import (
	"context"
	"testing"

	"github.com/athanorlabs/atomic-swap/common"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/tests"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestGetOrDeploySwapFactory_DeployNoForwarder(t *testing.T) {
	pk := tests.GetTakerTestKey(t)
	ec := extethclient.CreateTestClient(t, pk)
	tmpDir := t.TempDir()

	forwarder, err := contracts.DeployGSNForwarderWithKey(context.Background(), ec.Raw(), pk)
	require.NoError(t, err)

	_, err = getOrDeploySwapFactory(
		context.Background(),
		ethcommon.Address{},
		common.Development,
		tmpDir,
		ec,
		forwarder,
	)
	require.NoError(t, err)
}

func TestGetOrDeploySwapFactory_DeployForwarderAlso(t *testing.T) {
	pk := tests.GetTakerTestKey(t)
	ec := extethclient.CreateTestClient(t, pk)
	tmpDir := t.TempDir()

	_, err := getOrDeploySwapFactory(
		context.Background(),
		ethcommon.Address{},
		common.Development,
		tmpDir,
		ec,
		ethcommon.Address{},
	)
	require.NoError(t, err)
}

func TestGetOrDeploySwapFactory_Get(t *testing.T) {
	pk := tests.GetTakerTestKey(t)
	ec := extethclient.CreateTestClient(t, pk)
	tmpDir := t.TempDir()

	forwarder, err := contracts.DeployGSNForwarderWithKey(context.Background(), ec.Raw(), pk)
	require.NoError(t, err)
	t.Log(forwarder)

	// deploy and get address
	address, err := getOrDeploySwapFactory(
		context.Background(),
		ethcommon.Address{},
		common.Development,
		tmpDir,
		ec,
		forwarder,
	)
	require.NoError(t, err)

	addr2, err := getOrDeploySwapFactory(
		context.Background(),
		address,
		common.Development,
		tmpDir,
		ec,
		ethcommon.Address{},
	)
	require.NoError(t, err)
	require.Equal(t, address, addr2)
}
