package swap

import (
	"errors"
	"sync"

	"github.com/noot/atomic-swap/common"
)

var nextID uint64

type Status byte

const (
	// success set to true upon a successful swap (ie. creation of the XMR wallet)
	// refunded is set to true if the locked ether if refunded from the contract
	// aborted is set to true if the swap aborts before any funds are locked
	Ongoing Status = iota
	Success
	Refunded
	Aborted
)

// Info contains the details of the swap as well as its status.
type Info struct {
	id             uint64 // ID number of the swap (not the swap offer ID!)
	provides       common.ProvidesCoin
	providedAmount float64
	receivedAmount float64
	exchangeRate   common.ExchangeRate
	status         Status
}

func (i *Info) ID() uint64 {
	return i.id
}

func (i *Info) Provides() common.ProvidesCoin {
	return i.provides
}

func (i *Info) ProvidedAmount() float64 {
	return i.providedAmount
}

func (i *Info) ReceivedAmount() float64 {
	return i.receivedAmount
}

func (i *Info) ExchangeRate() common.ExchangeRate {
	return i.exchangeRate
}

func (i *Info) Status() Status {
	return i.status
}

func (i *Info) SetReceivedAmount(a float64) {
	i.receivedAmount = a
}

func (i *Info) SetExchangeRate(r common.ExchangeRate) {
	i.exchangeRate = r
}

func (i *Info) SetStatus(s Status) {
	i.status = s
}

func NewInfo(provides common.ProvidesCoin, providedAmount, receivedAmount float64,
	exchangeRate common.ExchangeRate, status Status) *Info {
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

func NewManager() *Manager {
	return &Manager{
		past: make(map[uint64]*Info),
	}
}

func (m *Manager) AddSwap(info *Info) error {
	m.Lock()
	defer m.Unlock()

	switch info.status {
	case Ongoing:
		if m.ongoing != nil {
			return errors.New("already have ongoing swap")
		}

		m.ongoing = info
	default:
		m.past[info.id] = info
	}

	return nil
}

func (m *Manager) GetPastSwap(id uint64) *Info {
	m.RLock()
	defer m.RUnlock()
	return m.past[id]
}

func (m *Manager) GetOngoingSwap() *Info {
	return m.ongoing
}

func (m *Manager) CompleteOngoingSwap() {
	m.Lock()
	defer m.Unlock()
	if m.ongoing == nil {
		return
	}

	m.past[m.ongoing.id] = m.ongoing
	m.ongoing = nil
}
