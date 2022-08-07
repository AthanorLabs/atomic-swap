package xmrtaker

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/noot/atomic-swap/common/types"
)

func newTestXMRTaker(t *testing.T) *Instance {
	b := newBackend(t)
	cfg := &Config{
		Backend:  b,
		Basepath: path.Join(t.TempDir(), "xmrtaker"),
	}

	xmrtaker, err := NewInstance(cfg)
	require.NoError(t, err)
	return xmrtaker
}

func TestXMRTaker_InitiateProtocol(t *testing.T) {
	a := newTestXMRTaker(t)
	offer := &types.Offer{
		ExchangeRate: 1,
	}
	s, err := a.InitiateProtocol(3.33, offer)
	require.NoError(t, err)
	require.Equal(t, a.swapStates[offer.GetID()], s)
}
