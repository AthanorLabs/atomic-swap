package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/athanorlabs/atomic-swap/coins"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	rcommon "github.com/athanorlabs/go-relayer/common"

	"github.com/cockroachdb/apd/v3"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type validator struct {
	ctx               context.Context
	ec                *ethclient.Client
	relayerCommission *apd.Decimal
	forwarderAddress  ethcommon.Address
}

func (v *validator) validateTransactionFunc(req *rcommon.SubmitTransactionRequest) error {
	// validate that:
	// 1. the `to` address is a swap contract;
	// 2. the function being called is `claimRelayer`;
	// 3. the fee passed to `claimRelayer` is equal to or greater
	// than our desired commission percentage.

	forwarderAddr, err := contracts.CheckSwapFactoryContractCode(
		v.ctx, v.ec, req.To,
	)
	if err != nil {
		return err
	}

	if forwarderAddr != v.forwarderAddress {
		return fmt.Errorf("swap contract does not have expected forwarder address: got %s, expected %s",
			forwarderAddr,
			v.forwarderAddress,
		)
	}

	// hardcoded, from swap_factory.go bindings
	claimRelayerSig := ethcommon.FromHex("0x73e4771c")
	if !bytes.Equal(claimRelayerSig, req.Data[:4]) {
		return fmt.Errorf("call must be to claimRelayer(); got call to function with sig 0x%x", req.Data[:4])
	}

	err = validateRelayerFee(req.Data[4:], v.relayerCommission)
	if err != nil {
		return err
	}

	return nil
}

func validateRelayerFee(data []byte, minFeePercentage *apd.Decimal) error {
	uint256Ty, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return fmt.Errorf("failed to create uint256 type: %w", err)
	}

	bytes32Ty, err := abi.NewType("bytes32", "", nil)
	if err != nil {
		return fmt.Errorf("failed to create bytes32 type: %w", err)
	}

	addressTy, err := abi.NewType("address", "", nil)
	if err != nil {
		return fmt.Errorf("failed to create address type: %w", err)
	}

	arguments := abi.Arguments{
		// Swap
		{
			Type: addressTy,
		},
		{
			Type: addressTy,
		},
		{
			Type: bytes32Ty,
		},
		{
			Type: bytes32Ty,
		},
		{
			Type: uint256Ty,
		},
		{
			Type: uint256Ty,
		},
		{
			Type: addressTy,
		},
		{
			Type: uint256Ty, // value
		},
		{
			Type: uint256Ty,
		},
		// _s
		{
			Type: bytes32Ty,
		},
		// _fee
		{
			Type: uint256Ty,
		},
	}
	args, err := arguments.Unpack(data)
	if err != nil {
		return err
	}

	value, ok := args[7].(*big.Int)
	if !ok {
		// this shouldn't happen afaik
		return errors.New("value argument was not marshalled into a *big.Int")
	}

	fee, ok := args[10].(*big.Int)
	if !ok {
		// this shouldn't happen afaik
		return errors.New("fee argument was not marshalled into a *big.Int")
	}

	valueD := apd.NewWithBigInt(
		new(apd.BigInt).SetMathBigInt(value), // swap value, in wei
		0,
	)
	feeD := apd.NewWithBigInt(
		new(apd.BigInt).SetMathBigInt(fee), // fee, in wei
		0,
	)

	percentage := new(apd.Decimal)
	_, err = coins.DecimalCtx().Quo(percentage, feeD, valueD)
	if err != nil {
		return err
	}

	if percentage.Cmp(minFeePercentage) == -1 {
		return fmt.Errorf("fee too low: percentage is %s, expected minimum %s",
			percentage,
			minFeePercentage,
		)
	}

	return nil
}
