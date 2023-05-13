// Copyright 2023 The AthanorLabs/atomic-swap Authors
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

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
)

var (
	log = logging.Logger("txsender")
)

// Sender signs and submits transactions to the chain
type Sender interface {
	SetSwapCreator(*contracts.SwapCreator)
	SetSwapCreatorAddr(ethcommon.Address)
	NewSwap(
		pubKeyClaim [32]byte,
		pubKeyRefund [32]byte,
		claimer ethcommon.Address,
		timeoutDuration *big.Int,
		nonce *big.Int,
		amount coins.EthAssetAmount,
	) (ethcommon.Hash, error)
	SetReady(swap *contracts.SwapCreatorSwap) (*ethtypes.Receipt, error)
	Claim(swap *contracts.SwapCreatorSwap, secret [32]byte) (*ethtypes.Receipt, error)
	Refund(swap *contracts.SwapCreatorSwap, secret [32]byte) (*ethtypes.Receipt, error)
}

type privateKeySender struct {
	ctx             context.Context
	ethClient       extethclient.EthClient
	swapCreatorAddr ethcommon.Address
	swapCreator     *contracts.SwapCreator
	erc20Contract   *contracts.IERC20
}

// NewSenderWithPrivateKey returns a new *privateKeySender
func NewSenderWithPrivateKey(
	ctx context.Context,
	ethClient extethclient.EthClient,
	swapCreatorAddr ethcommon.Address,
	swapCreator *contracts.SwapCreator,
	erc20Contract *contracts.IERC20,
) Sender {
	return &privateKeySender{
		ctx:             ctx,
		ethClient:       ethClient,
		swapCreatorAddr: swapCreatorAddr,
		swapCreator:     swapCreator,
		erc20Contract:   erc20Contract,
	}
}

func (s *privateKeySender) SetSwapCreator(contract *contracts.SwapCreator) {
	s.swapCreator = contract
}

func (s *privateKeySender) SetSwapCreatorAddr(_ ethcommon.Address) {}

func (s *privateKeySender) NewSwap(
	pubKeyClaim [32]byte,
	pubKeyRefund [32]byte,
	claimer ethcommon.Address,
	timeoutDuration *big.Int,
	nonce *big.Int,
	amount coins.EthAssetAmount,
) (ethcommon.Hash, error) {
	s.ethClient.Lock()
	defer s.ethClient.Unlock()

	value := amount.BigInt()

	// For token swaps, approving our contract to transfer tokens, and calling
	// NewSwap which performs the transfer, needs to be inside the same wallet
	// lock grab in case there are other simultaneous swaps happening with the
	// same token.
	if amount.IsToken() {
		txOpts, err := s.ethClient.TxOpts(s.ctx)
		if err != nil {
			return ethcommon.Hash{}, err
		}

		tx, err := s.erc20Contract.Approve(txOpts, s.swapCreatorAddr, value)
		if err != nil {
			return ethcommon.Hash{}, fmt.Errorf("approve tx creation failed, %w", err)
		}

		receipt, err := block.WaitForReceipt(s.ctx, s.ethClient.Raw(), tx.Hash())
		if err != nil {
			return ethcommon.Hash{}, fmt.Errorf("approve failed, %w", err)
		}

		log.Debugf("approve transaction included %s", common.ReceiptInfo(receipt))
		log.Infof("%s %s approved for use by SwapCreator's new_swap",
			amount.AsStandard().Text('f'), amount.StandardSymbol())
	}

	txOpts, err := s.ethClient.TxOpts(s.ctx)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	// transfer ETH if we're not doing an ERC20 swap
	if !amount.IsToken() {
		txOpts.Value = value
	}

	tx, err := s.swapCreator.NewSwap(txOpts, pubKeyClaim, pubKeyRefund, claimer, timeoutDuration, timeoutDuration,
		amount.TokenAddress(), value, nonce)
	if err != nil {
		err = fmt.Errorf("new_swap tx creation failed, %w", err)
		return ethcommon.Hash{}, err
	}

	return tx.Hash(), nil
}

func (s *privateKeySender) SetReady(swap *contracts.SwapCreatorSwap) (*ethtypes.Receipt, error) {
	s.ethClient.Lock()
	defer s.ethClient.Unlock()
	txOpts, err := s.ethClient.TxOpts(s.ctx)
	if err != nil {
		return nil, err
	}

	tx, err := s.swapCreator.SetReady(txOpts, *swap)
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

	tx, err := s.swapCreator.Claim(txOpts, *swap, secret)
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

	tx, err := s.swapCreator.Refund(txOpts, *swap, secret)
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
