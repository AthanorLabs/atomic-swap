package swap

import (
	"github.com/noot/atomic-swap/common"
)

type Status byte 

var (
	Ongoing Status = iota
	Success
	Refunded
	Aborted
)

// Info contains the details of the swap as well as its status.
type Info struct {
	provides common.ProvidesCoin
	providedAmount float64
	receivedAmount float64
	exchangeRate common.ExchangeRate
	status Status
}

func NewInfo(provides common.ProvidesCoin, providedAmount, receivedAmount float64, exchangeRate common.ExchangeRate) *Info {
	retunr &Info{
		provides: provides,
		providedAmount: providedAmount,
		receivedAmount: receivedAmount,
		exchangeRate: exchangeRate,
	}
}

func (i *Info) SetStatus(status Status) {
	i.status = status
}

// Manager tracks current and past swaps.
type Manager struct {
	swaps map[uint64]*info
}

func NewManager() *Manager {
	return &Manager{
		swaps: make(map[uint64]*info),
	}
}

func (m *Manager) AddSwap(id uint64, info *Info) {
	m.swaps[id] = info
}

func (m *Manager) GetSwap(id uint64) *Info {
	return m.swaps[id]
}