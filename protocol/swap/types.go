package swap

import (
	"github.com/Masterminds/semver/v3"

	"github.com/athanorlabs/atomic-swap/common/types"
)

// CurInfoVersion is the latest supported version of a serialised Info struct
var CurInfoVersion, _ = semver.NewVersion("v0.1.0")

type (
	Status = types.Status //nolint:revive
)

// Info contains the details of the swap as well as its status.
type Info struct {
	Version        *semver.Version     `json:"version"`
	ID             types.Hash          `json:"offer_id"` // swap offer ID
	Provides       types.ProvidesCoin  `json:"provides"`
	ProvidedAmount float64             `json:"provided_amount"`
	ReceivedAmount float64             `json:"received_amount"`
	ExchangeRate   types.ExchangeRate  `json:"exchange_rate"`
	EthAsset       types.EthAsset      `json:"eth_asset"`
	Status         Status              `json:"status"`
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
		Version:        CurInfoVersion,
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
