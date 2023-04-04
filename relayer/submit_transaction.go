// Copyright 2023 Athanor Labs (ON)
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

	reqSwapFactory, err := contracts.NewSwapFactory(req.SwapFactoryAddress, ec.Raw())
	if err != nil {
		return nil, err
	}

	reqForwarderAddr, err := reqSwapFactory.TrustedForwarder(&bind.CallOpts{Context: ctx})
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

	forwarderReq, err := createForwarderRequest(nonce, req.SwapFactoryAddress, req.Swap, secret)
	if err != nil {
		return nil, err
	}

	if err = checkForMinClaimBalance(ctx, ec); err != nil {
		return nil, err
	}

	// Lock the wallet's nonce until we get a receipt
	ec.Lock()
	defer ec.Unlock()

	txOpts, err := ec.TxOpts(ctx)
	if err != nil {
		return nil, err
	}

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

func checkForMinClaimBalance(ctx context.Context, ec extethclient.EthClient) error {
	balance, err := ec.Balance(ctx)
	if err != nil {
		return err
	}

	suggestedGasPrice, err := ec.Raw().SuggestGasPrice(ctx)
	if err != nil {
		return err
	}

	txCost := new(big.Int).Mul(suggestedGasPrice, big.NewInt(forwarderClaimGas))
	if balance.Cmp(txCost) < 0 {
		return fmt.Errorf("balance %s ETH is under the minimum %s ETH to relay claim",
			coins.FmtWeiAsETH(balance), coins.FmtWeiAsETH(txCost))
	}

	return nil
}
