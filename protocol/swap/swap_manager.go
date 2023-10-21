// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package swap provides the management layer used by swapd for tracking current and past
// swaps.
package swap

import (
	"errors"
	"sync"
	"time"

	"github.com/athanorlabs/atomic-swap/common/types"

	"github.com/ChainSafe/chaindb"
)

var errNoSwapWithOfferID = errors.New("unable to find swap with given offer ID")

// Manager tracks current and past swaps.
type Manager interface {
	AddSwap(info *Info) error
	WriteSwapToDB(info *Info) error
	GetPastIDs() ([]types.Hash, error)
	GetPastSwap(types.Hash) (*Info, error)
	GetOngoingSwap(hash types.Hash) (*Info, error)
	GetOngoingSwapSnapshot(types.Hash) (*Info, error)
	GetOngoingSwapOfferIDs() ([]*types.Hash, error)
	GetOngoingSwapsSnapshot() ([]*Info, error)
	CompleteOngoingSwap(info *Info) error
	HasOngoingSwap(types.Hash) bool
	GetStatusChan(offerID types.Hash) <-chan types.Status
	DeleteStatusChan(offerID types.Hash)
	PushNewStatus(offerID types.Hash, status types.Status)
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
	*statusManager
}

var _ Manager = (*manager)(nil)

// NewManager returns a new Manager that uses the given database.
// It loads all ongoing swaps into memory on construction.
// Completed swaps are not loaded into memory.
func NewManager(db Database) (Manager, error) {
	ongoing := make(map[types.Hash]*Info)

	stored, err := db.GetAllSwaps()
	if err != nil {
		return nil, err
	}

	for _, s := range stored {
		if !s.Status.IsOngoing() {
			continue
		}

		ongoing[s.OfferID] = s
	}

	return &manager{
		db:            db,
		ongoing:       ongoing,
		past:          make(map[types.Hash]*Info),
		statusManager: newStatusManager(),
	}, nil
}

// AddSwap adds the given swap *Info to the Manager.
func (m *manager) AddSwap(info *Info) error {
	m.Lock()
	defer m.Unlock()

	switch info.Status.IsOngoing() {
	case true:
		m.ongoing[info.OfferID] = info
	default:
		m.past[info.OfferID] = info
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

		ids[s.OfferID] = struct{}{}
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
	m.past[s.OfferID] = s
	return s, nil
}

// GetOngoingSwap returns the ongoing swap's *Info, if there is one. The
// returned Info structure of an active swap can be modified as the swap's state
// changes and should only be read or written by a single go process.
func (m *manager) GetOngoingSwap(offerID types.Hash) (*Info, error) {
	m.RLock()
	defer m.RUnlock()

	s, has := m.ongoing[offerID]
	if !has {
		return nil, errNoSwapWithOfferID
	}

	return s, nil
}

// GetOngoingSwapSnapshot returns a copy of the ongoing swap's Info, if the
// offerID has an ongoing swap.
func (m *manager) GetOngoingSwapSnapshot(offerID types.Hash) (*Info, error) {
	m.RLock()
	defer m.RUnlock()

	s, has := m.ongoing[offerID]
	if !has {
		return nil, errNoSwapWithOfferID
	}

	sc, err := s.DeepCopy()
	if err != nil {
		return nil, err
	}

	return sc, nil
}

// GetOngoingSwapOfferIDs returns a list of the offer IDs of all ongoing
// swaps.
func (m *manager) GetOngoingSwapOfferIDs() ([]*types.Hash, error) {
	m.RLock()
	defer m.RUnlock()

	offerIDs := make([]*types.Hash, 0, len(m.ongoing))
	for i := range m.ongoing {
		offerIDs = append(offerIDs, &m.ongoing[i].OfferID)
	}

	return offerIDs, nil
}

// GetOngoingSwapsSnapshot returns a copy of all ongoing swaps. If you need to
// modify the result, call `GetOngoingSwap` on the offerID to get the "live"
// Info object.
func (m *manager) GetOngoingSwapsSnapshot() ([]*Info, error) {
	m.RLock()
	defer m.RUnlock()

	swaps := make([]*Info, 0, len(m.ongoing))
	for _, s := range m.ongoing {
		sc, err := s.DeepCopy()
		if err != nil {
			return nil, err
		}
		swaps = append(swaps, sc)
	}

	return swaps, nil
}

// CompleteOngoingSwap marks the current ongoing swap as completed.
func (m *manager) CompleteOngoingSwap(info *Info) error {
	m.Lock()
	defer m.Unlock()

	_, has := m.ongoing[info.OfferID]
	if !has {
		return errNoSwapWithOfferID
	}

	now := time.Now()
	info.EndTime = &now

	m.past[info.OfferID] = info
	delete(m.ongoing, info.OfferID)

	// re-write to db, as status has changed
	return m.db.PutSwap(info)
}

// HasOngoingSwap returns true if the given ID is an ongoing swap.
func (m *manager) HasOngoingSwap(id types.Hash) bool {
	m.RLock()
	defer m.RUnlock()

	_, has := m.ongoing[id]
	return has
}

func (m *manager) getSwapFromDB(id types.Hash) (*Info, error) {
	s, err := m.db.GetSwap(id)
	if errors.Is(chaindb.ErrKeyNotFound, err) {
		return nil, errNoSwapWithOfferID
	}
	if err != nil {
		return nil, err
	}

	return s, nil
}
