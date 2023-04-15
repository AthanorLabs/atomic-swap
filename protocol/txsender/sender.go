// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

// Package txsender provides a common Sender interface for swapd instances. Each Sender
// implementation is responsible for signing and submitting transactions to the network.
// privateKeySender is the implementation using an ethereum private key directly managed
// by swapd. ExternalSender provides an API for interacting with an external entity like
// Metamask.
package txsender

import (
	"context"
	"fmt"
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	logging "github.com/ipfs/go-log"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
)

var (
	log = logging.Logger("txsender")
)

// Sender signs and submits transactions to the chain
type Sender interface {
	SetContract(*contracts.SwapCreator)
	SetContractAddress(ethcommon.Address)
	NewSwap(
		pubKeyClaim [32]byte,
		pubKeyRefund [32]byte,
		claimer ethcommon.Address,
		timeoutDuration *big.Int,
		nonce *big.Int,
		ethAsset types.EthAsset,
		amount *big.Int,
	) (*ethtypes.Receipt, error)
	SetReady(swap *contracts.SwapCreatorSwap) (*ethtypes.Receipt, error)
	Claim(swap *contracts.SwapCreatorSwap, secret [32]byte) (*ethtypes.Receipt, error)
	Refund(swap *contracts.SwapCreatorSwap, secret [32]byte) (*ethtypes.Receipt, error)
}

type privateKeySender struct {
	ctx           context.Context
	ethClient     extethclient.EthClient
	swapContract  *contracts.SwapCreator
	erc20Contract *contracts.IERC20
}

// NewSenderWithPrivateKey returns a new *privateKeySender
func NewSenderWithPrivateKey(
	ctx context.Context,
	ethClient extethclient.EthClient,
	swapContract *contracts.SwapCreator,
	erc20Contract *contracts.IERC20,
) Sender {
	return &privateKeySender{
		ctx:           ctx,
		ethClient:     ethClient,
		swapContract:  swapContract,
		erc20Contract: erc20Contract,
	}
}

func (s *privateKeySender) SetContract(contract *contracts.SwapCreator) {
	s.swapContract = contract
}

func (s *privateKeySender) SetContractAddress(_ ethcommon.Address) {}

func (s *privateKeySender) NewSwap(
	pubKeyClaim [32]byte,
	pubKeyRefund [32]byte,
	claimer ethcommon.Address,
	timeoutDuration *big.Int,
	nonce *big.Int,
	ethAsset types.EthAsset,
	value *big.Int,
) (*ethtypes.Receipt, error) {
	s.ethClient.Lock()
	defer s.ethClient.Unlock()

	// For token swaps, approving our contract to transfer tokens, and calling
	// NewSwap which performs the transfer, needs to be inside the same wallet
	// lock grab in case there are other simultaneous swaps happening with the
	// same token.
	if ethAsset.IsToken() {
		txOpts, err := s.ethClient.TxOpts(s.ctx)
		if err != nil {
			return nil, err
		}

		tx, err := s.erc20Contract.Approve(txOpts, s.ethClient.Address(), value)
		if err != nil {
			return nil, fmt.Errorf("approve tx creation failed, %w", err)
		}

		receipt, err := block.WaitForReceipt(s.ctx, s.ethClient.Raw(), tx.Hash())
		if err != nil {
			return nil, fmt.Errorf("approve failed, %w", err)
		}

		log.Infof("approve transaction included %s", common.ReceiptInfo(receipt))
	}

	txOpts, err := s.ethClient.TxOpts(s.ctx)
	if err != nil {
		return nil, err
	}

	// transfer ETH if we're not doing an ERC20 swap
	if ethAsset.IsETH() {
		txOpts.Value = value
	}

	tx, err := s.swapContract.NewSwap(txOpts, pubKeyClaim, pubKeyRefund, claimer, timeoutDuration, timeoutDuration,
		ethcommon.Address(ethAsset), value, nonce)
	if err != nil {
		err = fmt.Errorf("new_swap tx creation failed, %w", err)
		return nil, err
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ethClient.Raw(), tx.Hash())
	if err != nil {
		err = fmt.Errorf("new_swap failed, %w", err)
		return nil, err
	}

	return receipt, nil
}

func (s *privateKeySender) SetReady(swap *contracts.SwapCreatorSwap) (*ethtypes.Receipt, error) {
	s.ethClient.Lock()
	defer s.ethClient.Unlock()
	txOpts, err := s.ethClient.TxOpts(s.ctx)
	if err != nil {
		return nil, err
	}

	tx, err := s.swapContract.SetReady(txOpts, *swap)
	if err != nil {
		err = fmt.Errorf("set_ready tx creation failed, %w", err)
		return nil, err
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ethClient.Raw(), tx.Hash())
	if err != nil {
		err = fmt.Errorf("set_ready failed, %w", err)
		return nil, err
	}

	return receipt, nil
}

func (s *privateKeySender) Claim(
	swap *contracts.SwapCreatorSwap,
	secret [32]byte,
) (*ethtypes.Receipt, error) {
	s.ethClient.Lock()
	defer s.ethClient.Unlock()
	txOpts, err := s.ethClient.TxOpts(s.ctx)
	if err != nil {
		return nil, err
	}

	tx, err := s.swapContract.Claim(txOpts, *swap, secret)
	if err != nil {
		err = fmt.Errorf("claim tx creation failed, %w", err)
		return nil, err
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ethClient.Raw(), tx.Hash())
	if err != nil {
		err = fmt.Errorf("claim failed, %w", err)
		return nil, err
	}

	return receipt, nil
}

func (s *privateKeySender) Refund(
	swap *contracts.SwapCreatorSwap,
	secret [32]byte,
) (*ethtypes.Receipt, error) {
	s.ethClient.Lock()
	defer s.ethClient.Unlock()
	txOpts, err := s.ethClient.TxOpts(s.ctx)
	if err != nil {
		return nil, err
	}

	tx, err := s.swapContract.Refund(txOpts, *swap, secret)
	if err != nil {
		err = fmt.Errorf("refund tx creation failed, %w", err)
		return nil, err
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ethClient.Raw(), tx.Hash())
	if err != nil {
		err = fmt.Errorf("refund failed, %w", err)
		return nil, err
	}

	return receipt, nil
}
