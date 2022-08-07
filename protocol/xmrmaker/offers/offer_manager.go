package offers

import (
	"sync"

	"github.com/noot/atomic-swap/common/types"
	pcommon "github.com/noot/atomic-swap/protocol"
)

// Manager synchronises access to the offers map.
type Manager struct {
	mu       sync.Mutex // synchronises access to the offers map
	offers   map[types.Hash]*offerWithExtra
	basePath string
}

type offerWithExtra struct {
	offer *types.Offer
	extra *types.OfferExtra
}

// NewManager creates a new offers manager. The passed in basePath is the directory where the
// recovery file is for each individual swap is stored.
func NewManager(basePath string) *Manager {
	return &Manager{
		offers:   make(map[types.Hash]*offerWithExtra),
		basePath: basePath,
	}
}

// GetOffer returns the offer data structures for the passed ID or nil for both values
// if the offer ID is not found.
func (m *Manager) GetOffer(id types.Hash) (*types.Offer, *types.OfferExtra) {
	m.mu.Lock()
	defer m.mu.Unlock()

	offer, has := m.offers[id]
	if !has {
		return nil, nil
	}
	return offer.offer, offer.extra
}

// AddOffer adds a new offer to the manager and returns its OffersExtra data
func (m *Manager) AddOffer(o *types.Offer) *types.OfferExtra {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := o.GetID()
	offer, has := m.offers[id]
	if has {
		return offer.extra
	}

	extra := &types.OfferExtra{
		StatusCh: make(chan types.Status, 7), // TODO: Constant for this? How was 7 picked?
		InfoFile: pcommon.GetSwapInfoFilepath(m.basePath),
	}

	m.offers[id] = &offerWithExtra{
		offer: o,
		extra: extra,
	}
	return extra
}

// TakeOffer returns any offer with the matching id and removes the offer from the manager. Nil
// for both values is returned when the passed offer id is not currently managed.
func (m *Manager) TakeOffer(id types.Hash) (*types.Offer, *types.OfferExtra) {
	m.mu.Lock()
	defer m.mu.Unlock()

	offer, has := m.offers[id]
	if !has {
		return nil, nil
	}

	delete(m.offers, id)
	return offer.offer, offer.extra
}

// GetOffers returns all current offers. The returned slice is in random order and will not
// be the same from one invocation to the next.
func (m *Manager) GetOffers() []*types.Offer {
	m.mu.Lock()
	defer m.mu.Unlock()

	offers := make([]*types.Offer, 0, len(m.offers))
	for _, o := range m.offers {
		offers = append(offers, o.offer)
	}
	return offers
}

// ClearOffers clears all offers.
func (m *Manager) ClearOffers() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.offers = make(map[types.Hash]*offerWithExtra)
}
