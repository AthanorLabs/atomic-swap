// Copyright 2023 The AthanorLabs/atomic-swap Authors
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
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/pricefeed"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
)

// SwapService handles information about ongoing or past swaps.
type SwapService struct {
	ctx      context.Context
	sm       swap.Manager
	xmrtaker XMRTaker
	xmrmaker XMRMaker
	net      Net
	backend  ProtocolBackend
	rdb      RecoveryDB
}

// NewSwapService ...
func NewSwapService(
	ctx context.Context,
	sm swap.Manager,
	xmrtaker XMRTaker,
	xmrmaker XMRMaker,
	net Net,
	b ProtocolBackend,
	rdb RecoveryDB,
) *SwapService {
	return &SwapService{
		ctx:      ctx,
		sm:       sm,
		xmrtaker: xmrtaker,
		xmrmaker: xmrmaker,
		net:      net,
		backend:  b,
		rdb:      rdb,
	}
}

// PastSwap represents a past swap returned by swap_getPast.
type PastSwap struct {
	ID             types.Hash          `json:"id" validate:"required"`
	Provided       coins.ProvidesCoin  `json:"provided" validate:"required"`
	EthAsset       types.EthAsset      `json:"ethAsset"`
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
// It sorts them in order from newest to oldest.
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
			ID:             info.OfferID,
			Provided:       info.Provides,
			EthAsset:       info.EthAsset,
			ProvidedAmount: info.ProvidedAmount,
			ExpectedAmount: info.ExpectedAmount,
			ExchangeRate:   info.ExchangeRate,
			Status:         info.Status,
			StartTime:      info.StartTime,
			EndTime:        info.EndTime,
		}
	}

	sort.Slice(resp.Swaps, func(i, j int) bool {
		return resp.Swaps[j].StartTime.Before(resp.Swaps[i].StartTime)
	})

	return nil
}

