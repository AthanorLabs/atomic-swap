package contracts

import (
	"context"
	"testing"

	"github.com/athanorlabs/go-relayer/impls/gsnforwarder"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/tests"
)

func Test_registerDomainSeparatorIfNeeded(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	ctx := context.Background()
	privKey := tests.GetMakerTestKey(t)

	txOpts, err := newTXOpts(ctx, ec, privKey)
	require.NoError(t, err)

	forwarderAddr, tx, forwarder, err := gsnforwarder.DeployForwarder(txOpts, ec)
	require.NoError(t, err)
	_ = tests.MineTransaction(t, ec, tx)

	isRegistered, err := isDomainSeparatorRegistered(ctx, ec, forwarderAddr, forwarder)
	require.NoError(t, err)
	require.False(t, isRegistered)

	err = registerDomainSeparatorIfNeeded(ctx, ec, privKey, forwarderAddr)
	require.NoError(t, err)

	isRegistered, err = isDomainSeparatorRegistered(ctx, ec, forwarderAddr, forwarder)
	require.NoError(t, err)
	require.True(t, isRegistered)
}
