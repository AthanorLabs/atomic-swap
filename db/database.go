package db

import (
	"encoding/json"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/protocol/swap"

	"github.com/ChainSafe/chaindb"
)

const (
	offerPrefix = "offer"
	swapPrefix  = "swap"
	idLength    = 32
)

// Database is the persistent datastore used by swapd.
type Database struct {
	// offerTable is a key-value store where all the keys are prefixed by offerPrefix
	// in the underlying database.
	// the key is the 32-byte offer ID and the value is a JSON-marshalled *types.Offer.
	// offerTable entries are stored when offers are made by swapd.
	// they are removed when the offer is taken.
	offerTable chaindb.Database
	// swapTable is a key-value store where all the keys are prefixed by swapPrefix
	// in the underlying database.
	// the key is the 32-byte swap ID (which is the same as the ID of the offer taken
	// to start the swap) and the value is a JSON-marshalled *swap.Info.
	// swapTable entries are added when a swap begins, and they are never deleted;
	// only their `Status` field within *swap.Info may be updated.
	swapTable chaindb.Database
}

// NewDatabase returns a new *Database.
func NewDatabase(cfg *chaindb.Config) (*Database, error) {
	db, err := chaindb.NewBadgerDB(cfg)
	if err != nil {
		return nil, err
	}

	return &Database{
		offerTable: chaindb.NewTable(db, offerPrefix),
		swapTable:  chaindb.NewTable(db, swapPrefix),
	}, nil
}

// Close flushes and closes the database.
func (db *Database) Close() error {
	err := db.offerTable.Close()
	if err != nil {
		return err
	}

	err = db.swapTable.Close()
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

	key := offer.ID
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

	var offers []*types.Offer
	for iter.Valid() {
		key := iter.Key()

		// if the key becomes longer than 32, we're not iterating over offers
		if len(key) > idLength {
			break
		}

		encodedOffer := iter.Value()
		offer, err := types.UnmarshalOffer(encodedOffer)
		if err != nil {
			return nil, err
		}

		offers = append(offers, offer)
		iter.Next()
	}

	return offers, nil
}

// ClearAllOffers clears all offers from the database.
func (db *Database) ClearAllOffers() error {
	iter := db.offerTable.NewIterator()
	defer iter.Release()

	for iter.Valid() {
		offerID := iter.Key()
		err := db.offerTable.Del(offerID)
		if err != nil {
			return err
		}
		iter.Next()
	}

	return nil
}

// PutSwap puts the given swap in the database.
// If a swap with the same ID is already in the database, it overwrites it.
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
// It returns an error if it doesn't exist.
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

	var swaps []*swap.Info
	for iter.Valid() {
		key := iter.Key()

		// if the key becomes longer than 32, we're not iterating over swaps
		if len(key) > idLength {
			break
		}

		// value is the encoded swap
		encodedSwap := iter.Value()
		s, err := swap.UnmarshalInfo(encodedSwap)
		if err != nil {
			return nil, err
		}

		swaps = append(swaps, s)
		iter.Next()
	}

	return swaps, nil
}
