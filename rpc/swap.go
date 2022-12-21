package rpc

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"net/http"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
)

// SwapService handles information about ongoing or past swaps.
type SwapService struct {
	sm       SwapManager
	xmrtaker XMRTaker
	xmrmaker XMRMaker
	net      Net
	backend  ProtocolBackend
}

// NewSwapService ...
func NewSwapService(sm SwapManager, xmrtaker XMRTaker, xmrmaker XMRMaker, net Net, b ProtocolBackend) *SwapService {
	return &SwapService{
		sm:       sm,
		xmrtaker: xmrtaker,
		xmrmaker: xmrmaker,
		net:      net,
		backend:  b,
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
	OfferID string `json:"offerID"`
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
	OfferID string `json:"offerID"`
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
	PeerID peer.ID
	Offers []*types.Offer `json:"offers"`
}

// GetOffers returns the currently available offers.
func (s *SwapService) GetOffers(_ *http.Request, _ *interface{}, resp *GetOffersResponse) error {
	resp.PeerID = s.net.PeerID()
	resp.Offers = s.xmrmaker.GetOffers()
	return nil
}

// ClearOffersRequest ...
type ClearOffersRequest struct {
	OfferIDs []types.Hash `json:"offerIDs"`
}

// ClearOffers clears the provided offers. If there are no offers provided, it clears all offers.
func (s *SwapService) ClearOffers(_ *http.Request, req *ClearOffersRequest, _ *interface{}) error {
	err := s.xmrmaker.ClearOffers(req.OfferIDs)
	if err != nil {
		return err
	}

	s.net.Advertise()
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
	case types.ProvidesETH:
		ss = s.xmrtaker.GetOngoingSwapState(req.OfferID)
	case types.ProvidesXMR:
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
	ETHPrice     float64 `json:"ethPrice"`
	XMRPrice     float64 `json:"xmrPrice"`
	ExchangeRate float64 `json:"exchangeRate"`
}

// SuggestedExchangeRate returns the current mainnet exchange rate, expressed as the XMR/ETH price.
func (s *SwapService) SuggestedExchangeRate(_ *http.Request, _ *interface{}, resp *SuggestedExchangeRateResponse) error { //nolint:lll
	decimals := math.Pow(10, 8)

	ec := s.backend.ETHClient().Raw()
	ethPrice, err := common.GetETHUSDPrice(context.Background(), ec)
	if err != nil {
		return err
	}

	xmrPrice, err := common.GetXMRUSDPrice(context.Background(), ec)
	if err != nil {
		return err
	}

	ethPriceFloat := new(big.Float).SetInt(ethPrice)
	xmrPriceFloat := new(big.Float).SetInt(xmrPrice)
	exchangeRate := new(big.Float).Quo(xmrPriceFloat, ethPriceFloat)

	resp.ETHPrice = float64(ethPrice.Uint64()) / decimals
	resp.XMRPrice = float64(xmrPrice.Uint64()) / decimals
	resp.ExchangeRate, _ = exchangeRate.Float64()
	return nil
}
