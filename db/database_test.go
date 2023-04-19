// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package db

import (
	"errors"
	"testing"
	"time"

	"github.com/ChainSafe/chaindb"
	ethcommon "github.com/ethereum/go-ethereum/common"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/common/vjson"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
)

func init() {
	_ = logging.SetLogLevel("db", "debug")
}

var testPeerID, _ = peer.Decode("12D3KooWQQRJuKTZ35eiHGNPGDpQqjpJSdaxEMJRxi6NWFrrvQVi")

// infoAsJSON converts an Info object to a JSON string. Converting
// the struct to JSON is the easiest way to compare 2 structs for
// equality, as there are many pointer fields.
func infoAsJSON(t *testing.T, info *swap.Info) string {
	jsonData, err := vjson.MarshalStruct(info)
	require.NoError(t, err)
	return string(jsonData)
}

func TestDatabase_OfferTable(t *testing.T) {
	db, err := NewDatabase(&chaindb.Config{
		DataDir:  t.TempDir(),
		InMemory: true,
	})
	require.NoError(t, err)

	// put swap to ensure iterator over offers is ok
	infoA := &swap.Info{
		Version:              swap.CurInfoVersion,
		PeerID:               testPeerID,
		OfferID:              types.Hash{0x1},
		Provides:             coins.ProvidesXMR,
		ProvidedAmount:       coins.StrToDecimal("0.1"),
		ExpectedAmount:       coins.StrToDecimal("1"),
		ExchangeRate:         coins.StrToExchangeRate("0.1"),
		EthAsset:             types.EthAsset{},
		Status:               types.ExpectingKeys,
		LastStatusUpdateTime: time.Now(),
		MoneroStartHeight:    12345,
		StartTime:            time.Now().Add(-30 * time.Minute),
		EndTime:              nil,
		Timeout0:             nil,
		Timeout1:             nil,
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
	badOfferID := types.Hash{0x1, 0x2, 0x3}
	err = db.offerTable.Put(badOfferID[:], []byte(`{"key":"value"}`))
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

	// Put a swap entry tied to the bad offer in the database
	swapEntry := &swap.Info{
		Version:              swap.CurInfoVersion,
		PeerID:               testPeerID,
		OfferID:              badOfferID,
		Provides:             coins.ProvidesXMR,
		ProvidedAmount:       coins.StrToDecimal("0.1"),
		ExpectedAmount:       coins.StrToDecimal("1"),
		ExchangeRate:         coins.StrToExchangeRate("0.1"),
		EthAsset:             types.EthAsset{},
		Status:               types.ExpectingKeys,
		LastStatusUpdateTime: time.Now(),
		MoneroStartHeight:    12345,
		Timeout0:             nil,
		Timeout1:             nil,
		StartTime:            time.Now().Add(-30 * time.Minute),
		EndTime:              nil,
	}
	err = db.PutSwap(swapEntry)
	require.NoError(t, err)

	// Establish a baseline that both the good and bad entries exist before calling GetAllOffers
	exists, err := db.offerTable.Has(goodOffer.ID[:])
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = db.offerTable.Has(badOfferID[:])
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = db.swapTable.Has(badOfferID[:])
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

	exists, err = db.offerTable.Has(badOfferID[:])
	require.NoError(t, err)
	require.False(t, exists) // offer entry was pruned

	exists, err = db.swapTable.Has(badOfferID[:])
	require.NoError(t, err)
	require.False(t, exists) // swap info tied to removed offer pruned
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
	ethAsset := types.EthAsset(ethcommon.HexToAddress("0xa1E32d14AC4B6d8c1791CAe8E9baD46a1E15B7a8"))

	offerA := types.NewOffer(coins.ProvidesXMR, one, one, oneEx, ethAsset)
	err = db.PutOffer(offerA)
	require.NoError(t, err)

	startTime := time.Now().Add(-2 * time.Minute)
	timeout0 := time.Now().Add(30 * time.Minute)
	timeout1 := time.Now().Add(60 * time.Minute)

	infoA := &swap.Info{
		Version:              swap.CurInfoVersion,
		PeerID:               testPeerID,
		OfferID:              offerA.ID,
		Provides:             offerA.Provides,
		ProvidedAmount:       offerA.MinAmount,
		ExpectedAmount:       offerA.MinAmount,
		ExchangeRate:         offerA.ExchangeRate,
		EthAsset:             offerA.EthAsset,
		Status:               types.ContractReady,
		LastStatusUpdateTime: time.Now(),
		MoneroStartHeight:    12345,
		StartTime:            startTime,
		EndTime:              nil,
		Timeout0:             &timeout0,
		Timeout1:             &timeout1,
	}
	err = db.PutSwap(infoA)
	require.NoError(t, err)

	infoB := &swap.Info{
		Version:              swap.CurInfoVersion,
		PeerID:               testPeerID,
		OfferID:              types.Hash{0x2},
		Provides:             coins.ProvidesXMR,
		ProvidedAmount:       coins.StrToDecimal("1.5"),
		ExpectedAmount:       coins.StrToDecimal("0.15"),
		ExchangeRate:         coins.ToExchangeRate(coins.StrToDecimal("0.1")),
		EthAsset:             ethAsset,
		Status:               types.XMRLocked,
		LastStatusUpdateTime: time.Now(),
		MoneroStartHeight:    12345,
		StartTime:            startTime,
		EndTime:              nil,
		Timeout0:             &timeout0,
		Timeout1:             &timeout1,
	}
	err = db.PutSwap(infoB)
	require.NoError(t, err)

	res, err := db.GetSwap(offerA.ID)
	require.NoError(t, err)
	require.Equal(t, infoAsJSON(t, infoA), infoAsJSON(t, res))

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

	startTime := time.Now().Add(-2 * time.Minute)
	timeout0 := time.Now().Add(30 * time.Minute)
	timeout1 := time.Now().Add(60 * time.Minute)

	goodInfo := &swap.Info{
		Version:              swap.CurInfoVersion,
		PeerID:               testPeerID,
		OfferID:              types.Hash{0x1, 0x2, 0x3},
		Provides:             coins.ProvidesXMR,
		ProvidedAmount:       coins.StrToDecimal("1.5"),
		ExpectedAmount:       coins.StrToDecimal("0.15"),
		ExchangeRate:         coins.ToExchangeRate(coins.StrToDecimal("0.1")),
		EthAsset:             types.EthAsset{},
		Status:               types.ETHLocked,
		LastStatusUpdateTime: time.Now(),
		MoneroStartHeight:    12345,
		StartTime:            startTime,
		EndTime:              nil,
		Timeout0:             &timeout0,
		Timeout1:             &timeout1,
	}
	err = db.PutSwap(goodInfo)
	require.NoError(t, err)

	// Put a bad entry in the database
	badInfoID := types.Hash{0x4, 0x5, 0x6}
	err = db.swapTable.Put(badInfoID[:], []byte(`{"key":"value"}`))
	require.NoError(t, err)

	// Establish a baseline that both the good and bad entries exist before calling GetAllSwaps
	exists, err := db.swapTable.Has(goodInfo.OfferID[:])
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = db.swapTable.Has(badInfoID[:])
	require.NoError(t, err)
	require.True(t, exists)

	// Only the good offer should be returned by GetAllSwaps
	swaps, err := db.GetAllSwaps()
	require.NoError(t, err)
	require.Equal(t, 1, len(swaps))
	require.EqualValues(t, goodInfo.OfferID[:], swaps[0].OfferID[:])

	// GetAllSwaps should have pruned the bad swap info entry, but left the good entry
	exists, err = db.swapTable.Has(goodInfo.OfferID[:])
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
	startTime := time.Now().Add(-2 * time.Minute)
	timeout0 := time.Now().Add(30 * time.Minute)
	timeout1 := time.Now().Add(60 * time.Minute)

	infoA := &swap.Info{
		Version:              swap.CurInfoVersion,
		PeerID:               testPeerID,
		OfferID:              id,
		Provides:             coins.ProvidesXMR,
		ProvidedAmount:       coins.StrToDecimal("0.1"),
		ExpectedAmount:       coins.StrToDecimal("1"),
		ExchangeRate:         coins.StrToExchangeRate("0.1"),
		EthAsset:             types.EthAsset{},
		Status:               types.XMRLocked,
		LastStatusUpdateTime: time.Now(),
		MoneroStartHeight:    12345,
		StartTime:            startTime,
		EndTime:              nil,
		Timeout0:             &timeout0,
		Timeout1:             &timeout1,
	}
	err = db.PutSwap(infoA)
	require.NoError(t, err)

	// infoB mostly the same as infoA (same ID, importantly), but with
	// a couple updated fields.
	infoB := new(swap.Info)
	*infoB = *infoA
	infoB.Status = types.CompletedSuccess
	endTime := time.Now()
	infoB.EndTime = &endTime

	err = db.PutSwap(infoB)
	require.NoError(t, err)

	res, err := db.GetSwap(id)
	require.NoError(t, err)
	require.Equal(t, infoAsJSON(t, infoB), infoAsJSON(t, res))
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
