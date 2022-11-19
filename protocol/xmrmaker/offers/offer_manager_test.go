package offers

import (
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
	db.EXPECT().ClearAllOffers()

	infoDir := t.TempDir()
	mgr, err := NewManager(infoDir, db)
	require.NoError(t, err)

	for i := 0; i < numAdd; i++ {
		offer := types.NewOffer(types.ProvidesXMR, float64(i), float64(i), types.ExchangeRate(i),
			types.EthAssetETH)
		db.EXPECT().PutOffer(offer)
		offerExtra, err := mgr.AddOffer(offer, "", 0)
		require.NoError(t, err)
		require.NotNil(t, offerExtra)
	}

	offers := mgr.GetOffers()
	require.Len(t, offers, numAdd)
	for i := 0; i < numTake; i++ {
		id := offers[i].ID
		offer, offerExtra, err := mgr.TakeOffer(id)
		require.NoError(t, err)
		require.NotNil(t, offer)
		require.NotNil(t, offerExtra)
	}

	offers = mgr.GetOffers()
	require.Len(t, offers, numAdd-numTake)

	removeIDs := []types.Hash{offers[0].ID, offers[2].ID}
	db.EXPECT().DeleteOffer(offers[0].ID)
	db.EXPECT().DeleteOffer(offers[2].ID)
	mgr.ClearOfferIDs(removeIDs)
	offers = mgr.GetOffers()
	require.Len(t, offers, numAdd-numTake-2)

	mgr.ClearAllOffers()
	offers = mgr.GetOffers()
	require.Len(t, offers, 0)
}
