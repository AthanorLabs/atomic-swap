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
	OfferID types.Hash `json:"offerID"`
}

// GetPastResponse ...
type GetPastResponse struct {
	Swaps []*PastSwap `json:"swaps" validate:"required"`
}

// GetPast returns information about a past swap given its ID.
// If no ID is provided, all past swaps are returned.
// It sorts them in order from oldest to newest.
func (s *SwapService) GetPast(_ *http.Request, req *GetPastRequest, resp *GetPastResponse) error {
	var swaps []*swap.Info

	if types.IsHashZero(req.OfferID) {
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
		info, err := s.sm.GetPastSwap(req.OfferID)
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
	ID             types.Hash          `json:"id" validate:"required"`
	Provided       coins.ProvidesCoin  `json:"provided" validate:"required"`
	ProvidedAmount *apd.Decimal        `json:"providedAmount" validate:"required"`
	ExpectedAmount *apd.Decimal        `json:"expectedAmount" validate:"required"`
	ExchangeRate   *coins.ExchangeRate `json:"exchangeRate" validate:"required"`
	Status         types.Status        `json:"status" validate:"required"`
	StartTime      time.Time           `json:"startTime" validate:"required"`
	Timeout0       *time.Time          `json:"timeout0"`
	Timeout1       *time.Time          `json:"timeout1"`
}

// GetOngoingResponse ...
type GetOngoingResponse struct {
	Swaps []*OngoingSwap `json:"swaps" validate:"dive,required"`
}

// GetOngoingRequest ...
type GetOngoingRequest struct {
	OfferID types.Hash `json:"offerID" validate:"required"`
}

// GetOngoing returns information about the ongoing swap with the given ID, if there is one.
func (s *SwapService) GetOngoing(_ *http.Request, req *GetOngoingRequest, resp *GetOngoingResponse) error {
	var (
		swaps []*swap.Info
		err   error
	)

	if types.IsHashZero(req.OfferID) {
		swaps, err = s.sm.GetOngoingSwaps()
		if err != nil {
			return err
		}
	} else {
		info, err := s.sm.GetOngoingSwap(req.OfferID)
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
		swap.StartTime = info.StartTime
		swap.Timeout0 = info.Timeout0
		swap.Timeout1 = info.Timeout1
		resp.Swaps[i] = swap
	}

	sort.Slice(resp.Swaps, func(i, j int) bool {
		return resp.Swaps[i].StartTime.UnixNano() < resp.Swaps[j].StartTime.UnixNano()
	})

	return nil
}

// RefundRequest ...
type RefundRequest struct {
	OfferID types.Hash `json:"offerID" validate:"required"`
}

// RefundResponse ...
type RefundResponse struct {
	TxHash string `json:"transactionHash" validate:"required"`
}

// Refund refunds the ongoing swap if we are the ETH provider.
// TODO: remove in favour of swap_cancel?
func (s *SwapService) Refund(_ *http.Request, req *RefundRequest, resp *RefundResponse) error {
	info, err := s.sm.GetOngoingSwap(req.OfferID)
	if err != nil {
		return err
	}

	if info.Provides != coins.ProvidesETH {
		return errCannotRefund
	}

	txHash, err := s.xmrtaker.Refund(req.OfferID)
	if err != nil {
		return fmt.Errorf("failed to refund: %w", err)
	}

	resp.TxHash = txHash.String()
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
