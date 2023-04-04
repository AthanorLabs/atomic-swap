// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package rpc

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/pricefeed"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
)

// SwapService handles information about ongoing or past swaps.
type SwapService struct {
	ctx      context.Context
	sm       SwapManager
	xmrtaker XMRTaker
	xmrmaker XMRMaker
	net      Net
	backend  ProtocolBackend
}

// NewSwapService ...
func NewSwapService(
	ctx context.Context,
	sm SwapManager,
	xmrtaker XMRTaker,
	xmrmaker XMRMaker,
	net Net,
	b ProtocolBackend,
) *SwapService {
	return &SwapService{
		ctx:      ctx,
		sm:       sm,
		xmrtaker: xmrtaker,
		xmrmaker: xmrmaker,
		net:      net,
		backend:  b,
	}
}

// PastSwap represents a past swap returned by swap_getPast.
type PastSwap struct {
	ID             types.Hash          `json:"id" validate:"required"`
	Provided       coins.ProvidesCoin  `json:"provided" validate:"required"`
	ProvidedAmount *apd.Decimal        `json:"providedAmount" validate:"required"`
	ExpectedAmount *apd.Decimal        `json:"expectedAmount" validate:"required"`
	ExchangeRate   *coins.ExchangeRate `json:"exchangeRate" validate:"required"`
	Status         types.Status        `json:"status" validate:"required"`
	StartTime      time.Time           `json:"startTime" validate:"required"`
	EndTime        *time.Time          `json:"endTime"`
}

// GetPastRequest ...
type GetPastRequest struct {
	OfferID *types.Hash `json:"offerID,omitempty"`
}

// GetPastResponse ...
type GetPastResponse struct {
	Swaps []*PastSwap `json:"swaps" validate:"dive,required"`
}

// GetPast returns information about a past swap given its ID.
// If no ID is provided, all past swaps are returned.
// It sorts them in order from oldest to newest.
func (s *SwapService) GetPast(_ *http.Request, req *GetPastRequest, resp *GetPastResponse) error {
	var swaps []*swap.Info

	if req.OfferID == nil {
		ids, err := s.sm.GetPastIDs()
		if err != nil {
			return err
		}

		for _, id := range ids {
			info, err := s.sm.GetPastSwap(id)
			if err != nil {
				return fmt.Errorf("failed to get past swap %s: %w", id, err)
			}

			swaps = append(swaps, info)
		}
	} else {
		info, err := s.sm.GetPastSwap(*req.OfferID)
		if err != nil {
			return err
		}

		swaps = append(swaps, info)
	}

	resp.Swaps = make([]*PastSwap, len(swaps))
	for i, info := range swaps {
		resp.Swaps[i] = &PastSwap{
			ID:             info.ID,
			Provided:       info.Provides,
			ProvidedAmount: info.ProvidedAmount,
			ExpectedAmount: info.ExpectedAmount,
			ExchangeRate:   info.ExchangeRate,
			Status:         info.Status,
			StartTime:      info.StartTime,
			EndTime:        info.EndTime,
		}
	}

	sort.Slice(resp.Swaps, func(i, j int) bool {
		return resp.Swaps[i].StartTime.UnixNano() < resp.Swaps[j].StartTime.UnixNano()
	})

	return nil
}

// OngoingSwap represents an ongoing swap returned by swap_getOngoing.
type OngoingSwap struct {
	ID                        types.Hash          `json:"id" validate:"required"`
	Provided                  coins.ProvidesCoin  `json:"provided" validate:"required"`
	ProvidedAmount            *apd.Decimal        `json:"providedAmount" validate:"required"`
	ExpectedAmount            *apd.Decimal        `json:"expectedAmount" validate:"required"`
	ExchangeRate              *coins.ExchangeRate `json:"exchangeRate" validate:"required"`
	Status                    types.Status        `json:"status" validate:"required"`
	LastStatusUpdateTime      time.Time           `json:"lastStatusUpdateTime" validate:"required"`
	StartTime                 time.Time           `json:"startTime" validate:"required"`
	Timeout0                  *time.Time          `json:"timeout0"`
	Timeout1                  *time.Time          `json:"timeout1"`
	EstimatedTimeToCompletion time.Duration       `json:"estimatedTimeToCompletion" validate:"required"`
}

// GetOngoingRequest ...
type GetOngoingRequest struct {
	OfferID *types.Hash `json:"offerID,omitempty"`
}

