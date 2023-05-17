package swap

import (
	"sync"

	"github.com/athanorlabs/atomic-swap/common/types"
)

// statusManager provides lookup for the status channels. Status channels are
// ephemeral between runs of swapd.
type statusManager struct {
	mu             sync.Mutex
	statusChannels map[types.Hash]chan Status
}

func newStatusManager() *statusManager {
	return &statusManager{
		mu:             sync.Mutex{},
		statusChannels: make(map[types.Hash]chan Status),
	}
}

// getStatusChan returns any existing status channel or a new status channel for
// reading or writing.
func (sm *statusManager) getStatusChan(offerID types.Hash) chan Status {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	_, ok := sm.statusChannels[offerID]
	if !ok {
		sm.statusChannels[offerID] = newStatusChannel()
	}

	return sm.statusChannels[offerID]
}

// GetStatusChan returns any existing status channel or a new status channel for
// reading only.
func (sm *statusManager) GetStatusChan(offerID types.Hash) <-chan Status {
	return sm.getStatusChan(offerID)
}

// DeleteStatusChan deletes any status channel associated with the offer ID.
func (sm *statusManager) DeleteStatusChan(offerID types.Hash) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.statusChannels, offerID)
}

// PushNewStatus adds a new status to the offer ID's channel
func (sm *statusManager) PushNewStatus(offerID types.Hash, status types.Status) {
	ch := sm.getStatusChan(offerID)
	ch <- status
	// If the status is not ongoing, existing subscribers will get the status
	// via the channel since they already have a reference to it. New
	// subscribers will get the final status from the past swaps map.
	if !status.IsOngoing() {
		sm.DeleteStatusChan(offerID)
	}
}

// newStatusChannel creates a status channel using the the correct size
func newStatusChannel() chan Status {
	// The channel size should be large enough to handle the max number of
	// stages a swap can potentially go through.
	const statusChSize = 6
	ch := make(chan Status, statusChSize)
	return ch
}
