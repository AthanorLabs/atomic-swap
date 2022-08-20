package net

import (
	"testing"
	"time"

	"github.com/athanorlabs/atomic-swap/common/types"

	"github.com/stretchr/testify/require"
)

func TestHost_Discover(t *testing.T) {
	ha := newHost(t, defaultPort)
	err := ha.Start()
	require.NoError(t, err)
	hb := newHost(t, defaultPort+1)
	err = hb.Start()
	require.NoError(t, err)
	hc := newHost(t, defaultPort+2)
	err = hc.Start()
	require.NoError(t, err)

	defer func() {
		_ = ha.Stop()
		_ = hb.Stop()
		_ = hc.Stop()
	}()

	// connect a + b and b + c, see if c can discover a via DHT
	err = ha.h.Connect(ha.ctx, hb.addrInfo())
	require.NoError(t, err)

	err = hc.h.Connect(ha.ctx, hb.addrInfo())
	require.NoError(t, err)

	ha.Advertise()
	time.Sleep(initialAdvertisementTimeout)

	peers, err := hc.Discover(types.ProvidesXMR, time.Second)
	require.NoError(t, err)
	require.Equal(t, 1, len(peers))
	require.Equal(t, ha.h.ID(), peers[0].ID)
}