// GetOngoingResponse ...
type GetOngoingResponse struct {
	Swaps []*OngoingSwap `json:"swaps" validate:"dive,required"`
}

// GetOngoing returns information about the ongoing swap with the given ID, if there is one.
func (s *SwapService) GetOngoing(_ *http.Request, req *GetOngoingRequest, resp *GetOngoingResponse) error {
	env := s.backend.Env()

	var (
		swaps []*swap.Info
		err   error
	)

	if req.OfferID == nil {
		swaps, err = s.sm.GetOngoingSwaps()
		if err != nil {
			return err
		}
	} else {
		info, err := s.sm.GetOngoingSwap(*req.OfferID) //nolint:govet
		if err != nil {
			return err
		}

		swaps = []*swap.Info{&info}
	}

	resp.Swaps = make([]*OngoingSwap, len(swaps))
	for i, info := range swaps {
		swap := new(OngoingSwap)
		swap.ID = info.ID
		swap.Provided = info.Provides
		swap.ProvidedAmount = info.ProvidedAmount
		swap.ExpectedAmount = info.ExpectedAmount
		swap.ExchangeRate = info.ExchangeRate
		swap.Status = info.Status
		swap.LastStatusUpdateTime = info.LastStatusUpdateTime
		swap.StartTime = info.StartTime
		swap.Timeout0 = info.Timeout0
		swap.Timeout1 = info.Timeout1
		swap.EstimatedTimeToCompletion, err = estimatedTimeToCompletion(env, info.Status, info.LastStatusUpdateTime)
		if err != nil {
			return fmt.Errorf("failed to estimate time to completion for swap %s: %w", info.ID, err)
		}

		resp.Swaps[i] = swap
	}

	sort.Slice(resp.Swaps, func(i, j int) bool {
		return resp.Swaps[i].StartTime.UnixNano() < resp.Swaps[j].StartTime.UnixNano()
	})

	return nil
}

// GetStatusRequest ...
type GetStatusRequest struct {
	ID types.Hash `json:"id" validate:"required"`
}

// GetStatusResponse ...
type GetStatusResponse struct {
	Status      types.Status `json:"status" validate:"required"`
	Description string       `json:"info" validate:"required"`
	StartTime   time.Time    `json:"startTime" validate:"required"`
}

// GetStatus returns the status of the ongoing swap, if there is one.
func (s *SwapService) GetStatus(_ *http.Request, req *GetStatusRequest, resp *GetStatusResponse) error {
	info, err := s.sm.GetOngoingSwap(req.ID)
	if err != nil {
		return err
	}

	resp.Status = info.Status
	resp.Description = info.Status.Description()
	resp.StartTime = info.StartTime
	return nil
}

// GetOffersResponse ...
type GetOffersResponse struct {
	PeerID peer.ID        `json:"peerID" validate:"required"`
	Offers []*types.Offer `json:"offers" validate:"dive,required"`
}

// GetOffers returns our currently available offers.
func (s *SwapService) GetOffers(_ *http.Request, _ *interface{}, resp *GetOffersResponse) error {
	resp.PeerID = s.net.PeerID()
	resp.Offers = s.xmrmaker.GetOffers()
	return nil
}

// ClearOffersRequest ...
type ClearOffersRequest struct {
	OfferIDs []types.Hash `json:"offerIDs" validate:"dive,required"`
}

// ClearOffers clears our provided offers. If there are no offers provided, it clears all offers.
func (s *SwapService) ClearOffers(_ *http.Request, req *ClearOffersRequest, _ *interface{}) error {
	err := s.xmrmaker.ClearOffers(req.OfferIDs)
	if err != nil {
		return err
	}

	return nil
}

// CancelRequest ...
type CancelRequest struct {
	OfferID types.Hash `json:"offerID" validate:"required"`
}

// CancelResponse ...
type CancelResponse struct {
	Status types.Status `json:"status" validate:"required"`
}

// Cancel attempts to cancel the currently ongoing swap, if there is one.
func (s *SwapService) Cancel(_ *http.Request, req *CancelRequest, resp *CancelResponse) error {
	info, err := s.sm.GetOngoingSwap(req.OfferID)
	if err != nil {
		return fmt.Errorf("failed to get ongoing swap: %w", err)
	}

	var ss common.SwapState
	switch info.Provides {
	case coins.ProvidesETH:
		ss = s.xmrtaker.GetOngoingSwapState(req.OfferID)
	case coins.ProvidesXMR:
		ss = s.xmrmaker.GetOngoingSwapState(req.OfferID)
	}

	if ss == nil {
		return fmt.Errorf("failed to find swap state with ID %s", req.OfferID)
	}

	// Exit() is safe to be called concurrently, since it since it puts an exit event
	// into the swap state's eventCh, and events are handled sequentially.
	if err = ss.Exit(); err != nil {
		return err
	}

	s.net.CloseProtocolStream(req.OfferID)

	past, err := s.sm.GetPastSwap(info.ID)
	if err != nil {
		return err
	}

	resp.Status = past.Status
	return nil
}