// OngoingSwap represents an ongoing swap returned by swap_getOngoing.
type OngoingSwap struct {
	ID                        types.Hash          `json:"id" validate:"required"`
	Provided                  coins.ProvidesCoin  `json:"provided" validate:"required"`
	EthAsset                  types.EthAsset      `json:"ethAsset"`
	ProvidedAmount            *apd.Decimal        `json:"providedAmount" validate:"required"`
	ExpectedAmount            *apd.Decimal        `json:"expectedAmount" validate:"required"`
	ExchangeRate              *coins.ExchangeRate `json:"exchangeRate" validate:"required"`
	Status                    types.Status        `json:"status" validate:"required"`
	LastStatusUpdateTime      time.Time           `json:"lastStatusUpdateTime" validate:"required"`
	StartTime                 time.Time           `json:"startTime" validate:"required"`
	Timeout1                  *time.Time          `json:"timeout1"`
	Timeout2                  *time.Time          `json:"timeout2"`
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

// GetOngoing returns information about an ongoing swap given its ID.
// If no ID is provided, all ongoing swaps are returned.
// It sorts them in order from newest to oldest.
func (s *SwapService) GetOngoing(_ *http.Request, req *GetOngoingRequest, resp *GetOngoingResponse) error {
	env := s.backend.Env()

	var (
		swaps []*swap.Info
		err   error
	)

	if req.OfferID == nil {
		swaps, err = s.sm.GetOngoingSwapsSnapshot()
		if err != nil {
			return err
		}
	} else {
		info, err := s.sm.GetOngoingSwapSnapshot(*req.OfferID) //nolint:govet
		if err != nil {
			return err
		}

		swaps = []*swap.Info{info}
	}

	resp.Swaps = make([]*OngoingSwap, len(swaps))
	for i, info := range swaps {
		swap := new(OngoingSwap)
		swap.ID = info.OfferID
		swap.Provided = info.Provides
		swap.EthAsset = info.EthAsset
		swap.ProvidedAmount = info.ProvidedAmount
		swap.ExpectedAmount = info.ExpectedAmount
		swap.ExchangeRate = info.ExchangeRate
		swap.Status = info.Status
		swap.LastStatusUpdateTime = info.LastStatusUpdateTime
		swap.StartTime = info.StartTime
		swap.Timeout1 = info.Timeout1
		swap.Timeout2 = info.Timeout2
		swap.EstimatedTimeToCompletion, err = estimatedTimeToCompletion(env, info.Status, info.LastStatusUpdateTime)
		if err != nil {
			return fmt.Errorf("failed to estimate time to completion for swap %s: %w", info.OfferID, err)
		}

		resp.Swaps[i] = swap
	}

	sort.Slice(resp.Swaps, func(i, j int) bool {
		return resp.Swaps[j].StartTime.Before(resp.Swaps[i].StartTime)
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
	info, err := s.sm.GetOngoingSwapSnapshot(req.ID)
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
	info, err := s.sm.GetOngoingSwapSnapshot(req.OfferID)
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

	// Exit() is safe to be called concurrently, as it puts an exit event
	// into the swap state's eventCh, and events are handled sequentially.
	if err = ss.Exit(); err != nil {
		return err
	}

	s.net.CloseProtocolStream(req.OfferID)

	past, err := s.sm.GetPastSwap(info.OfferID)
	if err != nil {
		return err
	}

	resp.Status = past.Status
	return nil
}

// ManualTransactionRequest is used to call swap_claim or swap_refund.
type ManualTransactionRequest struct {
	OfferID types.Hash `json:"offerID" validate:"required"`
}

// ManualTransactionResponse is returned from swap_claim or swap_refund and contains
// the transaction hash of the claim or refund transaction.
type ManualTransactionResponse struct {
	TxHash types.Hash `json:"txHash" validate:"required"`
}

// Claim calls the `claim` method on the swap contract given swap's offer ID.
// It uses the swap recovery info stored in the database to do so.
// It does not require the swap to be ongoing.
// This is meant as a fail-safe in case of some unknown swap error.
// It returns the transaction hash of the claim transaction.
func (s *SwapService) Claim(_ *http.Request, req *ManualTransactionRequest, resp *ManualTransactionResponse) error {
	contractSwapInfo, err := s.rdb.GetContractSwapInfo(req.OfferID)
	if err != nil {
		return err
	}

	secret, err := s.rdb.GetSwapPrivateKey(req.OfferID)
	if err != nil {
		return err
	}

	ec := s.backend.ETHClient()
	swapCreator, err := contracts.NewSwapCreator(contractSwapInfo.SwapCreatorAddr, ec.Raw())
	if err != nil {
		return err
	}

	ec.Lock()
	defer ec.Unlock()

	txOpts, err := ec.TxOpts(s.backend.Ctx())
	if err != nil {
		return err
	}

	tx, err := swapCreator.Claim(txOpts, *contractSwapInfo.Swap, [32]byte(common.Reverse(secret.Bytes())))
	if err != nil {
		return err
	}

	resp.TxHash = tx.Hash()
	return nil
}

// Refund calls the `refund` method on the swap contract given swap's offer ID.
// It uses the swap recovery info stored in the database to do so.
// It does not require the swap to be ongoing.
// This is meant as a fail-safe in case of some unknown swap error.
// It returns the transaction hash of the refund transaction.
func (s *SwapService) Refund(_ *http.Request, req *ManualTransactionRequest, resp *ManualTransactionResponse) error {
	contractSwapInfo, err := s.rdb.GetContractSwapInfo(req.OfferID)
	if err != nil {
		return err
	}

	secret, err := s.rdb.GetSwapPrivateKey(req.OfferID)
	if err != nil {
		return err
	}

	ec := s.backend.ETHClient()
	swapCreator, err := contracts.NewSwapCreator(contractSwapInfo.SwapCreatorAddr, ec.Raw())
	if err != nil {
		return err
	}

	ec.Lock()
	defer ec.Unlock()

	txOpts, err := ec.TxOpts(s.backend.Ctx())
	if err != nil {
		return err
	}

	tx, err := swapCreator.Refund(txOpts, *contractSwapInfo.Swap, [32]byte(common.Reverse(secret.Bytes())))
	if err != nil {
		return err
	}

	resp.TxHash = tx.Hash()
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

// estimatedTimeToCompletion returns the estimated time for the swap to complete
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
	case types.SweepingXMR:
		return (moneroBlockTime * 2), nil
	default:
		return 0, fmt.Errorf("invalid status %s; must be ongoing status type", status)
	}
}
