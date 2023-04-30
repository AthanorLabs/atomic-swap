// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package contracts

import (
	"testing"
)

func Test_registerDomainSeparatorIfNeeded(t *testing.T) {
	// ec, _ := tests.NewEthClient(t)
	// ctx := context.Background()
	// privKey := tests.GetMakerTestKey(t)

	// txOpts, err := newTXOpts(ctx, ec, privKey)
	// require.NoError(t, err)

	// forwarderAddr, tx, forwarder, err := gsnforwarder.DeployForwarder(txOpts, ec)
	// require.NoError(t, err)
	// _ = tests.MineTransaction(t, ec, tx)

	// isRegistered, err := isDomainSeparatorRegistered(ctx, ec, forwarderAddr, forwarder)
	// require.NoError(t, err)
	// require.False(t, isRegistered)

	// err = registerDomainSeparatorIfNeeded(ctx, ec, privKey, forwarderAddr)
	// require.NoError(t, err)

	// isRegistered, err = isDomainSeparatorRegistered(ctx, ec, forwarderAddr, forwarder)
	// require.NoError(t, err)
	// require.True(t, isRegistered)
}
