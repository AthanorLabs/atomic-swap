// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package contracts is for go bindings generated from Solidity contracts as well as
// some utility functions for working with the contracts.
package contracts

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
)

// Swap stage values that match the names and indexes of the Stage enum in
// the SwapCreator contract
const (
	StageInvalid byte = iota
	StagePending
	StageReady
	StageCompleted
)

var (
	// SwapCreatorParsedABI is the parsed SwapCreator ABI. We can skip the error check,
	// as it can only fail if abigen generates JSON bindings that golang can't parse, in
	// which case it will be nil we'll see panics when vetting the binaries.
	SwapCreatorParsedABI, _ = SwapCreatorMetaData.GetAbi()

	claimedTopic  = common.GetTopic(ClaimedEventSignature)
	refundedTopic = common.GetTopic(RefundedEventSignature)

	NewSwapFunctionSignature = SwapCreatorParsedABI.Methods["newSwap"].Sig //nolint:revive
	ReadyEventSignature      = SwapCreatorParsedABI.Events["Ready"].Sig    //nolint:revive
	ClaimedEventSignature    = SwapCreatorParsedABI.Events["Claimed"].Sig  //nolint:revive
	RefundedEventSignature   = SwapCreatorParsedABI.Events["Refunded"].Sig //nolint:revive
)

// StageToString converts a contract Stage enum value to a string
func StageToString(stage byte) string {
	switch stage {
	case StageInvalid:
		return "Invalid"
	case StagePending:
		return "Pending"
	case StageReady:
		return "Ready"
	case StageCompleted:
		return "Completed"
	default:
		return fmt.Sprintf("UnknownStageValue(%d)", stage)
	}
}

// Hash abi-encodes the RelaySwap and returns the keccak256 hash of the encoded value.
func (s *SwapCreatorRelaySwap) Hash() types.Hash {
	uint256Ty, err := abi.NewType("uint256", "", nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create uint256 type: %s", err))
	}

	bytes32Ty, err := abi.NewType("bytes32", "", nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create bytes32 type: %s", err))
	}

	addressTy, err := abi.NewType("address", "", nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create address type: %s", err))
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
		{
			Type: uint256Ty,
		},
		{
			Type: bytes32Ty,
		},
		{
			Type: addressTy,
		},
	}

	args, err := arguments.Pack(
		s.Swap.Owner,
		s.Swap.Claimer,
		s.Swap.ClaimCommitment,
		s.Swap.RefundCommitment,
		s.Swap.Timeout1,
		s.Swap.Timeout2,
		s.Swap.Asset,
		s.Swap.Value,
		s.Swap.Nonce,
		s.Fee,
		s.RelayerHash,
		s.SwapCreator,
	)
	if err != nil {
		// As long as none of the *big.Int fields are nil, this cannot fail.
		// When receiving SwapCreatorRelaySwap objects from peers in
		// JSON, all *big.Int values are pre-validated to be non-nil.
		panic(fmt.Sprintf("failed to pack arguments: %s", err))
	}

	return crypto.Keccak256Hash(args)
}

// SwapID calculates and returns the same hashed swap identifier that newSwap
// emits and that is used to track the on-chain stage of a swap.
func (sfs *SwapCreatorSwap) SwapID() types.Hash {
	uint256Ty, err := abi.NewType("uint256", "", nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create uint256 type: %s", err))
	}

	bytes32Ty, err := abi.NewType("bytes32", "", nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create bytes32 type: %s", err))
	}

	addressTy, err := abi.NewType("address", "", nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create address type: %s", err))
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
		sfs.Owner,
		sfs.Claimer,
		sfs.ClaimCommitment,
		sfs.RefundCommitment,
		sfs.Timeout1,
		sfs.Timeout2,
		sfs.Asset,
		sfs.Value,
		sfs.Nonce,
	)
	if err != nil {
		// As long as none of the *big.Int fields are nil, this cannot fail.
		// When receiving SwapCreatorSwap objects from the database or peers in
		// JSON, all *big.Int values are pre-validated to be non-nil.
		panic(fmt.Sprintf("failed to pack arguments: %s", err))
	}

	return crypto.Keccak256Hash(args)
}

// GetSecretFromLog returns the secret from a Claimed or Refunded log
func GetSecretFromLog(log *ethtypes.Log, eventTopic [32]byte) (*mcrypto.PrivateSpendKey, error) {
	if eventTopic != claimedTopic && eventTopic != refundedTopic {
		return nil, errors.New("invalid event, must be one of Claimed or Refunded")
	}

	if len(log.Topics) < 3 {
		return nil, errors.New("log had not enough parameters")
	}

	s := log.Topics[2]
	if s == [32]byte{} {
		return nil, errors.New("got zero secret key from contract")
	}

	sk, err := mcrypto.NewPrivateSpendKey(common.Reverse(s[:]))
	if err != nil {
		return nil, err
	}

	return sk, nil
}

// CheckIfLogIDMatches returns true if the swap ID in the log matches the given ID, false otherwise.
func CheckIfLogIDMatches(log ethtypes.Log, eventTopic, id [32]byte) (bool, error) {
	if eventTopic != claimedTopic && eventTopic != refundedTopic {
		return false, errors.New("invalid event, must be one of Claimed or Refunded")
	}

	if len(log.Topics) < 2 {
		return false, errors.New("log had not enough parameters")
	}

	eventID := log.Topics[1]
	if !bytes.Equal(eventID[:], id[:]) {
		return false, nil
	}

	return true, nil
}

// GetIDFromLog returns the swap ID from a New log.
func GetIDFromLog(log *ethtypes.Log) ([32]byte, error) {
	abi := SwapCreatorParsedABI

	const event = "New"
	if log.Topics[0] != abi.Events[event].ID {
		// Wrong log
		return [32]byte{}, errors.New("wrong log topic")
	}

	data := log.Data
	res, err := abi.Unpack(event, data)
	if err != nil {
		return [32]byte{}, err
	}

	if len(res) == 0 {
		return [32]byte{}, errors.New("log didn't have enough parameters")
	}

	id := res[0].([32]byte)
	return id, nil
}

// GetTimeoutsFromLog returns the timeouts from a New event.
func GetTimeoutsFromLog(log *ethtypes.Log) (*big.Int, *big.Int, error) {
	abi := SwapCreatorParsedABI

	const event = "New"
	if log.Topics[0] != abi.Events[event].ID {
		// Wrong log
		return nil, nil, errors.New("wrong log topic")
	}

	data := log.Data
	res, err := abi.Unpack(event, data)
	if err != nil {
		return nil, nil, err
	}

	if len(res) < 5 {
		return nil, nil, errors.New("log didn't have enough parameters")
	}

	t1 := res[3].(*big.Int)
	t2 := res[4].(*big.Int)
	return t1, t2, nil
}

// GenerateNewSwapNonce generates a random nonce value for use with NewSwap
// transactions.
func GenerateNewSwapNonce() *big.Int {
	u256PlusOne := new(big.Int).Lsh(big.NewInt(1), 256)
	maxU256 := new(big.Int).Sub(u256PlusOne, big.NewInt(1))
	n, _ := rand.Int(rand.Reader, maxU256)
	return n
}
