package extethclient

import (
	"context"
	"testing"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/tests"
	"github.com/stretchr/testify/require"
)

func Test_validateEthClient_devSuccess(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	ctx := context.Background()
	err := validateEthClient(ctx, common.Development, ec)
	require.NoError(t, err)
}

func Test_validateEthClient_misMatchedenv(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	ctx := context.Background()
	err := validateEthClient(ctx, common.Mainnet, ec)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected chain ID of 1")
}
