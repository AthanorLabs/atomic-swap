package extethclient

import (
	"context"
	"testing"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/tests"
	"github.com/stretchr/testify/require"
)

func Test_validateChainID_devSuccess(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	ctx := context.Background()
	chainID, _ := ec.ChainID(ctx)
	err := validateChainID(common.Development, chainID)
	require.NoError(t, err)
}

func Test_validateChainID_mismatchedenv(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	ctx := context.Background()
	chainID, _ := ec.ChainID(ctx)
	err := validateChainID(common.Mainnet, chainID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected chain ID of 1")
}
