package offers

import (
	"github.com/athanorlabs/atomic-swap/common/types"
)

// Database contains the db functions used by the offer manager.
type Database interface {
	PutOffer(offer *types.Offer) error
	DeleteOffer(id types.Hash) error
	GetAllOffers() ([]*types.Offer, error)
	ClearAllOffers() error
}
