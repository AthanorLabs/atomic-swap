package types

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

// Hash represents a 32-byte hash
type Hash [32]byte

// String returns the hex-encoded hash
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

// IsZero returns true if the hash is all zeros, otherwise false
func (h Hash) IsZero() bool {
	return h == [32]byte{}
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
	id            Hash
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
		id:            sha3.Sum256(buf[:]),
		Provides:      coin,
		MinimumAmount: minAmount,
		MaximumAmount: maxAmount,
		ExchangeRate:  exRate,
		EthAsset:      ethAsset,
	}
}

// GetID returns the ID of the offer
func (o *Offer) GetID() Hash {
	if o.id.IsZero() {
		panic("offer was improperly initialised")
	}
	return o.id
}

// String ...
func (o *Offer) String() string {
	return fmt.Sprintf("Offer ID=%s Provides=%v MinimumAmount=%v MaximumAmount=%v ExchangeRate=%v EthAsset=%v",
		o.id,
		o.Provides,
		o.MinimumAmount,
		o.MaximumAmount,
		o.ExchangeRate,
		o.EthAsset,
	)
}

// MarshalJSON is a custom JSON marshaller for Offer which enables serialisation of the private id field
func (o Offer) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID            string
		Provides      ProvidesCoin
		MinimumAmount float64
		MaximumAmount float64
		ExchangeRate  ExchangeRate
		EthAsset      string
	}{
		ID:            o.id.String(),
		Provides:      o.Provides,
		MinimumAmount: o.MinimumAmount,
		MaximumAmount: o.MaximumAmount,
		ExchangeRate:  o.ExchangeRate,
		EthAsset:      ethcommon.Address(o.EthAsset).Hex(),
	})
}

// UnmarshalJSON is a custom JSON marshaller for Offer which enables deserialization of the private id field
func (o *Offer) UnmarshalJSON(data []byte) error {
	ou := &struct {
		ID            string
		Provides      ProvidesCoin
		MinimumAmount float64
		MaximumAmount float64
		ExchangeRate  ExchangeRate
		EthAsset      string
	}{}
	if err := json.Unmarshal(data, &ou); err != nil {
		return err
	}
	id, err := hex.DecodeString(ou.ID)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Offer ID err=%w", err)
	}
	if len(id) != len(o.id) {
		return fmt.Errorf("offer ID has invalid length=%d", len(id))
	}
	copy(o.id[:], id)
	o.Provides = ou.Provides
	o.MinimumAmount = ou.MinimumAmount
	o.MaximumAmount = ou.MaximumAmount
	o.ExchangeRate = ou.ExchangeRate
	if ou.EthAsset == "" {
		// Default to EthAssetETH
		o.EthAsset = EthAssetETH
	} else {
		o.EthAsset = EthAsset(ethcommon.HexToAddress(ou.EthAsset))
	}
	return nil
}

// OfferExtra represents extra data that is passed when an offer is made.
type OfferExtra struct {
	StatusCh chan Status
	InfoFile string
}
