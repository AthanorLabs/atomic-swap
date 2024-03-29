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
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"

	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
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

// TransferXMRRequest ...
type TransferXMRRequest struct {
	To     *mcrypto.Address `json:"to" validate:"required"`
	Amount *apd.Decimal     `json:"amount" validate:"required"`
}

// TransferXMRResponse ...
type TransferXMRResponse struct {
	TxID string `json:"txID"`
}

// TransferXMR transfers XMR from the swapd wallet.
func (s *PersonalService) TransferXMR(_ *http.Request, req *TransferXMRRequest, resp *TransferXMRResponse) error {
	txID, err := s.pb.TransferXMR(req.To, coins.MoneroToPiconero(req.Amount))
	if err != nil {
		return err
	}

	resp.TxID = txID
	return nil
}

// SweepXMRRequest ...
type SweepXMRRequest struct {
	To *mcrypto.Address `json:"to" validate:"required"`
}

// SweepXMRResponse ...
type SweepXMRResponse struct {
	TxIDs []string `json:"txIds"`
}

// SweepXMR sweeps XMR from the swapd wallet.
func (s *PersonalService) SweepXMR(_ *http.Request, req *SweepXMRRequest, resp *SweepXMRResponse) error {
	txIDs, err := s.pb.SweepXMR(req.To)
	if err != nil {
		return err
	}

	resp.TxIDs = txIDs
	return nil
}

// TransferETHRequest is JSON-RPC request object for TransferETH
type TransferETHRequest struct {
	To       ethcommon.Address `json:"to" validate:"required"`
	Amount   *apd.Decimal      `json:"amount" validate:"required"`
	GasLimit *uint64           `json:"gasLimit,omitempty"`
}

// TransferETHResponse is JSON-RPC response object for TransferETH
type TransferETHResponse struct {
	TxHash   ethcommon.Hash `json:"txHash"`
	GasLimit *uint64        `json:"gasLimit,omitempty"`
}

// TransferETH transfers ETH from the swapd wallet.
func (s *PersonalService) TransferETH(_ *http.Request, req *TransferETHRequest, resp *TransferETHResponse) error {
	receipt, err := s.pb.TransferETH(req.To, coins.EtherToWei(req.Amount), req.GasLimit)
	if err != nil {
		return err
	}

	resp.TxHash = receipt.TxHash
	return nil
}

// SweepETHRequest is JSON-RPC request object for SweepETH
type SweepETHRequest struct {
	To ethcommon.Address `json:"to" validate:"required"`
}

// SweepETHResponse is JSON-RPC response object for SweepETH
type SweepETHResponse struct {
	TxHash ethcommon.Hash `json:"txHash"` // Hash of sweep transfer transaction
}

// SweepETH sweeps all ETH out of the swapd wallet.
func (s *PersonalService) SweepETH(_ *http.Request, req *SweepETHRequest, resp *SweepETHResponse) error {
	receipt, err := s.pb.SweepETH(req.To)
	if err != nil {
		return err
	}

	resp.TxHash = receipt.TxHash
	return nil
}
