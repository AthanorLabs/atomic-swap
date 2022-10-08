package db

import (
	"encoding/json"

	"github.com/athanorlabs/atomic-swap/common/types"

	"github.com/ChainSafe/chaindb"
)

var (
	offerPrefix = "offer"
)

// Database is the persistent datastore used by swapd.
type Database struct {
	offerTable chaindb.Database
}

// NewDatabase returns a new *Database.
func NewDatabase(cfg *chaindb.Config) (*Database, error) {
	db, err := chaindb.NewBadgerDB(cfg)
	if err != nil {
		return nil, err
	}

	offerTable := chaindb.NewTable(db, offerPrefix)

	return &Database{
		offerTable: offerTable,
	}, nil
}

// Close flushes and closes the database.
func (db *Database) Close() error {
	err := db.offerTable.Flush()
	if err != nil {
		return err
	}

	err = db.offerTable.Close()
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
	for iter.Next() {
		// key is the offer ID
		key := iter.Key()
		if len(key) != 32 {
			panic("key (offer ID) length is not 32")
		}

		// value is the encoded offer
		value := iter.Value()
		var offer types.Offer
		err := json.Unmarshal(value, &offer)
		if err != nil {
			return nil, err
		}

		offers = append(offers, &offer)
	}

	return offers, nil
}

// ClearAllOffers clears all offers from the database.
func (db *Database) ClearAllOffers() error {
	iter := db.offerTable.NewIterator()
	defer iter.Release()

	for iter.Next() {
		// key is the offer ID
		key := iter.Key()
		if len(key) != 32 {
			panic("key (offer ID) length is not 32")
		}

		err := db.offerTable.Del(key)
		if err != nil {
			return err
		}
	}

	return nil
}
