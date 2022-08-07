package types

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"golang.org/x/crypto/sha3"
)

// Hash represents a 32-byte hash
type Hash [32]byte

// String returns the hex-encoded hash
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

// HexToHash decodes a hex-encoded string into a hash
func HexToHash(s string) (Hash, error) {
	h, err := hex.DecodeString(s)
	if err != nil {
		return [32]byte{}, err
	}

	var hash [32]byte
	copy(hash[:], h)
	return hash, nil
}

// Offer represents a swap offer
type Offer struct {
	ID            Hash
	idMu          sync.Mutex `json:"-"` // prevents race conditions on the first call to GetID
	Provides      ProvidesCoin
	MinimumAmount float64
	MaximumAmount float64
	ExchangeRate  ExchangeRate
}

// GetID returns the ID of the offer
func (o *Offer) GetID() Hash {
	o.idMu.Lock()
	defer o.idMu.Unlock()

	if o.ID != [32]byte{} {
		return o.ID
	}

	b, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}

	var buf [8]byte
	_, err = rand.Read(buf[:])
	if err != nil {
		panic(err)
	}

	o.ID = sha3.Sum256(append(b, buf[:]...))
	return o.ID
}

// String ...
func (o *Offer) String() string {
	return fmt.Sprintf("Offer ID=%s Provides=%v MinimumAmount=%v MaximumAmount=%v ExchangeRate=%v",
		o.ID,
		o.Provides,
		o.MinimumAmount,
		o.MaximumAmount,
		o.ExchangeRate,
	)
}

// OfferExtra represents extra data that is passed when an offer is made.
type OfferExtra struct {
	StatusCh chan Status
	InfoFile string
}
