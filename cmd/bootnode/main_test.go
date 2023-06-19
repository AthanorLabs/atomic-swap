// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/daemon"
	"github.com/athanorlabs/atomic-swap/rpcclient"
)

func getFreePort(t *testing.T) uint16 {
	port, err := common.GetFreeTCPPort()
	require.NoError(t, err)
	return uint16(port)
}

func TestBootnode(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rpcPort := getFreePort(t)
	dataDir := t.TempDir()

	flags := []string{
		"bootnode",
		fmt.Sprintf("--%s=127.0.0.1", flagLibp2pIP),
		fmt.Sprintf("--%s=debug", cliutil.FlagLogLevel),
		fmt.Sprintf("--%s=%s", flagDataDir, dataDir),
		fmt.Sprintf("--%s=%d", flagRPCPort, rpcPort),
		fmt.Sprintf("--%s=0", flagLibp2pPort),
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := cliApp().RunContext(ctx, flags)
		// We may want to replace context.Cancelled with nil at some point in the code
		assert.ErrorIs(t, context.Canceled, err)
	}()

	// Ensure the bootnode fully starts before some basic sanity checks
	daemon.WaitForSwapdStart(t, rpcPort)

	cli := rpcclient.NewClient(ctx, rpcPort)
	versionResp, err := cli.Version()
	require.NoError(t, err)
	require.NotEmpty(t, versionResp.P2PVersion)
	t.Logf("Bootnode p2p version is: %s", versionResp.P2PVersion)
	require.Nil(t, versionResp.SwapCreatorAddr) // bootnode does not know the address

	addressResp, err := cli.Addresses()
	require.NoError(t, err)
	require.Greater(t, len(addressResp.Addrs), 1)

	// We check the contract code below, but we don't need the daemon for that
	cli.Shutdown()
	wg.Wait()
}
