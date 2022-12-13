package net

import (
	"testing"
	"time"

	"github.com/athanorlabs/atomic-swap/common/types"

	"github.com/stretchr/testify/require"
)

var testAdvertisementSleepDuration = time.Millisecond * 100

func TestHost_Discover(t *testing.T) {
	ha := newHost(t, basicTestConfig(t))
	err := ha.Start()
	require.NoError(t, err)
	hb := newHost(t, basicTestConfig(t))
	err = hb.Start()
	require.NoError(t, err)
	hc := newHost(t, basicTestConfig(t))
	err = hc.Start()
	require.NoError(t, err)

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

	peerIDs, err := hc.Discover(types.ProvidesXMR, time.Second)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(peerIDs), 1)
	require.NotEmpty(t, peerIDs[0])
	require.NotEqual(t, peerIDs[0], hc.PeerID()) // should be ha or hb's ID
}
