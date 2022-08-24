package net

import (
	"testing"
	"time"

	"github.com/athanorlabs/atomic-swap/common/types"

	"github.com/stretchr/testify/require"
)

var (
	testAdvertisementSleepDuration = time.Second
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

	err = hc.h.Connect(hc.ctx, hb.addrInfo())
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(ha.h.Network().Peers()), 1)
	require.GreaterOrEqual(t, len(hb.h.Network().Peers()), 2)
	require.GreaterOrEqual(t, len(hc.h.Network().Peers()), 1)

	ha.Advertise()
	hb.Advertise()
	hc.Advertise()
	time.Sleep(testAdvertisementSleepDuration)

	peers, err := hc.Discover(types.ProvidesXMR, time.Second)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(peers), 1)
}
