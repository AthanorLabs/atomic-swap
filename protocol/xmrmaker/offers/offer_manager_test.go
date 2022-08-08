package offers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/noot/atomic-swap/common/types"
)

func Test_Manager_simple(t *testing.T) {
	const NumAdd = 10
	const NumTake = 5

	infoDir := t.TempDir()
	mgr := NewManager(infoDir)

	for i := 0; i < NumAdd; i++ {
		offer := types.NewOffer(types.ProvidesXMR, float64(i), float64(i), types.ExchangeRate(i))
		offerExtra := mgr.AddOffer(offer)
		require.NotNil(t, offerExtra)
	}

	offers := mgr.GetOffers()
	require.Len(t, offers, NumAdd)
	for i := 0; i < NumTake; i++ {
		offer, offerExtra := mgr.TakeOffer(offers[i].GetID())
		require.NotNil(t, offer)
		require.NotNil(t, offerExtra)
		require.True(t, strings.HasPrefix(offerExtra.InfoFile, infoDir))
	}
	require.Len(t, mgr.GetOffers(), NumAdd-NumTake)
	mgr.ClearOffers()
	require.Len(t, mgr.GetOffers(), 0)
}
