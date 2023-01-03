package xmrtaker

import (
	"path"
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
)

func newTestXMRTaker(t *testing.T) *Instance {
	b := newBackend(t)
	cfg := &Config{
		Backend: b,
		DataDir: path.Join(t.TempDir(), "xmrtaker"),
	}

	xmrtaker, err := NewInstance(cfg)
	require.NoError(t, err)
	return xmrtaker
}

func TestXMRTaker_InitiateProtocol(t *testing.T) {
	a := newTestXMRTaker(t)
	zero := new(apd.Decimal)
	one := apd.New(1, 0)
	offer := types.NewOffer(types.ProvidesETH, zero, zero, types.ToExchangeRate(one), types.EthAssetETH)
	providesAmount := apd.New(333, -2) // 3.33
	s, err := a.InitiateProtocol(providesAmount, offer)
	require.NoError(t, err)
	require.Equal(t, a.swapStates[offer.ID], s)
}
