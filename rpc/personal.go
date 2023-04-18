// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package rpc

import (
	"context"
	"fmt"
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
	Timeout uint64 `json:"timeout" validate:"required"` // timeout in seconds
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
	GasPrice uint64 `json:"gasPrice" validate:"required"`
}

// SetGasPrice sets the gas price (in Wei) to be used for ethereum transactions.
func (s *PersonalService) SetGasPrice(_ *http.Request, req *SetGasPriceRequest, _ *interface{}) error {
	s.pb.ETHClient().SetGasPrice(req.GasPrice)
	return nil
}

// TokenInfo looks up the ERC20 token's metadata
func (s *PersonalService) TokenInfo(
	_ *http.Request,
	req *rpctypes.TokenInfoRequest,
	resp *rpctypes.TokenInfoResponse,
) error {
	tokenInfo, err := s.pb.ETHClient().ERC20Info(s.ctx, req.TokenAddr)
	if err != nil {
		return err
	}

	*resp = *tokenInfo
	return nil
}

// Balances returns combined information of both the Monero and Ethereum account addresses
// and balances.
func (s *PersonalService) Balances(
	_ *http.Request,
	req *rpctypes.BalancesRequest, // optional, can be nil
	resp *rpctypes.BalancesResponse,
) error {
	mAddr, mBal, err := s.xmrmaker.GetMoneroBalance()
	if err != nil {
		return err
	}

	eBal, err := s.pb.ETHClient().Balance(s.ctx)
	if err != nil {
		return err
	}

	var tokenBalances []*coins.ERC20TokenAmount
	if req != nil {
		ec := s.pb.ETHClient()
		for _, tokenAddr := range req.TokenAddrs {
			balance, err := ec.ERC20Balance(s.ctx, tokenAddr)
			if err != nil {
				return fmt.Errorf("unable to get balance for %s: %w", tokenAddr, err)
			}

			tokenBalances = append(tokenBalances, balance)
		}
	}

	*resp = rpctypes.BalancesResponse{
		MoneroAddress:           mAddr,
		PiconeroBalance:         coins.NewPiconeroAmount(mBal.Balance),
		PiconeroUnlockedBalance: coins.NewPiconeroAmount(mBal.UnlockedBalance),
		BlocksToUnlock:          mBal.BlocksToUnlock,
		EthAddress:              s.pb.ETHClient().Address(),
		WeiBalance:              eBal,
		TokenBalances:           tokenBalances,
	}
	return nil
}
