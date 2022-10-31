package db

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/protocol/swap"

	"github.com/ChainSafe/chaindb"
)

const (
	offerPrefix = "offer"
	swapPrefix  = "swap"
)

// Database is the persistent datastore used by swapd.
type Database struct {
	offerTable chaindb.Database
	swapTable  chaindb.Database
}

// NewDatabase returns a new *Database.
func NewDatabase(cfg *chaindb.Config) (*Database, error) {
	db, err := chaindb.NewBadgerDB(cfg)
	if err != nil {
		return nil, err
	}

	offerTable := chaindb.NewTable(db, offerPrefix)
	swapTable := chaindb.NewTable(db, swapPrefix)

	return &Database{
		offerTable: offerTable,
		swapTable:  swapTable,
	}, nil
}

// Close flushes and closes the database.
func (db *Database) Close() error {
	err := db.offerTable.Close()
	if err != nil {
		return err
	}

	return nil
}

// PutOffer puts an offer in the database.
func (db *Database) PutOffer(offer *types.Offer) error {
	val, err := json.Marshal(offer)
	if err != nil {
		return err
	}

	key := offer.GetID()
	return db.offerTable.Put(key[:], val)
}

// DeleteOffer deletes an offer from the database.
func (db *Database) DeleteOffer(id types.Hash) error {
	return db.offerTable.Del(id[:])
}

// GetAllOffers returns all offers in the database.
func (db *Database) GetAllOffers() ([]*types.Offer, error) {
	iter := db.offerTable.NewIterator()
	defer iter.Release()

	offers := []*types.Offer{}
	for iter.Valid() {
		// key is the offer ID
		key := iter.Key()
		if !bytes.Contains(key, []byte(offerPrefix)) {
			break
		}

		if len(key) != 32 {
			panic(fmt.Sprintf("key (offer ID) length is not 32: %s", key))
		}

		// value is the encoded offer
		value := iter.Value()
		var offer types.Offer
		err := json.Unmarshal(value, &offer)
		if err != nil {
			return nil, err
		}

		offers = append(offers, &offer)
		iter.Next()
	}

	return offers, nil
}

// ClearAllOffers clears all offers from the database.
func (db *Database) ClearAllOffers() error {
	iter := db.offerTable.NewIterator()
	defer iter.Release()

	for iter.Valid() {
		// key is the offer ID
		key := iter.Key()
		if len(key) != 32 {
			panic("key (offer ID) length is not 32")
		}

		err := db.offerTable.Del(key)
		if err != nil {
			return err
		}
		iter.Next()
	}

	return nil
}

// PutSwap puts the given swap in the database.
func (db *Database) PutSwap(s *swap.Info) error {
	val, err := json.Marshal(s)
	if err != nil {
		return err
	}

	key := s.ID
	return db.swapTable.Put(key[:], val)
}

// HasSwap returns whether the db contains a swap with the given ID.
func (db *Database) HasSwap(id types.Hash) (bool, error) {
	return db.swapTable.Has(id[:])
}

// GetSwap returns a swap with the given ID, if it exists.
func (db *Database) GetSwap(id types.Hash) (*swap.Info, error) {
	value, err := db.swapTable.Get(id[:])
	if err != nil {
		return nil, err
	}

	var s swap.Info
	err = json.Unmarshal(value, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// GetAllSwaps returns all swaps in the database.
func (db *Database) GetAllSwaps() ([]*swap.Info, error) {
	iter := db.swapTable.NewIterator()
	defer iter.Release()

	swaps := []*swap.Info{}
	for iter.Valid() {
		// key is the swap ID
		key := iter.Key()
		if !bytes.Contains(key, []byte(swapPrefix)) {
			break
		}

		if len(key) != 32 {
			panic("key (swap ID) length is not 32")
		}

		// value is the encoded swap
		value := iter.Value()
		var s swap.Info
		err := json.Unmarshal(value, &s)
		if err != nil {
			return nil, err
		}

		swaps = append(swaps, &s)
		iter.Next()
	}

	return swaps, nil
}
