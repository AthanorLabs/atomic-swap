package swap

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/cockroachdb/apd/v3"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
)

var (
	// CurInfoVersion is the latest supported version of a serialised Info struct
	CurInfoVersion, _ = semver.NewVersion("0.2.0")

	errInfoVersionMissing = errors.New("required 'version' field missing in swap Info")
)

type (
	Status = types.Status //nolint:revive
)

// Info contains the details of the swap as well as its status.
type Info struct {
	Version        *semver.Version     `json:"version"`
	ID             types.Hash          `json:"offerID"` // swap offer ID
	Provides       coins.ProvidesCoin  `json:"provides"`
	ProvidedAmount *apd.Decimal        `json:"providedAmount"`
	ExpectedAmount *apd.Decimal        `json:"expectedAmount"`
	ExchangeRate   *coins.ExchangeRate `json:"exchangeRate"`
	EthAsset       types.EthAsset      `json:"ethAsset"`
	Status         Status              `json:"status"`
	// MoneroStartHeight is the Monero block number when the swap begins.
	MoneroStartHeight uint64            `json:"moneroStartHeight"`
	statusCh          chan types.Status `json:"-"`
}

// NewInfo creates a new *Info from the given parameters.
// Note that the swap ID is the same as the offer ID.
func NewInfo(
	id types.Hash,
	provides coins.ProvidesCoin,
	providedAmount, expectedAmount *apd.Decimal,
	exchangeRate *coins.ExchangeRate,
	ethAsset types.EthAsset,
	status Status,
	moneroStartHeight uint64,
	statusCh chan types.Status,
) *Info {
	info := &Info{
		Version:           CurInfoVersion,
		ID:                id,
		Provides:          provides,
		ProvidedAmount:    providedAmount,
		ExpectedAmount:    expectedAmount,
		ExchangeRate:      exchangeRate,
		EthAsset:          ethAsset,
		Status:            status,
		MoneroStartHeight: moneroStartHeight,
		statusCh:          statusCh,
	}
	return info
}

// StatusCh returns the swap's status update channel.
func (i *Info) StatusCh() chan types.Status {
	return i.statusCh
}

// SetStatus ...
func (i *Info) SetStatus(s Status) {
	i.Status = s
}

// UnmarshalInfo deserializes a JSON Info struct, checking the version for compatibility
// before attempting to deserialize the whole blob.
func UnmarshalInfo(jsonData []byte) (*Info, error) {
	ov := struct {
		Version *semver.Version `json:"version"`
	}{}
	if err := json.Unmarshal(jsonData, &ov); err != nil {
		return nil, err
	}
	if ov.Version == nil {
		return nil, errInfoVersionMissing
	}
	if ov.Version.GreaterThan(CurInfoVersion) {
		return nil, fmt.Errorf("info version %q not supported, latest is %q", ov.Version, CurInfoVersion)
	}
	info := &Info{}
	if err := json.Unmarshal(jsonData, info); err != nil {
		return nil, err
	}
	// TODO: Are there additional sanity checks we can perform on the Provided and Received amounts
	//       (or other fields) here when decoding the JSON?
	return info, nil
}
