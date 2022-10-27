package swap

import (
	"errors"
	"sync"

	"github.com/athanorlabs/atomic-swap/common/types"
)

var _ Manager = &manager{}

var errNoSwapWithID = errors.New("unable to find swap with given ID")

// Manager tracks current and past swaps.
type Manager interface {
	AddSwap(info *Info) error
	GetPastIDs() ([]types.Hash, error)
	GetPastSwap(types.Hash) (*Info, error)
	GetOngoingSwap(types.Hash) (*Info, error)
	CompleteOngoingSwap(types.Hash) error
}

type manager struct {
	db Database
	sync.RWMutex
	ongoing map[types.Hash]*Info
	past    map[types.Hash]*Info
}

// NewManager returns a new Manager that uses the given database.
// It loads all ongoing swaps into memory on construction.
// Completed swaps are not loaded into memory.
func NewManager(db Database) (*manager, error) {
	ongoing := make(map[types.Hash]*Info)

	stored, err := db.GetAllSwaps()
	if err != nil {
		return nil, err
	}

	for _, s := range stored {
		if !s.Status.IsOngoing() {
			continue
		}

		ongoing[s.ID] = s
	}

	return &manager{
		db:      db,
		ongoing: ongoing,
		past:    make(map[types.Hash]*Info),
	}, nil
}

// AddSwap adds the given swap *Info to the Manager.
func (m *manager) AddSwap(info *Info) error {
	m.Lock()
	defer m.Unlock()

	switch info.Status.IsOngoing() {
	case true:
		m.ongoing[info.ID] = info
	default:
		m.past[info.ID] = info
	}

	return m.db.PutSwap(info)
}

// GetPastIDs returns all past swap IDs.
func (m *manager) GetPastIDs() ([]types.Hash, error) {
	m.RLock()
	defer m.RUnlock()
	ids := make(map[types.Hash]struct{})
	for id := range m.past {
		ids[id] = struct{}{}
	}

	// TODO: do we want to cache all past swaps since we're already fetching them?
	stored, err := m.db.GetAllSwaps()
	if err != nil {
		return nil, err
	}

	for _, s := range stored {
		if s.Status.IsOngoing() {
			continue
		}

		ids[s.ID] = struct{}{}
	}

	idArr := make([]types.Hash, len(ids))
	i := 0
	for id := range ids {
		idArr[i] = id
		i++
	}

	return idArr, nil
}

// GetPastSwap returns a swap's *Info given its ID.
func (m *manager) GetPastSwap(id types.Hash) (*Info, error) {
	m.RLock()
	defer m.RUnlock()
	s, has := m.past[id]
	if has {
		return s, nil
	}

	// TODO: do we want to cache this swap?
	return m.getSwapFromDB(id)
}

// GetOngoingSwap returns the ongoing swap's *Info, if there is one.
func (m *manager) GetOngoingSwap(id types.Hash) (*Info, error) {
	m.RLock()
	defer m.RUnlock()
	s, has := m.ongoing[id]
	if !has {
		return nil, errNoSwapWithID
	}

	return s, nil
}

// CompleteOngoingSwap marks the current ongoing swap as completed.
func (m *manager) CompleteOngoingSwap(id types.Hash) error {
	m.Lock()
	defer m.Unlock()
	s, has := m.ongoing[id]
	if !has {
		return nil
	}

	m.past[id] = s
	delete(m.ongoing, id)

	// re-write to db, as status has changed
	return m.db.PutSwap(s)
}

func (m *manager) getSwapFromDB(id types.Hash) (*Info, error) {
	has, err := m.db.HasSwap(id)
	if err != nil {
		return nil, err
	}

	if !has {
		return nil, errNoSwapWithID
	}

	return m.db.GetSwap(id)
}
