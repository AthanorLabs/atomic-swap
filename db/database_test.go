package db

import (
	"errors"
	"testing"

	"github.com/ChainSafe/chaindb"
	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
)

func init() {
	_ = logging.SetLogLevel("db", "debug")
}

func TestDatabase_OfferTable(t *testing.T) {
	db, err := NewDatabase(&chaindb.Config{
		DataDir:  t.TempDir(),
		InMemory: true,
	})
	require.NoError(t, err)

	// put swap to ensure iterator over offers is ok
	infoA := &swap.Info{
		ID:       types.Hash{0x1},
		Provides: coins.ProvidesXMR,
	}
	err = db.PutSwap(infoA)
	require.NoError(t, err)

	one := coins.StrToDecimal("1")
	oneEx := coins.ToExchangeRate(one)
	offerA := types.NewOffer(coins.ProvidesXMR, one, one, oneEx, types.EthAssetETH)
	err = db.PutOffer(offerA)
	require.NoError(t, err)

	offerB := types.NewOffer(coins.ProvidesXMR, one, one, oneEx, types.EthAssetETH)
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

func TestDatabase_GetAllOffers_InvalidEntry(t *testing.T) {
	db, err := NewDatabase(&chaindb.Config{
		DataDir:  t.TempDir(),
		InMemory: true,
	})
	require.NoError(t, err)
	defer func() { require.NoError(t, db.Close()) }()

	// Put a bad offer that won't deserialize in the database
	badOfferKey := types.Hash{0x1, 0x2, 0x3}
	err = db.offerTable.Put(badOfferKey[:], []byte(`{"key":"value"}`))
	require.NoError(t, err)

	// Put a good offer in the database
	goodOffer := types.NewOffer(
		coins.ProvidesXMR,
		coins.StrToDecimal("1"),
		coins.StrToDecimal("2"),
		coins.ToExchangeRate(coins.StrToDecimal("0.10")),
		types.EthAssetETH,
	)
	err = db.PutOffer(goodOffer)
	require.NoError(t, err)

	// Establish a baseline that both the good and bad entries exist before calling GetAllOffers
	exists, err := db.offerTable.Has(goodOffer.ID[:])
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = db.offerTable.Has(badOfferKey[:])
	require.NoError(t, err)
	require.True(t, exists)

	// Only the good offer should be returned by GetAllOffers
	offers, err := db.GetAllOffers()
	require.NoError(t, err)
	require.Equal(t, 1, len(offers))
	require.EqualValues(t, goodOffer.ID[:], offers[0].ID[:])

	// GetAllOffers should have pruned the bad offer, but left the good offer
	exists, err = db.offerTable.Has(goodOffer.ID[:])
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = db.offerTable.Has(badOfferKey[:])
	require.NoError(t, err)
	require.False(t, exists) // entry was pruned
}

func TestDatabase_SwapTable(t *testing.T) {
	cfg := &chaindb.Config{
		DataDir:  t.TempDir(),
		InMemory: true,
	}

	db, err := NewDatabase(cfg)
	require.NoError(t, err)

	one := coins.StrToDecimal("1")
	oneEx := coins.ToExchangeRate(one)

	offerA := types.NewOffer(coins.ProvidesXMR, one, one, oneEx, types.EthAssetETH)
	err = db.PutOffer(offerA)
	require.NoError(t, err)

	infoA := &swap.Info{
		ID:       types.Hash{0x1},
		Version:  swap.CurInfoVersion,
		Provides: coins.ProvidesXMR,
	}
	err = db.PutSwap(infoA)
	require.NoError(t, err)

	infoB := &swap.Info{
		ID:       types.Hash{0x2},
		Version:  swap.CurInfoVersion,
		Provides: coins.ProvidesXMR,
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

func TestDatabase_GetAllSwaps_InvalidEntry(t *testing.T) {
	db, err := NewDatabase(&chaindb.Config{
		DataDir:  t.TempDir(),
		InMemory: true,
	})
	require.NoError(t, err)

	goodInfo := &swap.Info{
		Version:           swap.CurInfoVersion,
		ID:                types.Hash{0x1, 0x2, 0x3},
		Provides:          coins.ProvidesXMR,
		ProvidedAmount:    coins.StrToDecimal("1.5"),
		ExpectedAmount:    coins.StrToDecimal("0.15"),
		ExchangeRate:      coins.ToExchangeRate(coins.StrToDecimal("0.1")),
		EthAsset:          types.EthAsset{},
		Status:            0,
		MoneroStartHeight: 0,
	}
	err = db.PutSwap(goodInfo)
	require.NoError(t, err)

	// Put a bad entry in the database
	badInfoID := types.Hash{0x4, 0x5, 0x6}
	err = db.swapTable.Put(badInfoID[:], []byte(`{"key":"value"}`))
	require.NoError(t, err)

	// Establish a baseline that both the good and bad entries exist before calling GetAllSwaps
	exists, err := db.swapTable.Has(goodInfo.ID[:])
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = db.swapTable.Has(badInfoID[:])
	require.NoError(t, err)
	require.True(t, exists)

	// Only the good offer should be returned by GetAllSwaps
	swaps, err := db.GetAllSwaps()
	require.NoError(t, err)
	require.Equal(t, 1, len(swaps))
	require.EqualValues(t, goodInfo.ID[:], swaps[0].ID[:])

	// GetAllSwaps should have pruned the bad swap info entry, but left the good entry
	exists, err = db.swapTable.Has(goodInfo.ID[:])
	require.NoError(t, err)
	require.True(t, exists) // entry still exists

	exists, err = db.swapTable.Has(badInfoID[:])
	require.NoError(t, err)
	require.False(t, exists) // entry was pruned
}

func TestDatabase_SwapTable_Update(t *testing.T) {
	cfg := &chaindb.Config{
		DataDir:  t.TempDir(),
		InMemory: true,
	}

	db, err := NewDatabase(cfg)
	require.NoError(t, err)

	id := types.Hash{0x1}
	infoA := &swap.Info{
		ID:       id,
		Provides: coins.ProvidesXMR,
	}
	err = db.PutSwap(infoA)
	require.NoError(t, err)

	infoB := &swap.Info{
		ID:       id,
		Status:   types.CompletedSuccess,
		Provides: coins.ProvidesXMR,
	}

	err = db.PutSwap(infoB)
	require.NoError(t, err)

	res, err := db.GetSwap(id)
	require.NoError(t, err)
	require.Equal(t, infoB, res)
}

func TestDatabase_SwapTable_GetSwap_err(t *testing.T) {
	cfg := &chaindb.Config{
		DataDir:  t.TempDir(),
		InMemory: true,
	}

	db, err := NewDatabase(cfg)
	require.NoError(t, err)

	_, err = db.GetSwap(types.Hash{0x1})
	require.True(t, errors.Is(chaindb.ErrKeyNotFound, err))
}
