package db

import (
	"testing"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/protocol/swap"

	"github.com/ChainSafe/chaindb"
	"github.com/stretchr/testify/require"
)

func TestDatabase_OfferTable(t *testing.T) {
	cfg := &chaindb.Config{
		DataDir:  t.TempDir(),
		InMemory: true,
	}

	db, err := NewDatabase(cfg)
	require.NoError(t, err)

	// put swap to ensure iterator over offers is ok
	infoA := &swap.Info{
		ID: types.Hash{0x1},
	}
	err = db.PutSwap(infoA)
	require.NoError(t, err)

	offerA := types.NewOffer(
		types.ProvidesXMR,
		float64(1),
		float64(1),
		types.ExchangeRate(1),
		types.EthAssetETH,
	)
	err = db.PutOffer(offerA)
	require.NoError(t, err)

	offerB := types.NewOffer(
		types.ProvidesXMR,
		float64(1),
		float64(1),
		types.ExchangeRate(1),
		types.EthAssetETH,
	)
	err = db.PutOffer(offerB)
	require.NoError(t, err)

	offers, err := db.GetAllOffers()
	require.NoError(t, err)
	require.Equal(t, 2, len(offers))

	err = db.ClearAllOffers()
	require.NoError(t, err)

	offers, err = db.GetAllOffers()
	require.NoError(t, err)
	require.Equal(t, 0, len(offers))
}

func TestDatabase_SwapTable(t *testing.T) {
	cfg := &chaindb.Config{
		DataDir:  t.TempDir(),
		InMemory: true,
	}

	db, err := NewDatabase(cfg)
	require.NoError(t, err)

	offerA := types.NewOffer(
		types.ProvidesXMR,
		float64(1),
		float64(1),
		types.ExchangeRate(1),
		types.EthAssetETH,
	)
	err = db.PutOffer(offerA)
	require.NoError(t, err)

	infoA := &swap.Info{
		ID: types.Hash{0x1},
	}
	err = db.PutSwap(infoA)
	require.NoError(t, err)

	infoB := &swap.Info{
		ID: types.Hash{0x2},
	}
	err = db.PutSwap(infoB)
	require.NoError(t, err)

	res, err := db.GetSwap(types.Hash{0x1})
	require.NoError(t, err)
	require.Equal(t, infoA, res)

	swaps, err := db.GetAllSwaps()
	require.NoError(t, err)
	require.Equal(t, 2, len(swaps))
}
