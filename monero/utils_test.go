// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package monero

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWaitForBlocks(t *testing.T) {
	c := CreateWalletClient(t)

	heightBefore, err := c.GetHeight()
	require.NoError(t, err)

	heightAfter, err := WaitForBlocks(context.Background(), c, 2)
	require.NoError(t, err)
	require.GreaterOrEqual(t, heightAfter-heightBefore, uint64(2))
}
