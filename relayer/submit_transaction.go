// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package relayer

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/net/message"
)

const (
	claimRelayerGas = 100000 // worst case gas usage for the claimRelayer call (ether)
	// actual cost is 83967 but that fails in unit tests on "out of gas".
)

// ValidateAndSendTransaction sends the relayed transaction to the network if it validates successfully.
func ValidateAndSendTransaction(
	ctx context.Context,
	req *message.RelayClaimRequest,
	ec extethclient.EthClient,
	ourSwapCreatorAddr ethcommon.Address,
) (*message.RelayClaimResponse, error) {
	err := validateClaimRequest(ctx, req, ec.Raw(), ec.Address(), ourSwapCreatorAddr)
	if err != nil {
		return nil, err
	}

	reqSwapCreator, err := contracts.NewSwapCreator(req.RelaySwap.SwapCreator, ec.Raw())
	if err != nil {
		return nil, err
	}

	// The size of request.Secret was vetted when it was deserialized
	secret := [32]byte(req.Secret)

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
	txOpts.GasLimit = claimRelayerGas
	log.Debugf("relaying tx with gas price %s and gas limit %d", gasPrice, txOpts.GasLimit)

	v := req.Signature[64]
	r := [32]byte(req.Signature[:32])
	s := [32]byte(req.Signature[32:64])

	// err = simulateClaimRelayer(
	// 	ctx,
	// 	ec,
	// 	txOpts,
	// 	req.RelaySwap,
	// 	secret,
	// 	v, r, s,
	// )
	// if err != nil {
	// 	return nil, err
	// }

	tx, err := reqSwapCreator.ClaimRelayer(
		txOpts,
		*req.RelaySwap,
		secret,
		v,
		r,
		s,
	)
	if err != nil {
		log.Errorf("failed to call ClaimRelayer: %s", err)
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

	txCost := new(big.Int).Mul(gasPrice, big.NewInt(claimRelayerGas))
	if balance.BigInt().Cmp(txCost) < 0 {
		return nil, fmt.Errorf("balance %s ETH is under the minimum %s ETH to relay claim",
			balance.AsEtherString(), coins.FmtWeiAsETH(txCost))
	}

	return gasPrice, nil
}

// simulateExecute calls the swap creator's ClaimRelayer function
// with CallContract which executes the method call without mining it into the blockchain.
// https://pkg.go.dev/github.com/ethereum/go-ethereum/ethclient#Client.CallContract
func simulateClaimRelayer(
	ctx context.Context,
	ec extethclient.EthClient,
	txOpts *bind.TransactOpts,
	relaySwap *contracts.SwapCreatorRelaySwap,
	secret [32]byte,
	v uint8,
	r, s [32]byte,
) error {
	// Pack the "claimRelayer" method call
	packed, err := contracts.SwapCreatorParsedABI.Pack(
		"claimRelayer",
		relaySwap,
		secret,
		v,
		r,
		s,
	)
	if err != nil {
		return err
	}

	callMessage := ethereum.CallMsg{
		From:       txOpts.From,
		To:         &relaySwap.SwapCreator,
		Gas:        txOpts.GasLimit,
		GasPrice:   txOpts.GasPrice,
		GasFeeCap:  txOpts.GasFeeCap,
		GasTipCap:  txOpts.GasTipCap,
		Value:      txOpts.Value,
		Data:       packed,
		AccessList: []types.AccessTuple{},
	}

	// Call the "claimRelayer" method
	data, err := ec.Raw().CallContract(ctx, callMessage, nil)
	if err != nil {
		return err
	}

	// Unpack the response data
	response := struct {
		Success bool
		Ret     []byte
	}{Success: false, Ret: []byte{}}

	err = contracts.SwapCreatorParsedABI.UnpackIntoInterface(&response, "claimRelayer", data)
	if err != nil {
		return err
	}

	if !response.Success {
		return errors.New("relayed transaction failed on simulation")
	}

	return nil
}
