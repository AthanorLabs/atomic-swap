package net

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
)

func TestHost_Query(t *testing.T) {
	ha := newHost(t, 0) // OS assigned port
	err := ha.Start()
	require.NoError(t, err)
	hb := newHost(t, 0)
	err = hb.Start()
	require.NoError(t, err)

	err = ha.h.Connect(ha.ctx, hb.addrInfo())
	require.NoError(t, err)

	resp, err := ha.Query(hb.addrInfo())
	require.NoError(t, err)
	require.Equal(t, []*types.Offer{}, resp.Offers)
}
