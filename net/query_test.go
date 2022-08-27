package net

import (
	"testing"

	"github.com/athanorlabs/atomic-swap/common/types"

	"github.com/stretchr/testify/require"
)

func TestHost_Query(t *testing.T) {
	ha := newHost(t, defaultPort)
	err := ha.Start()
	require.NoError(t, err)
	hb := newHost(t, defaultPort+1)
	err = hb.Start()
	require.NoError(t, err)

	err = ha.h.Connect(ha.ctx, hb.addrInfo())
	require.NoError(t, err)

	resp, err := ha.Query(hb.addrInfo())
	require.NoError(t, err)
	require.Equal(t, []*types.Offer{}, resp.Offers)
}
