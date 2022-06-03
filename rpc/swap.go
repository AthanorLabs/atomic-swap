package rpc

import (
	"fmt"
	"net/http"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
)

// SwapService handles information about ongoing or past swaps.
type SwapService struct {
	sm    SwapManager
	alice Alice
	bob   Bob
	net   Net
}

// NewSwapService ...
func NewSwapService(sm SwapManager, alice Alice, bob Bob, net Net) *SwapService {
	return &SwapService{
		sm:    sm,
		alice: alice,
		bob:   bob,
		net:   net,
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
	Provided       types.ProvidesCoin `json:"provided"`
	ProvidedAmount float64            `json:"providedAmount"`
	ReceivedAmount float64            `json:"receivedAmount"`
	ExchangeRate   types.ExchangeRate `json:"exchangeRate"`
	Status         string             `json:"status"`
}

// GetPast returns information about a past swap, given its ID.
func (s *SwapService) GetPast(_ *http.Request, req *GetPastRequest, resp *GetPastResponse) error {
	info := s.sm.GetPastSwap(req.ID)
	if info == nil {
		return errNoSwapWithID
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
	ID             uint64             `json:"id"`
	Provided       types.ProvidesCoin `json:"provided"`
	ProvidedAmount float64            `json:"providedAmount"`
	ReceivedAmount float64            `json:"receivedAmount"`
	ExchangeRate   types.ExchangeRate `json:"exchangeRate"`
	Status         string             `json:"status"`
}

// GetOngoing returns information about the ongoing swap, if there is one.
func (s *SwapService) GetOngoing(_ *http.Request, _ *interface{}, resp *GetOngoingResponse) error {
	info := s.sm.GetOngoingSwap()
	if info == nil {
		return errNoOngoingSwap
	}

	resp.ID = info.ID()
	resp.Provided = info.Provides()
	resp.ProvidedAmount = info.ProvidedAmount()
	resp.ReceivedAmount = info.ReceivedAmount()
	resp.ExchangeRate = info.ExchangeRate()
	resp.Status = info.Status().String()
	return nil
}

// RefundResponse ...
type RefundResponse struct {
	TxHash string `json:"transactionHash"`
}

// Refund refunds the ongoing swap if we are the ETH provider.
// TODO: remove in favour of swap_cancel?
func (s *SwapService) Refund(_ *http.Request, _ *interface{}, resp *RefundResponse) error {
	info := s.sm.GetOngoingSwap()
	if info == nil {
		return errNoOngoingSwap
	}

	if info.Provides() != types.ProvidesETH {
		return errCannotRefund
	}

	txHash, err := s.alice.Refund()
	if err != nil {
		return fmt.Errorf("failed to refund: %w", err)
	}

	resp.TxHash = txHash.String()
	return nil
}

// GetStageResponse ...
type GetStageResponse struct {
	Stage string `json:"stage"`
	Info  string `json:"info"`
}

// GetStage returns the stage of the ongoing swap, if there is one.
func (s *SwapService) GetStage(_ *http.Request, _ *interface{}, resp *GetStageResponse) error {
	info := s.sm.GetOngoingSwap()
	if info == nil {
		return errNoOngoingSwap
	}

	resp.Stage = info.Status().String()
	resp.Info = info.Status().Info()
	return nil
}

// GetOffersResponse ...
type GetOffersResponse struct {
	Offers []*types.Offer `json:"offers"`
}

// GetOffers returns the currently available offers.
func (s *SwapService) GetOffers(_ *http.Request, _ *interface{}, resp *GetOffersResponse) error {
	resp.Offers = s.bob.GetOffers()
	return nil
}

// CancelResponse ...
type CancelResponse struct {
	Status types.Status `json:"status"`
}

// Cancel attempts to cancel the currently ongoing swap, if there is one.
func (s *SwapService) Cancel(_ *http.Request, _ *interface{}, resp *CancelResponse) error {
	info := s.sm.GetOngoingSwap()
	if info == nil {
		return errNoOngoingSwap
	}

	var ss common.SwapState
	switch info.Provides() {
	case types.ProvidesETH:
		ss = s.alice.GetOngoingSwapState()
	case types.ProvidesXMR:
		ss = s.bob.GetOngoingSwapState()
	}

	if err := ss.Exit(); err != nil {
		return err
	}
	s.net.CloseProtocolStream()

	info = s.sm.GetPastSwap(info.ID())
	resp.Status = info.Status()
	return nil
}
