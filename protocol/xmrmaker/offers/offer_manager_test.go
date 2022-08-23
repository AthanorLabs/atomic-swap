package offers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/noot/atomic-swap/common/types"
)

func Test_Manager(t *testing.T) {
	const NumAdd = 10
	const NumTake = 5

	infoDir := t.TempDir()
	mgr := NewManager(infoDir)

	for i := 0; i < NumAdd; i++ {
		offer := types.NewOffer(types.ProvidesXMR, float64(i), float64(i), types.ExchangeRate(i),
			types.EthAssetETH)
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

	offers = mgr.GetOffers()
	require.Len(t, offers, NumAdd-NumTake)

	removeIDs := []types.Hash{offers[0].GetID(), offers[2].GetID()}
	mgr.ClearOfferIDs(removeIDs)
	offers = mgr.GetOffers()
	require.Len(t, offers, NumAdd-NumTake-2)

	mgr.ClearAllOffers()
	offers = mgr.GetOffers()
	require.Len(t, offers, 0)
}
