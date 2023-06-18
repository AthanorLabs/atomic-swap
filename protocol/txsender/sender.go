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
	"errors"
	"fmt"
	"math/big"

	"github.com/cockroachdb/apd/v3"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	logging "github.com/ipfs/go-log/v2"

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
		claimCommitment [32]byte,
		refundCommitment [32]byte,
		claimer ethcommon.Address,
		timeoutDuration *big.Int,
		nonce *big.Int,
		amount coins.EthAssetAmount,
		saveNewSwapTxCallback func(txHash ethcommon.Hash) error,
	) (*ethtypes.Receipt, error)
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
	claimCommitment [32]byte,
	refundCommitment [32]byte,
	claimer ethcommon.Address,
	timeoutDuration *big.Int,
	nonce *big.Int,
	amount coins.EthAssetAmount,
	saveNewSwapTxCallback func(txHash ethcommon.Hash) error,
) (*ethtypes.Receipt, error) {
	// For token swaps, approving our contract to transfer tokens, and calling
	// NewSwap which performs the transfer, need to be inside the same wallet
	// lock grab in case there are other simultaneous swaps happening with the
	// same token.
	s.ethClient.Lock()
	defer s.ethClient.Unlock()

	value := amount.BigInt()

	if amount.IsToken() {
		if err := s.approveTransferFrom(amount); err != nil {
			return nil, err
		}
	}

	txOpts, err := s.ethClient.TxOpts(s.ctx)
	if err != nil {
		return nil, err
	}

	// transfer ETH if we're not doing an ERC20 swap
	if !amount.IsToken() {
		txOpts.Value = value
	}

	tx, err := s.swapCreator.NewSwap(txOpts, claimCommitment, refundCommitment, claimer, timeoutDuration, timeoutDuration,
		amount.TokenAddress(), value, nonce)
	if err != nil {
		err = fmt.Errorf("new_swap tx creation failed, %w", err)
		return nil, err
	}

	if err = saveNewSwapTxCallback(tx.Hash()); err != nil {
		return nil, err
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ethClient.Raw(), tx.Hash())
	if err != nil {
		return nil, fmt.Errorf("NewSwap tx %s failed waiting for receipt, %w", tx.Hash(), err)
	}

	log.Infof("newSwap TX succeeded, %s", common.ReceiptInfo(receipt))

	return receipt, nil
}

// approveTransferFrom grants the SwapCreator contract permission to transfer
// the passed amount of tokens. Since the transfer happens inside the NewSwap
// transaction, we need to grant the contract's address approval for the
// transfer amount first. The ethClient lock should already have been grabbed
// before invoking this method.
func (s *privateKeySender) approveTransferFrom(amount coins.EthAssetAmount) error {
	if !amount.IsToken() {
		panic("this function should only be called with a token asset")
	}

	if amount.AsStd().IsZero() {
		return errors.New("approveContractToTransfer can not be called with a zero amount")
	}

	balance, err := s.ethClient.ERC20Balance(s.ctx, amount.TokenAddress())
	if err != nil {
		return err
	}

	if balance.AsStd().Cmp(amount.AsStd()) < 0 {
		return fmt.Errorf("balance of %s %s is under the %s swap amount",
			balance.AsStdString(), balance.StdSymbol(), amount.AsStdString())
	}

	token := balance.TokenInfo

	// Make a free call to the Allowance function to determine if we need to spend
	// gas approving the SwapCreator contract to transfer the token.
	bindOpts := &bind.CallOpts{Context: s.ctx}
	allowedAmtBI, err := s.erc20Contract.Allowance(bindOpts, s.ethClient.Address(), s.swapCreatorAddr)
	if err != nil {
		return err
	}
	allowedAmt := coins.NewERC20TokenAmountFromBigInt(allowedAmtBI, token)

	if amount.AsStd().Cmp(allowedAmt.AsStd()) <= 0 {
		log.Infof("swapCreator was already approved to transfer %s %s (needed %s)",
			allowedAmt.AsStdString(), allowedAmt.StdSymbol(), amount.AsStdString())
		return nil
	}

	// If the previous approved amount is not zero, some ERC20 contracts like
	// USDT require us to first zero out the approved amount before setting
	// a new value.
	if !allowedAmt.AsStd().IsZero() {
		log.Debugf("zeroing approved token amount before raising limit from %s to %s %s",
			allowedAmt.AsStdString(), amount.AsStdString(), amount.StdSymbol())

		zeroTokenAmt := coins.NewTokenAmountFromDecimals(new(apd.Decimal), token)
		if err = s.approveNoChecks(zeroTokenAmt); err != nil {
			return err
		}
	}

	return s.approveNoChecks(amount)
}

// approveNoChecks is a helper method to gives the SwapCreator contract
// permission to transfer the passed-in amount of tokens. It's caller should be
// doing other checks and have already grabbed the the ethClient's lock.
func (s *privateKeySender) approveNoChecks(amount coins.EthAssetAmount) error {
	txOpts, err := s.ethClient.TxOpts(s.ctx)
	if err != nil {
		return err
	}

	tx, err := s.erc20Contract.Approve(txOpts, s.swapCreatorAddr, amount.BigInt())
	if err != nil {
		return fmt.Errorf("token approve tx for %s %s creation failed, %w",
			amount.AsStdString(), amount.StdSymbol(), err)
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ethClient.Raw(), tx.Hash())
	if err != nil {
		return fmt.Errorf("approveNoChecks tx %s failed waiting for receipt, %w", tx.Hash(), err)
	}

	log.Infof("%s %s approved for use by SwapCreator's new_swap, %s",
		amount.AsStdString(), amount.StdSymbol(), common.ReceiptInfo(receipt))

	return nil
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
