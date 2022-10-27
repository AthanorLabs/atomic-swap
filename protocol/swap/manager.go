package swap

import (
	"sync"

	"github.com/athanorlabs/atomic-swap/common/types"
)

// Manager tracks current and past swaps.
type Manager interface {
	AddSwap(info *Info) error
	GetPastIDs() []types.Hash
	GetPastSwap(types.Hash) *Info
	GetOngoingSwap(types.Hash) *Info
	CompleteOngoingSwap(types.Hash)
}

type manager struct {
	sync.RWMutex
	ongoing map[types.Hash]*Info
	past    map[types.Hash]*Info
}

// NewManager ...
func NewManager() Manager {
	return &manager{
		ongoing: make(map[types.Hash]*Info),
		past:    make(map[types.Hash]*Info),
	}
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

	return nil
}

// GetPastIDs returns all past swap IDs.
func (m *manager) GetPastIDs() []types.Hash {
	m.RLock()
	defer m.RUnlock()
	ids := make([]types.Hash, len(m.past))
	i := 0
	for id := range m.past {
		ids[i] = id
		i++
	}
	return ids
}

// GetPastSwap returns a swap's *Info given its ID.
func (m *manager) GetPastSwap(id types.Hash) *Info {
	m.RLock()
	defer m.RUnlock()
	return m.past[id]
}

// GetOngoingSwap returns the ongoing swap's *Info, if there is one.
func (m *manager) GetOngoingSwap(id types.Hash) *Info {
	m.RLock()
	defer m.RUnlock()
	return m.ongoing[id]
}

// CompleteOngoingSwap marks the current ongoing swap as completed.
func (m *manager) CompleteOngoingSwap(id types.Hash) {
	m.Lock()
	defer m.Unlock()
	s, has := m.ongoing[id]
	if !has {
		return
	}

	m.past[id] = s
	delete(m.ongoing, id)
}
