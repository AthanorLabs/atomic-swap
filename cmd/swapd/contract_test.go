package main

import (
	"context"
	"testing"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/tests"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestGetOrDeploySwapFactory(t *testing.T) {
	pk := tests.GetTakerTestKey(t)
	ec, _ := tests.NewEthClient(t)
	tmpDir := t.TempDir()

	_, addr, err := getOrDeploySwapFactory(
		context.Background(),
		ethcommon.Address{},
		common.Development,
		tmpDir,
		pk,
		ec,
		ethcommon.Address{},
	)
	require.NoError(t, err)
	t.Log(addr)

	_, addr2, err := getOrDeploySwapFactory(
		context.Background(),
		addr,
		common.Development,
		tmpDir,
		pk,
		ec,
		ethcommon.Address{},
	)
	require.NoError(t, err)
	require.Equal(t, addr, addr2)
}
