package swap

import (
	"github.com/athanorlabs/atomic-swap/common/types"
)

type (
	Status = types.Status //nolint:revive
)

// Info contains the details of the swap as well as its status.
type Info struct {
	ID             types.Hash // swap offer ID
	Provides       types.ProvidesCoin
	ProvidedAmount float64
	ReceivedAmount float64
	ExchangeRate   types.ExchangeRate
	EthAsset       types.EthAsset
	Status         Status
	statusCh       <-chan types.Status `json:"-"`
}

// NewInfo creates a new *Info from the given parameters.
// Note that the swap ID is the same as the offer ID.
func NewInfo(
	id types.Hash,
	provides types.ProvidesCoin,
	providedAmount, receivedAmount float64,
	exchangeRate types.ExchangeRate,
	ethAsset types.EthAsset,
	status Status,
	statusCh <-chan types.Status,
) *Info {
	info := &Info{
		ID:             id,
		Provides:       provides,
		ProvidedAmount: providedAmount,
		ReceivedAmount: receivedAmount,
		ExchangeRate:   exchangeRate,
		EthAsset:       ethAsset,
		Status:         status,
		statusCh:       statusCh,
	}
	return info
}

// NewEmptyInfo returns an empty *Info
func NewEmptyInfo() *Info {
	return &Info{}
}

// StatusCh returns the swap's status update channel.
func (i *Info) StatusCh() <-chan types.Status {
	return i.statusCh
}

// SetStatus ...
func (i *Info) SetStatus(s Status) {
	i.Status = s
}
