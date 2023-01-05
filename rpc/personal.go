package rpc

import (
	"context"
	"net/http"
	"time"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
)

// PersonalService handles private keys and wallets.
type PersonalService struct {
	ctx      context.Context
	xmrmaker XMRMaker
	pb       ProtocolBackend
}

// NewPersonalService ...
func NewPersonalService(ctx context.Context, xmrmaker XMRMaker, pb ProtocolBackend) *PersonalService {
	return &PersonalService{
		ctx:      ctx,
		xmrmaker: xmrmaker,
		pb:       pb,
	}
}

// SetSwapTimeoutRequest ...
type SetSwapTimeoutRequest struct {
	Timeout uint64 `json:"timeout"` // timeout in seconds
}

// SetSwapTimeout ...
func (s *PersonalService) SetSwapTimeout(_ *http.Request, req *SetSwapTimeoutRequest, _ *interface{}) error {
	timeout := time.Second * time.Duration(req.Timeout)
	s.pb.SetSwapTimeout(timeout)
	return nil
}

// GetSwapTimeoutResponse ...
type GetSwapTimeoutResponse struct {
	Timeout uint64 `json:"timeout"` // timeout in seconds
}

// GetSwapTimeout ...
func (s *PersonalService) GetSwapTimeout(_ *http.Request, _ *interface{}, resp *GetSwapTimeoutResponse) error {
	resp.Timeout = uint64(s.pb.SwapTimeout().Seconds())
	return nil
}

// SetGasPriceRequest ...
type SetGasPriceRequest struct {
	GasPrice uint64
}

// SetGasPrice sets the gas price (in wei) to be used for ethereum transactions.
func (s *PersonalService) SetGasPrice(_ *http.Request, req *SetGasPriceRequest, _ *interface{}) error {
	s.pb.ETHClient().SetGasPrice(req.GasPrice)
	return nil
}

// Balances returns combined information of both the Monero and Ethereum account addresses
// and balances.
func (s *PersonalService) Balances(_ *http.Request, _ *interface{}, resp *rpctypes.BalancesResponse) error {
	mAddr, mBal, err := s.xmrmaker.GetMoneroBalance()
	if err != nil {
		return err
	}

	eBal, err := s.pb.ETHClient().Balance(s.ctx)
	if err != nil {
		return err
	}

	*resp = rpctypes.BalancesResponse{
		MoneroAddress:           mAddr,
		PiconeroBalance:         coins.NewPiconeroAmount(mBal.Balance),
		PiconeroUnlockedBalance: coins.NewPiconeroAmount(mBal.UnlockedBalance),
		BlocksToUnlock:          mBal.BlocksToUnlock,
		EthAddress:              s.pb.ETHClient().Address().String(),
		WeiBalance:              coins.BigInt2Wei(eBal),
	}
	return nil
}
