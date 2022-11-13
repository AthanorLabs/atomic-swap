package offers

import (
	"errors"
	"sync"

	"github.com/athanorlabs/atomic-swap/common/types"

	logging "github.com/ipfs/go-log"
)

const statusChSize = 6 // the max number of stages a swap can potentially go through

var (
	log = logging.Logger("offers")

	errOfferDoesNotExist = errors.New("offer with given ID does not exist")
)

// Manager synchronises access to the offers map.
type Manager struct {
	mu      sync.RWMutex // synchronises access to the offers map
	offers  map[types.Hash]*offerWithExtra
	dataDir string
	db      Database
}

type offerWithExtra struct {
	offer *types.Offer
	extra *types.OfferExtra
}

// NewManager creates a new offers manager. The passed in dataDir is the directory where the
// recovery file is for each individual swap is stored.
func NewManager(dataDir string, db Database) (*Manager, error) {
	log.Infof("loading offers from db...")
	// load offers from the database, if there are any
	savedOffers, err := db.GetAllOffers()
	if err != nil {
		return nil, err
	}

	offers := make(map[types.Hash]*offerWithExtra)

	for _, offer := range savedOffers {
		extra := &types.OfferExtra{
			StatusCh: make(chan types.Status, statusChSize),
			//InfoFile: pcommon.GetSwapInfoFilepath(dataDir, offer.ID.String()),
		}

		offers[offer.ID] = &offerWithExtra{
			offer: offer,
			extra: extra,
		}

		log.Infof("loaded offer %s from database", offer.ID)
	}

	return &Manager{
		offers:  offers,
		dataDir: dataDir,
		db:      db,
	}, nil
}

// GetOffer returns the offer data structures for the passed ID or nil for both values
// if the offer ID is not found.
func (m *Manager) GetOffer(id types.Hash) (*types.Offer, *types.OfferExtra, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	offer, has := m.offers[id]
	if !has {
		return nil, nil, errOfferDoesNotExist
	}
	return offer.offer, offer.extra, nil
}

// AddOffer adds a new offer to the manager and returns its OffersExtra data
func (m *Manager) AddOffer(
	offer *types.Offer,
	relayerEndpoint string,
	relayerCommission float64,
) (*types.OfferExtra, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := offer.ID
	oe, has := m.offers[id]
	if has {
		return oe.extra, nil
	}

	err := m.db.PutOffer(offer)
	if err != nil {
		return nil, err
	}

	extra := &types.OfferExtra{
		StatusCh: make(chan types.Status, statusChSize),
		//InfoFile:          pcommon.GetSwapInfoFilepath(m.dataDir, id.String()),
		RelayerEndpoint:   relayerEndpoint,
		RelayerCommission: relayerCommission,
	}

	m.offers[id] = &offerWithExtra{
		offer: offer,
		extra: extra,
	}

	return extra, nil
}

// TakeOffer returns any offer with the matching id and removes the offer from the manager. Nil
// for both values is returned when the passed offer id is not currently managed.
func (m *Manager) TakeOffer(id types.Hash) (*types.Offer, *types.OfferExtra, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	offer, has := m.offers[id]
	if !has {
		return nil, nil, errOfferDoesNotExist
	}

	err := m.db.DeleteOffer(id)
	if err != nil {
		return nil, nil, err
	}

	delete(m.offers, id)
	return offer.offer, offer.extra, nil
}

// GetOffers returns all current offers. The returned slice is in random order and will not
// be the same from one invocation to the next.
func (m *Manager) GetOffers() []*types.Offer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	offers := make([]*types.Offer, 0, len(m.offers))
	for _, o := range m.offers {
		offers = append(offers, o.offer)
	}
	return offers
}

// ClearAllOffers clears all offers.
func (m *Manager) ClearAllOffers() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	err := m.db.ClearAllOffers()
	if err != nil {
		return err
	}

	m.offers = make(map[types.Hash]*offerWithExtra)
	return nil
}

// ClearOfferIDs clears the passed in offer IDs if they exist.
func (m *Manager) ClearOfferIDs(ids []types.Hash) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range ids {
		_ = m.db.DeleteOffer(id)
		delete(m.offers, id)
	}
}

// DeleteOffer deletes the offer with the given ID, if it exists.
func (m *Manager) DeleteOffer(id types.Hash) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_ = m.db.DeleteOffer(id)
	delete(m.offers, id)
}

// NumOffers returns the current number of offers.
func (m *Manager) NumOffers() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.offers)
}
