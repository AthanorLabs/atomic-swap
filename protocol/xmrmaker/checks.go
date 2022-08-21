package xmrmaker

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/noot/atomic-swap/net/message"
)

// checkContractSwapID checks that the `Swap` type sent matches the swap ID when hashed
func checkContractSwapID(msg *message.NotifyETHLocked) error {
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
			Type: uint256Ty,
		},
		{
			Type: uint256Ty,
		},
	}

	args, err := arguments.Pack(
		msg.ContractSwap.Owner,
		msg.ContractSwap.Claimer,
		msg.ContractSwap.PubKeyClaim,
		msg.ContractSwap.PubKeyRefund,
		msg.ContractSwap.Timeout0,
		msg.ContractSwap.Timeout1,
		msg.ContractSwap.Asset,
		msg.ContractSwap.Value,
		msg.ContractSwap.Nonce,
	)
	if err != nil {
		return fmt.Errorf("failed to pack arguments: %w", err)
	}

	hash := crypto.Keccak256Hash(args)
	if !bytes.Equal(hash[:], msg.ContractSwapID[:]) {
		return errSwapIDMismatch
	}

	return nil
}
