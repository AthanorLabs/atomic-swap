// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package swap

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"
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
	RelayerFee     *big.Int            `json:"relayerFee,omitempty"`
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
	// Timeout1 is the first swap timeout; before this timeout,
	// the ETH-maker is able to refund the ETH (if `ready` has not
	// been set to true in the contract). After this timeout,
	// the ETH-taker is able to claim, and the ETH-maker can
	// no longer refund.
	Timeout1 *time.Time `json:"timeout1,omitempty"`
	// Timeout2 is the second swap timeout; before this timeout
	// (and after Timeout1), the ETH-taker is able to claim, but
	// after this timeout, the ETH-taker can no longer claim, only
	// the ETH-maker can refund.
	Timeout2 *time.Time `json:"timeout2,omitempty"`

	// rwMu handles synchronization when LastStatusUpdateTime, Timeout1,
	// Timeout2 and EndTime are updated. This Info struct is modified by the
	// maker or taker's swapState go process as the state of the swap
	// progresses. The swapState go process does not need synchronization when
	// reading its own changes, but it needs to grab a write lock when modifying
	// the the structure. Readers from other go-processes only get copies of
	// this structure. They exclusively use the DeepCopy method to get their
	// copy, which grabs the read lock ensuring that they always capture the
	// up-to-date state of this Info struct.
	rwMu sync.RWMutex
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
		StartTime:            time.Now(),
		EndTime:              nil,
		Timeout1:             nil,
		Timeout2:             nil,
		rwMu:                 sync.RWMutex{},
	}
	return info
}

// SetStatus updates the status and status modification timestamp
func (i *Info) SetStatus(s Status) {
	i.rwMu.Lock()
	defer i.rwMu.Unlock()

	i.Status = s
	i.LastStatusUpdateTime = time.Now()
}

// SetTimeouts sets the 2 timeout fields, , grabbing the needed lock before
// modifying fields.
func (i *Info) SetTimeouts(t1 *time.Time, t2 *time.Time) {
	i.rwMu.Lock()
	defer i.rwMu.Unlock()

	i.Timeout1 = t1
	i.Timeout2 = t2
}

// MarkSwapComplete sets the EndTime field to the current wall time, grabbing the
// needed lock before modifying fields.
func (i *Info) MarkSwapComplete() {
	i.rwMu.Lock()
	defer i.rwMu.Unlock()

	now := time.Now()
	i.EndTime = &now
}

// IsTaker returns true if the node is the xmr-taker in the swap.
func (i *Info) IsTaker() bool {
	return i.Provides == coins.ProvidesETH
}

// UnmarshalInfo unmarshalls the passed JSON into a freshly created Info object.
func UnmarshalInfo(jsonData []byte) (*Info, error) {
	info := new(Info)
	if err := json.Unmarshal(jsonData, info); err != nil {
		return nil, err
	}
	return info, nil
}

// UnmarshalJSON deserializes a JSON Info struct, checking the version for
// compatibility.
func (i *Info) UnmarshalJSON(jsonData []byte) error {
	iv := struct {
		Version *semver.Version `json:"version"`
	}{}
	if err := json.Unmarshal(jsonData, &iv); err != nil {
		return err
	}

	if iv.Version == nil {
		return errInfoVersionMissing
	}

	if iv.Version.GreaterThan(CurInfoVersion) {
		return fmt.Errorf("info version %q not supported, latest is %q", iv.Version, CurInfoVersion)
	}

	// Assuming any version less than the current version is forwards
	// compatible. If that is not the case in the future, add code here to
	// upgrade the older version to the current version when deserializing.
	// (Or error if it is completely incompatible.)

	// Unmarshal without recursion
	type _Info Info
	if err := vjson.UnmarshalStruct(jsonData, (*_Info)(i)); err != nil {
		return err
	}

	// TODO: Are there additional sanity checks we can perform on the Provided and Received amounts
	//       (or other fields) here when decoding the JSON?
	return nil
}

// DeepCopy returns a deep copy of the Info data structure
func (i *Info) DeepCopy() (*Info, error) {
	i.rwMu.RLock()
	defer i.rwMu.RUnlock()

	// This is not the most efficient means of getting a deep copy, but for our
	// needs it is fast enough and least prone to human error, as the structure
	// has numerous nested pointer types.
	jsonData, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	clone := new(Info)
	if err = clone.UnmarshalJSON(jsonData); err != nil {
		return nil, err
	}

	return clone, nil
}
