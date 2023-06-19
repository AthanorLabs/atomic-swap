// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package common

import (
	"context"
	"testing"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/stretchr/testify/require"
)

func TestConfigDefaultsForEnv(t *testing.T) {
	for _, env := range []Environment{Development, Stagenet, Mainnet} {
		conf := ConfigDefaultsForEnv(env)
		require.Equal(t, env, conf.Env)
		// testing for pointer inequality, each call returns a new instance
		require.True(t, conf != ConfigDefaultsForEnv(env))
	}
}

// Performs a connectivity test to the public bootnodes to ensure that they are
// online. We don't want CI to fail if a bootnode is offline, so just run this
// manually if you think we have issues.
func TestPublicBootnodes(t *testing.T) {
	t.Skip("run manually if needed")
	ctx := context.Background()

	// using go-libp2p directly to avoid circular dependencies
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	require.NoError(t, err)

	failures := 0

	for _, bnStr := range publicBootnodes {
		addrInfo, err := peer.AddrInfoFromString(bnStr)
		require.NoError(t, err)
		h.Peerstore().AddAddrs(addrInfo.ID, addrInfo.Addrs, peerstore.PermanentAddrTTL)
		err = h.Connect(ctx, *addrInfo)
		if err != nil {
			failures++
			t.Logf("OFFLINE: %s (%s)", bnStr, err)
		}
	}

	t.Logf("%d of %d bootnodes are offline", failures, len(publicBootnodes))
	require.Zero(t, failures)
}
