package types

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"golang.org/x/crypto/sha3"
)

var (
	// CurOfferVersion is the latest supported version of a serialised Offer struct
	CurOfferVersion, _ = semver.NewVersion("0.1.0")

	errOfferVersionMissing = errors.New("required 'version' field missing in offer")
)

// Offer represents a swap offer
type Offer struct {
	Version       semver.Version `json:"version"`
	ID            Hash           `json:"offerId"`
	Provides      ProvidesCoin   `json:"provides"`
	MinimumAmount float64        `json:"minAmount"`
	MaximumAmount float64        `json:"maxAmount"`
	ExchangeRate  ExchangeRate   `json:"exchangeRate"`
	EthAsset      EthAsset       `json:"ethAsset"`
}

// NewOffer creates and returns an Offer with an initialised ID and Version fields
func NewOffer(coin ProvidesCoin, minAmount float64, maxAmount float64, exRate ExchangeRate, ethAsset EthAsset) *Offer {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic(err)
	}
	return &Offer{
		Version:       *CurOfferVersion,
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

// IsSet returns true if the offer's fields are all set.
func (o *Offer) IsSet() bool {
	return !IsHashZero(o.ID) &&
		o.Provides != "" &&
		o.MinimumAmount != 0 &&
		o.MaximumAmount != 0 &&
		o.ExchangeRate != 0
}

// OfferExtra represents extra data that is passed when an offer is made.
type OfferExtra struct {
	StatusCh          chan Status
	InfoFile          string
	RelayerEndpoint   string
	RelayerCommission float64
}

// UnmarshalOffer deserializes a JSON offer, checking the version for compatibility before
// attempting to deserialize the whole blob.
func UnmarshalOffer(jsonData []byte) (*Offer, error) {
	ov := struct {
		Version *semver.Version `json:"version"`
	}{}
	if err := json.Unmarshal(jsonData, &ov); err != nil {
		return nil, err
	}
	if ov.Version == nil {
		return nil, errOfferVersionMissing
	}
	if ov.Version.GreaterThan(CurOfferVersion) {
		return nil, fmt.Errorf("offer version %q not supported, latest is %q", ov.Version, CurOfferVersion)
	}
	o := &Offer{}
	if err := json.Unmarshal(jsonData, o); err != nil {
		return nil, err
	}
	return o, nil
}
