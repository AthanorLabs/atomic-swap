package extethclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/tests"
)

func Test_validateEthClient_devSuccess(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	ctx := context.Background()
	err := validateEthClient(ctx, common.Development, ec)
	require.NoError(t, err)
}

func Test_validateEthClient_mismatchedEnv(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	ctx := context.Background()

	err := validateEthClient(ctx, common.Mainnet, ec)
	require.Error(t, err)
	assert.ErrorContains(t, err, "expected Ethereum mainnet chain ID (1), but found 1337")

	err = validateEthClient(ctx, common.Stagenet, ec)
	require.Error(t, err)
	assert.ErrorContains(t, err, "expected Goerli chain ID (5), but found 1337")
}
