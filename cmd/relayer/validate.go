package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"

	rcommon "github.com/athanorlabs/go-relayer/common"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	contracts "github.com/athanorlabs/atomic-swap/ethereum"
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
	ctx              context.Context
	ec               *ethclient.Client
	relayerFee       *big.Int
	forwarderAddress ethcommon.Address
}

func (v *validator) validateTransactionFunc(req *rcommon.SubmitTransactionRequest) error {
	// validate that:
	// 1. the `to` address is a swap contract;
	// 2. the function being called is `claimRelayer`;
	// 3. the fee passed to `claimRelayer` is equal to or greater
	// than our desired fee.

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

	err = validateRelayerFee(args, v.relayerFee)
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

func validateRelayerFee(args map[string]interface{}, minFee *big.Int) error {
	fee, ok := args["fee"].(*big.Int)
	if !ok {
		// this shouldn't happen afaik
		return errors.New("fee argument was not marshalled into a *big.Int")
	}

	if fee.Cmp(minFee) < 0 {
		return fmt.Errorf("fee too low: got %s, expected minimum %s",
			fee,
			minFee,
		)
	}

	return nil
}
