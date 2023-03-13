package offers

import (
	"testing"

	"github.com/ChainSafe/chaindb"
	"github.com/cockroachdb/apd/v3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/db"
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
		iDecimal := apd.New(int64(i), 0)
		offer := types.NewOffer(
			coins.ProvidesXMR,
			iDecimal,
			iDecimal,
			coins.ToExchangeRate(iDecimal),
			types.EthAssetETH,
		)
		db.EXPECT().PutOffer(offer)
		offerExtra, err := mgr.AddOffer(offer, false)
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

func Test_Manager_NoErrorDeletingOfferNotOnDisk(t *testing.T) {
	dataDir := t.TempDir()
	testDB, err := db.NewDatabase(&chaindb.Config{DataDir: dataDir})
	require.NoError(t, err)

	mgr, err := NewManager(dataDir, testDB)
	require.NoError(t, err)

	offer := types.NewOffer(
		coins.ProvidesXMR,
		coins.StrToDecimal("1"),
		coins.StrToDecimal("2"),
		coins.ToExchangeRate(coins.StrToDecimal("0.1")),
		types.EthAssetETH,
	)
	offerExtra, err := mgr.AddOffer(offer, false)
	require.NoError(t, err)
	require.NotNil(t, offerExtra)

	// First time we verify that we can trigger a real database error
	err = testDB.Close()
	require.NoError(t, err)
	err = mgr.DeleteOffer(offer.ID)
	require.ErrorContains(t, err, "Closed")

	// Recreate the database and the manager. The offer still exists,
	// because the code above did not succeed in deleting it from disk.
	testDB, err = db.NewDatabase(&chaindb.Config{DataDir: dataDir})
	require.NoError(t, err)
	mgr, err = NewManager(dataDir, testDB)
	require.NoError(t, err)

	// Verify that the entry still exists after restart
	offer2, offer2Extras, err := mgr.GetOffer(offer.ID)
	require.NoError(t, err)
	require.Equal(t, offer.ID, offer2.ID)
	require.NotNil(t, offer2Extras)

	// Deleting the offer should produce no error
	err = mgr.DeleteOffer(offer.ID)
	require.NoError(t, err)

	// Getting the offer fails
	_, _, err = mgr.GetOffer(offer.ID)
	require.ErrorIs(t, err, errOfferDoesNotExist)

	// Double deletion is not an error
	err = mgr.DeleteOffer(offer.ID)
	require.NoError(t, err)
}
