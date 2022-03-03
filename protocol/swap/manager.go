package swap

import (
	"sync"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
)

var nextID uint64

type (
	Status = common.ExitStatus
)

const (
	Ongoing  = common.Ongoing
	Success  = common.Success
	Refunded = common.Refunded
	Aborted  = common.Aborted
)

// Info contains the details of the swap as well as its status.
type Info struct {
	id             uint64 // ID number of the swap (not the swap offer ID!)
	provides       types.ProvidesCoin
	providedAmount float64
	receivedAmount float64
	exchangeRate   types.ExchangeRate
	status         Status
}

// ID returns the swap ID.
func (i *Info) ID() uint64 {
	if i == nil {
		return 0
	}

	return i.id
}

// Provides returns the coin that was provided for this swap.
func (i *Info) Provides() types.ProvidesCoin {
	if i == nil {
		return ""
	}

	return i.provides
}

// ProvidedAmount returns the amount of coin provided for this swap, in standard units.
func (i *Info) ProvidedAmount() float64 {
	return i.providedAmount
}

// ReceivedAmount returns the amount of coin received for this swap, in standard units.
func (i *Info) ReceivedAmount() float64 {
	return i.receivedAmount
}

// ExchangeRate returns the exchange rate for this swap, represented by a ratio of XMR/ETH.
func (i *Info) ExchangeRate() types.ExchangeRate {
	return i.exchangeRate
}

// Status returns the swap's status.
func (i *Info) Status() Status {
	return i.status
}

// SetReceivedAmount ...
func (i *Info) SetReceivedAmount(a float64) {
	i.receivedAmount = a
}

// SetExchangeRate ...
func (i *Info) SetExchangeRate(r types.ExchangeRate) {
	i.exchangeRate = r
}

// SetStatus ...
func (i *Info) SetStatus(s Status) {
	if i == nil {
		return
	}

	i.status = s
}

// NewInfo ...
func NewInfo(provides types.ProvidesCoin, providedAmount, receivedAmount float64,
	exchangeRate types.ExchangeRate, status Status) *Info {
	info := &Info{
		id:             nextID,
		provides:       provides,
		providedAmount: providedAmount,
		receivedAmount: receivedAmount,
		exchangeRate:   exchangeRate,
		status:         status,
	}
	nextID++
	return info
}

// Manager tracks current and past swaps.
type Manager struct {
	sync.RWMutex
	ongoing *Info
	past    map[uint64]*Info
}

// NewManager ...
func NewManager() *Manager {
	return &Manager{
		past: make(map[uint64]*Info),
	}
}

// AddSwap adds the given swap *Info to the Manager.
func (m *Manager) AddSwap(info *Info) error {
	m.Lock()
	defer m.Unlock()

	switch info.status {
	case common.Ongoing:
		if m.ongoing != nil {
			return errHaveOngoingSwap
		}

		m.ongoing = info
	default:
		m.past[info.id] = info
	}

	return nil
}

// GetPastIDs returns all past swap IDs.
func (m *Manager) GetPastIDs() []uint64 {
	m.RLock()
	defer m.RUnlock()
	ids := make([]uint64, len(m.past))
	i := 0
	for id := range m.past {
		ids[i] = id
		i++
	}
	return ids
}

// GetPastSwap returns a swap's *Info given its ID.
func (m *Manager) GetPastSwap(id uint64) *Info {
	m.RLock()
	defer m.RUnlock()
	return m.past[id]
}

// GetOngoingSwap returns the ongoing swap's *Info, if there is one.
func (m *Manager) GetOngoingSwap() *Info {
	return m.ongoing
}

// CompleteOngoingSwap marks the current ongoing swap as completed.
func (m *Manager) CompleteOngoingSwap() {
	m.Lock()
	defer m.Unlock()
	if m.ongoing == nil {
		return
	}

	m.past[m.ongoing.id] = m.ongoing
	m.ongoing = nil
}
