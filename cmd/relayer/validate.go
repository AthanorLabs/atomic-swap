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

var (
	uint256Ty, _ = abi.NewType("uint256", "", nil)
	bytes32Ty, _ = abi.NewType("bytes32", "", nil)
	addressTy, _ = abi.NewType("address", "", nil)
	arguments    = abi.Arguments{
		{
			Name: "owner",
			Type: addressTy,
		},
		{
			Name: "claimer",
			Type: addressTy,
		},
		{
			Name: "pubKeyClaim",
			Type: bytes32Ty,
		},
		{
			Name: "pubKeyRefund",
			Type: bytes32Ty,
		},
		{
			Name: "timeout0",
			Type: uint256Ty,
		},
		{
			Name: "timeout1",
			Type: uint256Ty,
		},
		{
			Name: "asset",
			Type: addressTy,
		},
		{
			Name: "value",
			Type: uint256Ty,
		},
		{
			Name: "nonce",
			Type: uint256Ty,
		},
		{
			Name: "_s",
			Type: bytes32Ty,
		},
		{
			Name: "fee",
			Type: uint256Ty,
		},
	}
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

	args, err := unpackData(req.Data[4:])
	if err != nil {
		return err
	}

	err = validateRelayerFee(args, v.relayerCommission)
	if err != nil {
		return err
	}

	return nil
}

func unpackData(data []byte) (map[string]interface{}, error) {
	args := make(map[string]interface{})
	err := arguments.UnpackIntoMap(args, data)
	if err != nil {
		return nil, err
	}

	return args, nil
}

func validateRelayerFee(args map[string]interface{}, minFeePercentage *apd.Decimal) error {
	value, ok := args["value"].(*big.Int)
	if !ok {
		// this shouldn't happen afaik
		return errors.New("value argument was not marshalled into a *big.Int")
	}

	fee, ok := args["fee"].(*big.Int)
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
	_, err := coins.DecimalCtx().Quo(percentage, feeD, valueD)
	if err != nil {
		return err
	}

	if percentage.Cmp(minFeePercentage) < 0 {
		return fmt.Errorf("fee too low: percentage is %s, expected minimum %s",
			percentage,
			minFeePercentage,
		)
	}

	return nil
}