// SuggestedExchangeRateResponse ...
type SuggestedExchangeRateResponse struct {
	ETHUpdatedAt time.Time           `json:"ethUpdatedAt" validate:"required"`
	ETHPrice     *apd.Decimal        `json:"ethPrice" validate:"required"`
	XMRUpdatedAt time.Time           `json:"xmrUpdatedAt" validate:"required"`
	XMRPrice     *apd.Decimal        `json:"xmrPrice" validate:"required"`
	ExchangeRate *coins.ExchangeRate `json:"exchangeRate" validate:"required"`
}

// SuggestedExchangeRate returns the current mainnet exchange rate, expressed as the XMR/ETH price.
func (s *SwapService) SuggestedExchangeRate(_ *http.Request, _ *interface{}, resp *SuggestedExchangeRateResponse) error { //nolint:lll
	ec := s.backend.ETHClient().Raw()

	xmrFeed, err := pricefeed.GetXMRUSDPrice(s.ctx, ec)
	if err != nil {
		return err
	}

	ethFeed, err := pricefeed.GetETHUSDPrice(s.ctx, ec)
	if err != nil {
		return err
	}

	exchangeRate, err := coins.CalcExchangeRate(xmrFeed.Price, ethFeed.Price)
	if err != nil {
		return err
	}

	resp.XMRUpdatedAt = xmrFeed.UpdatedAt
	resp.XMRPrice = xmrFeed.Price

	resp.ETHUpdatedAt = ethFeed.UpdatedAt
	resp.ETHPrice = ethFeed.Price

	resp.ExchangeRate = exchangeRate
	return nil
}

// estimatedTimeToCompletionreturns the estimated time for the swap to complete
// in the optimistic case based on the given status and the time the status was updated.
func estimatedTimeToCompletion(
	env common.Environment,
	status types.Status,
	lastStatusUpdateTime time.Time,
) (time.Duration, error) {
	if time.Until(lastStatusUpdateTime) > 0 {
		return 0, fmt.Errorf("last status update time must be less than now")
	}

	timeForStatus, err := estimatedTimeToCompletionForStatus(env, status)
	if err != nil {
		return 0, err
	}

	estimatedTime := timeForStatus - time.Since(lastStatusUpdateTime)
	if estimatedTime < 0 {
		// TODO: add explanation as to why time to completion can't be estimated,
		// probably because we need to wait for the countparty to refund, or
		// monero block times were longer than expected.
		return 0, nil
	}

	return estimatedTime.Round(time.Second), nil
}

// estimatedTimeToCompletionForStatus returns the estimated time for the swap to complete
// in the optimistic case based on the given status, assuming the status was updated just now.
func estimatedTimeToCompletionForStatus(env common.Environment, status types.Status) (time.Duration, error) {
	var (
		moneroBlockTime time.Duration
		ethBlockTime    time.Duration
	)

	switch env {
	case common.Development:
		moneroBlockTime = time.Second
		ethBlockTime = time.Second
	default:
		moneroBlockTime = time.Minute * 2
		ethBlockTime = time.Second * 12
	}

	// we assume the Monero lock step will take 10 blocks, and for the taker,
	// there is the additional 2 blocks to transfer the funds from the swap wallet
	// to the original wallet.
	//
	// we also assume all Ethereum txs will take at maximum 2 blocks
	// to be included.
	switch status {
	case types.ExpectingKeys:
		return (moneroBlockTime * 12) + (ethBlockTime * 6), nil
	case types.KeysExchanged:
		return (moneroBlockTime * 10) + (ethBlockTime * 6), nil
	case types.ETHLocked:
		return (moneroBlockTime * 12) + (ethBlockTime * 4), nil
	case types.XMRLocked:
		return (moneroBlockTime * 10) + (ethBlockTime * 4), nil
	case types.ContractReady:
		return (moneroBlockTime * 2) + (ethBlockTime * 2), nil
	default:
		return 0, fmt.Errorf("invalid status %s; must be ongoing status type", status)
	}
}
