package offers

import (
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
)

func Test_Manager(t *testing.T) {
	const numAdd = 10
	const numTake = 5

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := NewMockDatabase(ctrl)

	db.EXPECT().GetAllOffers()

	infoDir := t.TempDir()
	mgr, err := NewManager(infoDir, db)
	require.NoError(t, err)

	for i := 0; i < numAdd; i++ {
		offer := types.NewOffer(types.ProvidesXMR, float64(i), float64(i), types.ExchangeRate(i),
			types.EthAssetETH)
		offerExtra := mgr.AddOffer(offer)
		require.NotNil(t, offerExtra)
	}

	offers := mgr.GetOffers()
	require.Len(t, offers, numAdd)
	for i := 0; i < numTake; i++ {
		offer, offerExtra := mgr.TakeOffer(offers[i].GetID())
		require.NotNil(t, offer)
		require.NotNil(t, offerExtra)
		require.True(t, strings.HasPrefix(offerExtra.InfoFile, infoDir))
	}

	offers = mgr.GetOffers()
	require.Len(t, offers, numAdd-numTake)

	removeIDs := []types.Hash{offers[0].GetID(), offers[2].GetID()}
	mgr.ClearOfferIDs(removeIDs)
	offers = mgr.GetOffers()
	require.Len(t, offers, numAdd-numTake-2)

	mgr.ClearAllOffers()
	offers = mgr.GetOffers()
	require.Len(t, offers, 0)
}
