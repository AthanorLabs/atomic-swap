package swapnet

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
)

func TestHost_Query(t *testing.T) {
	ha := newHost(t, basicTestConfig(t))
	err := ha.Start()
	require.NoError(t, err)
	hb := newHost(t, basicTestConfig(t))
	err = hb.Start()
	require.NoError(t, err)

	err = ha.h.Connect(ha.ctx, hb.h.AddrInfo())
	require.NoError(t, err)

	resp, err := ha.Query(hb.h.PeerID())
	require.NoError(t, err)
	require.Equal(t, []*types.Offer{}, resp.Offers)
}
