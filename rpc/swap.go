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

// PastSwapInfo contains a swap ID and its start and end time.
type PastSwapInfo struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

// GetPastIDsResponse ...
type GetPastIDsResponse struct {
	IDs []*PastSwapInfo `json:"ids"`
}

// GetPastIDs returns all past swap IDs and their start and end times.
// It sorts them in order from oldest to newest.
func (s *SwapService) GetPastIDs(_ *http.Request, _ *interface{}, resp *GetPastIDsResponse) error {
	ids, err := s.sm.GetPastIDs()
	if err != nil {
		return err
	}

	resp.IDs = make([]*PastSwapInfo, len(ids))
	for i, id := range ids {
		info, err := s.sm.GetPastSwap(id)
		if err != nil {
			return fmt.Errorf("failed to get past swap %s: %w", id, err)
		}

		resp.IDs[i] = &PastSwapInfo{
			ID:        id.String(),
			StartTime: info.StartTime,
			EndTime:   info.EndTime,
		}
	}

	sort.Slice(resp.IDs, func(i, j int) bool {
		return resp.IDs[i].StartTime.UnixNano() > resp.IDs[j].StartTime.UnixNano()
	})

	return nil
}

// GetPastRequest ...
type GetPastRequest struct {
	OfferID string `json:"offerID"`
}

// GetPastResponse ...
type GetPastResponse struct {
	Provided       coins.ProvidesCoin  `json:"provided"`
	ProvidedAmount *apd.Decimal        `json:"providedAmount"`
	ExpectedAmount *apd.Decimal        `json:"expectedAmount"`
	ExchangeRate   *coins.ExchangeRate `json:"exchangeRate"`
	Status         string              `json:"status"`
	StartTime      time.Time           `json:"startTime"`
	EndTime        time.Time           `json:"endTime"`
}

// GetPast returns information about a past swap, given its ID.
func (s *SwapService) GetPast(_ *http.Request, req *GetPastRequest, resp *GetPastResponse) error {
	offerID, err := offerIDStringToHash(req.OfferID)
	if err != nil {
		return err
	}

	info, err := s.sm.GetPastSwap(offerID)
	if err != nil {
		return err
	}

	resp.Provided = info.Provides
	resp.ProvidedAmount = info.ProvidedAmount
	resp.ExpectedAmount = info.ExpectedAmount
	resp.ExchangeRate = info.ExchangeRate
	resp.Status = info.Status.String()
	resp.StartTime = info.StartTime
	resp.EndTime = info.EndTime
	return nil
}

// OngoingSwap represents an ongoing swap returned by swap_getOngoing.
type OngoingSwap struct {
	Provided       coins.ProvidesCoin  `json:"provided"`
	ProvidedAmount *apd.Decimal        `json:"providedAmount"`
	ExpectedAmount *apd.Decimal        `json:"expectedAmount"`
	ExchangeRate   *coins.ExchangeRate `json:"exchangeRate"`
	Status         string              `json:"status"`
	StartTime      time.Time           `json:"startTime"`
}

// GetOngoingResponse ...
type GetOngoingResponse struct {
	Swaps []*OngoingSwap `json:"swaps"`
}

// GetOngoingRequest ...
type GetOngoingRequest struct {
	OfferID string `json:"offerID"`
}

// GetOngoing returns information about the ongoing swap with the given ID, if there is one.
func (s *SwapService) GetOngoing(_ *http.Request, req *GetOngoingRequest, resp *GetOngoingResponse) error {
	var (
		swaps []*swap.Info
		err   error
	)

	if req.OfferID == "" {
		swaps, err = s.sm.GetOngoingSwaps()
		if err != nil {
			return err
		}
	} else {
		offerID, err := offerIDStringToHash(req.OfferID)
		if err != nil {
			return err
		}

		info, err := s.sm.GetOngoingSwap(offerID)
		if err != nil {
			return err
		}

		swaps = []*swap.Info{&info}
	}

	resp.Swaps = make([]*OngoingSwap, len(swaps))
	for i, info := range swaps {
		resp.Swaps[i].Provided = info.Provides
		resp.Swaps[i].ProvidedAmount = info.ProvidedAmount
		resp.Swaps[i].ExpectedAmount = info.ExpectedAmount
		resp.Swaps[i].ExchangeRate = info.ExchangeRate
		resp.Swaps[i].Status = info.Status.String()
		resp.Swaps[i].StartTime = info.StartTime
	}

	return nil
}

// RefundRequest ...
type RefundRequest struct {
	OfferID string `json:"offerID"`
}

// RefundResponse ...
type RefundResponse struct {
	TxHash string `json:"transactionHash"`
}

// Refund refunds the ongoing swap if we are the ETH provider.
// TODO: remove in favour of swap_cancel?
func (s *SwapService) Refund(_ *http.Request, req *RefundRequest, resp *RefundResponse) error {
	offerID, err := offerIDStringToHash(req.OfferID)
	if err != nil {
		return err
	}

	info, err := s.sm.GetOngoingSwap(offerID)
	if err != nil {
		return err
	}

	if info.Provides != coins.ProvidesETH {
		return errCannotRefund
	}

	txHash, err := s.xmrtaker.Refund(offerID)
	if err != nil {
		return fmt.Errorf("failed to refund: %w", err)
	}

	resp.TxHash = txHash.String()
	return nil
}

// GetStageRequest ...
type GetStageRequest struct {
	OfferID string `json:"offerID"`
}

// GetStageResponse ...
type GetStageResponse struct {
	Stage     string    `json:"stage"`
	Info      string    `json:"info"`
	StartTime time.Time `json:"startTime"`
}

// GetStage returns the stage of the ongoing swap, if there is one.
func (s *SwapService) GetStage(_ *http.Request, req *GetStageRequest, resp *GetStageResponse) error {
	offerID, err := offerIDStringToHash(req.OfferID)
	if err != nil {
		return err
	}

	info, err := s.sm.GetOngoingSwap(offerID)
	if err != nil {
		return err
	}

	resp.Stage = info.Status.String()
	resp.Info = info.Status.Info()
	resp.StartTime = info.StartTime
	return nil
}

// GetOffersResponse ...
type GetOffersResponse struct {
	PeerID peer.ID        `json:"peerID"`
	Offers []*types.Offer `json:"offers"`
}

// GetOffers returns our currently available offers.
func (s *SwapService) GetOffers(_ *http.Request, _ *interface{}, resp *GetOffersResponse) error {
	resp.PeerID = s.net.PeerID()
	resp.Offers = s.xmrmaker.GetOffers()
	return nil
}

// ClearOffersRequest ...
type ClearOffersRequest struct {
	OfferIDs []types.Hash `json:"offerIDs"`
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
	OfferID types.Hash `json:"offerID"`
}

// CancelResponse ...
type CancelResponse struct {
	Status types.Status `json:"status"`
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

func offerIDStringToHash(s string) (types.Hash, error) {
	return types.HexToHash(s)
}

// SuggestedExchangeRateResponse ...
type SuggestedExchangeRateResponse struct {
	ETHUpdatedAt time.Time           `json:"ethUpdatedAt"`
	ETHPrice     *apd.Decimal        `json:"ethPrice"`
	XMRUpdatedAt time.Time           `json:"xmrUpdatedAt"`
	XMRPrice     *apd.Decimal        `json:"xmrPrice"`
	ExchangeRate *coins.ExchangeRate `json:"exchangeRate"`
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
