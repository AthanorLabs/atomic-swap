// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package rpc

import (
	"math/big"
	"net/http"

	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/db"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

// RecoveryDB contains methods for retrieving swap recovery info from the database.
type RecoveryDB interface {
	GetContractSwapInfo(id types.Hash) (*db.EthereumSwapInfo, error)
	GetSwapPrivateKey(id types.Hash) (*mcrypto.PrivateSpendKey, error)
	GetCounterpartySwapPrivateKey(id types.Hash) (*mcrypto.PrivateSpendKey, error)
}

// DatabaseService ...
type DatabaseService struct {
	rdb RecoveryDB
}

// NewDatabaseService returns a new DatabaseService.
func NewDatabaseService(rdb RecoveryDB) *DatabaseService {
	return &DatabaseService{
		rdb: rdb,
	}
}

// GetContractSwapInfoRequest ...
type GetContractSwapInfoRequest struct {
	OfferID types.Hash `json:"offerID" validate:"required"`
}

// GetContractSwapInfoResponse ...
type GetContractSwapInfoResponse struct {
	StartNumber     *big.Int                   `json:"startBlockNumber" validate:"required"`
	SwapID          types.Hash                 `json:"swapID" validate:"required"`
	Swap            *contracts.SwapCreatorSwap `json:"swap" validate:"required"`
	SwapCreatorAddr ethcommon.Address          `json:"swapCreatorAddr" validate:"required"`
}

// GetContractSwapInfo returns the contract swap info for the given swap ID from the database.
func (s *DatabaseService) GetContractSwapInfo(
	_ *http.Request,
	req *GetContractSwapInfoRequest,
	resp *GetContractSwapInfoResponse,
) error {
	info, err := s.rdb.GetContractSwapInfo(req.OfferID)
	if err != nil {
		return err
	}

	resp.StartNumber = info.StartNumber
	resp.SwapID = info.SwapID
	resp.Swap = info.Swap
	resp.SwapCreatorAddr = info.SwapCreatorAddr
	return nil
}

// GetSwapSecretRequest ...
type GetSwapSecretRequest struct {
	OfferID types.Hash `json:"offerID" validate:"required"`
}

// GetSwapSecretResponse ...
type GetSwapSecretResponse struct {
	Secret *mcrypto.PrivateSpendKey `json:"secret" validate:"required"`
}

// GetSwapSecret returns our swap secret for the given swap ID from the database.
func (s *DatabaseService) GetSwapSecret(
	_ *http.Request,
	req *GetSwapSecretRequest,
	resp *GetSwapSecretResponse,
) error {
	key, err := s.rdb.GetSwapPrivateKey(req.OfferID)
	if err != nil {
		return err
	}

	resp.Secret = key
	return nil
}
