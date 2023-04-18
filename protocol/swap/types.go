// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package swap

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/cockroachdb/apd/v3"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/common/vjson"
)

var (
	// CurInfoVersion is the latest supported version of a serialised Info struct
	CurInfoVersion, _ = semver.NewVersion("0.3.0")

	errInfoVersionMissing = errors.New("required 'version' field missing in swap Info")
)

type (
	Status = types.Status //nolint:revive
)

// Info contains the details of the swap as well as its status.
type Info struct {
	Version        *semver.Version     `json:"version"`
	PeerID         peer.ID             `json:"peerID" validate:"required"`
	OfferID        types.Hash          `json:"offerID" validate:"required"`
	Provides       coins.ProvidesCoin  `json:"provides" validate:"required"`
	ProvidedAmount *apd.Decimal        `json:"providedAmount" validate:"required"`
	ExpectedAmount *apd.Decimal        `json:"expectedAmount" validate:"required"`
	ExchangeRate   *coins.ExchangeRate `json:"exchangeRate" validate:"required"`
	EthAsset       types.EthAsset      `json:"ethAsset"`
	Status         Status              `json:"status" validate:"required"`
	// LastStatusUpdateTime is the time at which the status was last updated.
	LastStatusUpdateTime time.Time `json:"lastStatusUpdateTime" validate:"required"`
	// MoneroStartHeight is the Monero block number when the swap begins.
	MoneroStartHeight uint64 `json:"moneroStartHeight" validate:"required"`
	// StartTime is the time at which the swap is initiated via
	// key exchange.
	// This may vary slightly between the maker/taker.
	StartTime time.Time `json:"startTime" validate:"required"`
	// EndTime is the time at which the swap completes; ie.
	// when the node has claimed or refunded its funds.
	EndTime *time.Time `json:"endTime,omitempty"`
	// Timeout0 is the first swap timeout; before this timeout,
	// the ETH-maker is able to refund the ETH (if `ready` has not
	// been set to true in the contract). After this timeout,
	// the ETH-taker is able to claim, and the ETH-maker can
	// no longer refund.
	Timeout0 *time.Time `json:"timeout0,omitempty"`
	// Timeout1 is the second swap timeout; before this timeout
	// (and after Timeout0), the ETH-taker is able to claim, but
	// after this timeout, the ETH-taker can no longer claim, only
	// the ETH-maker can refund.
	Timeout1 *time.Time        `json:"timeout1,omitempty"`
	statusCh chan types.Status `json:"-"`
}

// NewInfo creates a new *Info from the given parameters.
// Note that the swap ID is the same as the offer ID.
func NewInfo(
	peerID peer.ID,
	offerID types.Hash,
	provides coins.ProvidesCoin,
	providedAmount, expectedAmount *apd.Decimal,
	exchangeRate *coins.ExchangeRate,
	ethAsset types.EthAsset,
	status Status,
	moneroStartHeight uint64,
	statusCh chan types.Status,
) *Info {
	info := &Info{
		Version:              CurInfoVersion,
		PeerID:               peerID,
		OfferID:              offerID,
		Provides:             provides,
		ProvidedAmount:       providedAmount,
		ExpectedAmount:       expectedAmount,
		ExchangeRate:         exchangeRate,
		EthAsset:             ethAsset,
		Status:               status,
		LastStatusUpdateTime: time.Now(),
		MoneroStartHeight:    moneroStartHeight,
		statusCh:             statusCh,
		StartTime:            time.Now(),
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
	i.LastStatusUpdateTime = time.Now()
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

	info := new(Info)
	if err := vjson.UnmarshalStruct(jsonData, info); err != nil {
		return nil, err
	}

	// TODO: Are there additional sanity checks we can perform on the Provided and Received amounts
	//       (or other fields) here when decoding the JSON?
	return info, nil
}
