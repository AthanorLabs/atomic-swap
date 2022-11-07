package types

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	//ethcommon "github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

var errInvalidHashString = errors.New("hash string length is not 64")

// Hash represents a 32-byte hash
type Hash [32]byte

// EmptyHash is an empty Hash
var EmptyHash = Hash{}

// String returns the hex-encoded hash
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

// IsZero returns true if the hash is all zeros, otherwise false
func (h Hash) IsZero() bool {
	return h == [32]byte{}
}

// MarshalJSON marshals a Hash into a hex string
func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

// UnmarshalJSON unmarshals a hex string into a Hash
func (h *Hash) UnmarshalJSON(data []byte) error {
	var hexStr string
	err := json.Unmarshal(data, &hexStr)
	if err != nil {
		return err
	}

	if len(hexStr) != 64 {
		return errInvalidHashString
	}

	d, err := hex.DecodeString(hexStr)
	if err != nil {
		return err
	}

	copy(h[:], d[:])
	return nil
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
	Provides      ProvidesCoin
	MinimumAmount float64
	MaximumAmount float64
	ExchangeRate  ExchangeRate
	EthAsset      EthAsset
}

// NewOffer creates and returns an Offer with an initialised id field
func NewOffer(coin ProvidesCoin, minAmount float64, maxAmount float64, exRate ExchangeRate, ethAsset EthAsset) *Offer {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic(err)
	}

	return &Offer{
		ID:            sha3.Sum256(buf[:]),
		Provides:      coin,
		MinimumAmount: minAmount,
		MaximumAmount: maxAmount,
		ExchangeRate:  exRate,
		EthAsset:      ethAsset,
	}
}

// String ...
func (o *Offer) String() string {
	return fmt.Sprintf("Offer ID=%s Provides=%v MinimumAmount=%v MaximumAmount=%v ExchangeRate=%v EthAsset=%v",
		o.ID,
		o.Provides,
		o.MinimumAmount,
		o.MaximumAmount,
		o.ExchangeRate,
		o.EthAsset,
	)
}

// OfferExtra represents extra data that is passed when an offer is made.
type OfferExtra struct {
	StatusCh          chan Status
	InfoFile          string
	RelayerEndpoint   string
	RelayerCommission float64
}
