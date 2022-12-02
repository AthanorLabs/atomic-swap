// Package swap provides the management layer used by swapd for tracking current and past
// swaps.
package swap

import (
	"errors"
	"sync"

	"github.com/athanorlabs/atomic-swap/common/types"

	"github.com/ChainSafe/chaindb"
)

var _ Manager = &manager{}

var errNoSwapWithID = errors.New("unable to find swap with given ID")

// Manager tracks current and past swaps.
type Manager interface {
	AddSwap(info *Info) error
	WriteSwapToDB(info *Info) error
	GetPastIDs() ([]types.Hash, error)
	GetPastSwap(types.Hash) (*Info, error)
	GetOngoingSwap(types.Hash) (Info, error)
	GetOngoingSwaps() ([]Info, error)
	CompleteOngoingSwap(info *Info) error
}

// manager implements Manager.
// Note that ongoing swaps are fully populated, but past swaps
// are only stored in memory if they've completed during
// this swapd run, or if they've recently been retrieved.
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

// WriteSwapToDB writes the swap to the database.
func (m *manager) WriteSwapToDB(info *Info) error {
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

	s, err := m.getSwapFromDB(id)
	if err != nil {
		return nil, err
	}

	// cache the swap, since it's recently accessed
	m.past[s.ID] = s
	return s, nil
}

// GetOngoingSwap returns the ongoing swap's *Info, if there is one.
func (m *manager) GetOngoingSwap(id types.Hash) (Info, error) {
	m.RLock()
	defer m.RUnlock()
	s, has := m.ongoing[id]
	if !has {
		return Info{}, errNoSwapWithID
	}

	return *s, nil
}

// GetOngoingSwaps returns all ongoing swaps.
func (m *manager) GetOngoingSwaps() ([]Info, error) {
	m.RLock()
	defer m.RUnlock()
	swaps := make([]Info, len(m.ongoing))
	i := 0
	for _, s := range m.ongoing {
		swaps[i] = *s
		i++
	}
	return swaps, nil
}

// CompleteOngoingSwap marks the current ongoing swap as completed.
func (m *manager) CompleteOngoingSwap(info *Info) error {
	m.Lock()
	defer m.Unlock()
	_, has := m.ongoing[info.ID]
	if !has {
		return errNoSwapWithID
	}

	m.past[info.ID] = info
	delete(m.ongoing, info.ID)

	// re-write to db, as status has changed
	return m.db.PutSwap(info)
}

func (m *manager) getSwapFromDB(id types.Hash) (*Info, error) {
	s, err := m.db.GetSwap(id)
	if errors.Is(chaindb.ErrKeyNotFound, err) {
		return nil, errNoSwapWithID
	}
	if err != nil {
		return nil, err
	}

	return s, nil
}
