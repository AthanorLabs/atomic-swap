package rpc

import (
	"fmt"
	"net/http"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
)

// SwapService handles information about ongoing or past swaps.
type SwapService struct {
	sm       SwapManager
	xmrtaker XMRTaker
	xmrmaker XMRMaker
	net      Net
}

// NewSwapService ...
func NewSwapService(sm SwapManager, xmrtaker XMRTaker, xmrmaker XMRMaker, net Net) *SwapService {
	return &SwapService{
		sm:       sm,
		xmrtaker: xmrtaker,
		xmrmaker: xmrmaker,
		net:      net,
	}
}

// GetPastIDsResponse ...
type GetPastIDsResponse struct {
	IDs []string `json:"ids"`
}

// GetPastIDs returns all past swap IDs
func (s *SwapService) GetPastIDs(_ *http.Request, _ *interface{}, resp *GetPastIDsResponse) error {
	ids, err := s.sm.GetPastIDs()
	if err != nil {
		return err
	}

	resp.IDs = make([]string, len(ids))
	for i := range resp.IDs {
		resp.IDs[i] = ids[i].String()
	}
	return nil
}

// GetPastRequest ...
type GetPastRequest struct {
	OfferID string `json:"offerID"`
}

// GetPastResponse ...
type GetPastResponse struct {
	Provided       types.ProvidesCoin `json:"provided"`
	ProvidedAmount float64            `json:"providedAmount"`
	ReceivedAmount float64            `json:"receivedAmount"`
	ExchangeRate   types.ExchangeRate `json:"exchangeRate"`
	Status         string             `json:"status"`
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
	resp.ReceivedAmount = info.ReceivedAmount
	resp.ExchangeRate = info.ExchangeRate
	resp.Status = info.Status.String()
	return nil
}

// GetOngoingResponse ...
type GetOngoingResponse struct {
	Provided       types.ProvidesCoin `json:"provided"`
	ProvidedAmount float64            `json:"providedAmount"`
	ReceivedAmount float64            `json:"receivedAmount"`
	ExchangeRate   types.ExchangeRate `json:"exchangeRate"`
	Status         string             `json:"status"`
}

// GetOngoingRequest ...
type GetOngoingRequest struct {
	OfferID string `json:"id"`
}

// GetOngoing returns information about the ongoing swap, if there is one.
func (s *SwapService) GetOngoing(_ *http.Request, req *GetOngoingRequest, resp *GetOngoingResponse) error {
	offerID, err := offerIDStringToHash(req.OfferID)
	if err != nil {
		return err
	}

	info, err := s.sm.GetOngoingSwap(offerID)
	if err != nil {
		return err
	}

	resp.Provided = info.Provides
	resp.ProvidedAmount = info.ProvidedAmount
	resp.ReceivedAmount = info.ReceivedAmount
	resp.ExchangeRate = info.ExchangeRate
	resp.Status = info.Status.String()
	return nil
}

// RefundRequest ...
type RefundRequest struct {
	OfferID string `json:"id"`
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

	if info.Provides != types.ProvidesETH {
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
	OfferID string `json:"id"`
}

// GetStageResponse ...
type GetStageResponse struct {
	Stage string `json:"stage"`
	Info  string `json:"info"`
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
	return nil
}

// GetOffersResponse ...
type GetOffersResponse struct {
	Offers []*types.Offer `json:"offers"`
}

// GetOffers returns the currently available offers.
func (s *SwapService) GetOffers(_ *http.Request, _ *interface{}, resp *GetOffersResponse) error {
	resp.Offers = s.xmrmaker.GetOffers()
	return nil
}

// ClearOffersRequest ...
type ClearOffersRequest struct {
	IDs []string `json:"ids"`
}

// ClearOffers clears the provided offers. If there are no offers provided, it clears all offers.
func (s *SwapService) ClearOffers(_ *http.Request, req *ClearOffersRequest, _ *interface{}) error {
	err := s.xmrmaker.ClearOffers(req.IDs)
	if err != nil {
		return err
	}

	s.net.Advertise()
	return nil
}

// CancelRequest ...
type CancelRequest struct {
	OfferID string `json:"id"`
}

// CancelResponse ...
type CancelResponse struct {
	Status types.Status `json:"status"`
}

// Cancel attempts to cancel the currently ongoing swap, if there is one.
func (s *SwapService) Cancel(_ *http.Request, req *CancelRequest, resp *CancelResponse) error {
	offerID, err := offerIDStringToHash(req.OfferID)
	if err != nil {
		return err
	}

	info, err := s.sm.GetOngoingSwap(offerID)
	if err != nil {
		return err
	}

	var ss common.SwapState
	switch info.Provides {
	case types.ProvidesETH:
		ss = s.xmrtaker.GetOngoingSwapState(offerID)
	case types.ProvidesXMR:
		ss = s.xmrmaker.GetOngoingSwapState(offerID)
	}

	if err = ss.Exit(); err != nil {
		return err
	}

	s.net.CloseProtocolStream(offerID)

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
