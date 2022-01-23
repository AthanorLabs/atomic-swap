package rpc

import (
	"errors"
	"net/http"

	"github.com/noot/atomic-swap/common"
)

// SwapService handles information about ongoing or past swaps.
type SwapService struct {
	sm SwapManager
}

// NewSwapService ...
func NewSwapService(sm SwapManager) *SwapService {
	return &SwapService{
		sm: sm,
	}
}

// GetPastIDsResponse ...
type GetPastIDsResponse struct {
	IDs []uint64 `json:"ids"`
}

// GetPastIDs returns all past swap IDs
func (s *SwapService) GetPastIDs(_ *http.Request, _ *interface{}, resp *GetPastIDsResponse) error {
	resp.IDs = s.sm.GetPastIDs()
	return nil
}

// GetPastRequest ...
type GetPastRequest struct {
	ID uint64 `json:"id"`
}

// GetPastResponse ...
type GetPastResponse struct {
	Provided       common.ProvidesCoin `json:"provided"`
	ProvidedAmount float64             `json:"providedAmount"`
	ReceivedAmount float64             `json:"receivedAmount"`
	ExchangeRate   common.ExchangeRate `json:"exchangeRate"`
	Status         string              `json:"status"`
}

// GetPast returns information about a past swap, given its ID.
func (s *SwapService) GetPast(_ *http.Request, req *GetPastRequest, resp *GetPastResponse) error {
	info := s.sm.GetPastSwap(req.ID)
	if info == nil {
		return errors.New("unable to find swap with given ID")
	}

	resp.Provided = info.Provides()
	resp.ProvidedAmount = info.ProvidedAmount()
	resp.ReceivedAmount = info.ReceivedAmount()
	resp.ExchangeRate = info.ExchangeRate()
	resp.Status = info.Status().String()
	return nil
}

// GetOngoingResponse ...
type GetOngoingResponse struct {
	ID             uint64              `json:"id"`
	Provided       common.ProvidesCoin `json:"provided"`
	ProvidedAmount float64             `json:"providedAmount"`
	ReceivedAmount float64             `json:"receivedAmount"`
	ExchangeRate   common.ExchangeRate `json:"exchangeRate"`
	Status         string              `json:"status"`
}

// GetOngoing returns information about the ongoing swap, if there is one.
func (s *SwapService) GetOngoing(_ *http.Request, _ *interface{}, resp *GetOngoingResponse) error {
	info := s.sm.GetOngoingSwap()
	if info == nil {
		return errors.New("no current ongoing swap")
	}

	resp.ID = info.ID()
	resp.Provided = info.Provides()
	resp.ProvidedAmount = info.ProvidedAmount()
	resp.ReceivedAmount = info.ReceivedAmount()
	resp.ExchangeRate = info.ExchangeRate()
	resp.Status = info.Status().String()
	return nil
}
