// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package relayer

import (
	"context"
	"fmt"
	"math/big"

	"github.com/athanorlabs/go-relayer/impls/gsnforwarder"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/net/message"
)

// ValidateAndSendTransaction sends the relayed transaction to the network if it validates successfully.
func ValidateAndSendTransaction(
	ctx context.Context,
	req *message.RelayClaimRequest,
	ec extethclient.EthClient,
	ourSFContractAddr ethcommon.Address,
) (*message.RelayClaimResponse, error) {

	err := validateClaimRequest(ctx, req, ec.Raw(), ourSFContractAddr)
	if err != nil {
		return nil, err
	}

	reqSwapCreator, err := contracts.NewSwapCreator(req.SwapCreatorAddr, ec.Raw())
	if err != nil {
		return nil, err
	}

	reqForwarderAddr, err := reqSwapCreator.TrustedForwarder(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, err
	}

	reqForwarder, domainSeparator, err := getForwarderAndDomainSeparator(ctx, ec.Raw(), reqForwarderAddr)
	if err != nil {
		return nil, err
	}

	nonce, err := reqForwarder.GetNonce(&bind.CallOpts{Context: ctx}, req.Swap.Claimer)
	if err != nil {
		return nil, err
	}

	// The size of request.Secret was vetted when it was deserialized
	secret := (*[32]byte)(req.Secret)

	forwarderReq, err := createForwarderRequest(nonce, req.SwapCreatorAddr, req.Swap, secret)
	if err != nil {
		return nil, err
	}

	gasPrice, err := checkForMinClaimBalance(ctx, ec)
	if err != nil {
		return nil, err
	}

	// Lock the wallet's nonce until we get a receipt
	ec.Lock()
	defer ec.Unlock()

	txOpts, err := ec.TxOpts(ctx)
	if err != nil {
		return nil, err
	}
	txOpts.GasPrice = gasPrice

	tx, err := reqForwarder.Execute(
		txOpts,
		*forwarderReq,
		*domainSeparator,
		gsnforwarder.ForwardRequestTypehash,
		nil,
		req.Signature,
	)
	if err != nil {
		return nil, err
	}

	receipt, err := block.WaitForReceipt(ctx, ec.Raw(), tx.Hash())
	if err != nil {
		return nil, err
	}

	log.Infof("relayed claim %s", common.ReceiptInfo(receipt))

	return &message.RelayClaimResponse{TxHash: tx.Hash()}, nil
}

// checkForMinClaimBalance verifies that we have enough gas to relay a claim and
// returns the gas price that was used for the calculation.
func checkForMinClaimBalance(ctx context.Context, ec extethclient.EthClient) (*big.Int, error) {
	balance, err := ec.Balance(ctx)
	if err != nil {
		return nil, err
	}

	gasPrice, err := ec.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	txCost := new(big.Int).Mul(gasPrice, big.NewInt(forwarderClaimGas))
	if balance.BigInt().Cmp(txCost) < 0 {
		return nil, fmt.Errorf("balance %s ETH is under the minimum %s ETH to relay claim",
			balance.AsEtherString(), coins.FmtWeiAsETH(txCost))
	}

	return gasPrice, nil
}
