package types

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/cockroachdb/apd/v3"
	"golang.org/x/crypto/sha3"

	"github.com/athanorlabs/atomic-swap/coins"
)

var (
	// CurOfferVersion is the latest supported version of a serialised Offer struct
	CurOfferVersion, _ = semver.NewVersion("1.0.0")

	errOfferVersionMissing = errors.New(`required "version" field missing in offer`)
	errOfferIDNotSet       = errors.New(`"offerID" is not set`)
	errExchangeRateNil     = errors.New(`"exchangeRate" is not set`)
	errMinGreaterThanMax   = errors.New(`"minAmount" must be less than or equal to "maxAmount"`)
)

// Offer represents a swap offer
type Offer struct {
	Version      semver.Version      `json:"version"`
	ID           Hash                `json:"offerID"`
	Provides     coins.ProvidesCoin  `json:"provides"`
	MinAmount    *apd.Decimal        `json:"minAmount"` // Min XMR amount
	MaxAmount    *apd.Decimal        `json:"maxAmount"` // Max XMR amount
	ExchangeRate *coins.ExchangeRate `json:"exchangeRate"`
	EthAsset     EthAsset            `json:"ethAsset"`
}

// NewOffer creates and returns an Offer with an initialised ID and Version fields
func NewOffer(
	coin coins.ProvidesCoin,
	minAmount *apd.Decimal,
	maxAmount *apd.Decimal,
	exRate *coins.ExchangeRate,
	ethAsset EthAsset,
) *Offer {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic(err)
	}
	return &Offer{
		Version:      *CurOfferVersion,
		ID:           sha3.Sum256(buf[:]),
		Provides:     coin,
		MinAmount:    minAmount,
		MaxAmount:    maxAmount,
		ExchangeRate: exRate,
		EthAsset:     ethAsset,
	}
}

// String ...
func (o *Offer) String() string {
	return fmt.Sprintf("OfferID:%s Provides:%s MinAmount:%s MaxAmount:%s ExchangeRate:%s EthAsset:%s",
		o.ID,
		o.Provides,
		o.MinAmount.String(),
		o.MaxAmount.String(),
		o.ExchangeRate.String(),
		o.EthAsset,
	)
}

// IsSet returns true if the offer's fields are all set.
func (o *Offer) IsSet() bool {
	return !IsHashZero(o.ID) &&
		o.Provides != "" &&
		o.MinAmount != nil &&
		o.MaxAmount != nil &&
		o.ExchangeRate != nil
}

func (o *Offer) validate() error {
	if IsHashZero(o.ID) {
		return errOfferIDNotSet
	}

	if err := coins.ValidatePositive("minAmount", coins.NumMoneroDecimals, o.MinAmount); err != nil {
		return err
	}
	if err := coins.ValidatePositive("maxAmount", coins.NumMoneroDecimals, o.MaxAmount); err != nil {
		return err
	}

	if o.MinAmount.Cmp(o.MaxAmount) > 0 {
		return errMinGreaterThanMax
	}

	// The JSON decoder for ExchangeRate does validation, but it can't check for nil, as
	// it won't get invoked when the value is not present.
	if o.ExchangeRate == nil {
		return errExchangeRateNil
	}

	return nil
}

// OfferExtra represents extra data that is passed when an offer is made.
type OfferExtra struct {
	StatusCh          chan Status  `json:"-"`
	RelayerEndpoint   string       `json:"relayerEndpoint"`
	RelayerCommission *apd.Decimal `json:"relayerCommission"`
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

// MarshalJSON provides JSON marshalling for the Offer type
func (o *Offer) MarshalJSON() ([]byte, error) {
	if err := o.validate(); err != nil {
		return nil, err
	}
	// Do standard JSON marshal without recursion
	type _Offer Offer
	return json.Marshal((*_Offer)(o))
}

// UnmarshalJSON provides JSON unmarshalling the Offer type
func (o *Offer) UnmarshalJSON(data []byte) error {
	// Do standard JSON marshal without recursion
	type _Offer Offer
	if err := json.Unmarshal(data, (*_Offer)(o)); err != nil {
		return err
	}
	return o.validate()
}
