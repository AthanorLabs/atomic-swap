// Package contracts is for go bindings generated from Solidity contracts as well as
// some utility functions for working with the contracts.
package contracts

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
)

// Swap stage values that match the names and indexes of the Stage enum in
// the SwapFactory contract
const (
	StageInvalid byte = iota
	StagePending
	StageReady
	StageCompleted
)

var (
	claimedTopic  = common.GetTopic(common.ClaimedEventSignature)
	refundedTopic = common.GetTopic(common.RefundedEventSignature)
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

// GetSecretFromLog returns the secret from a Claimed or Refunded log
func GetSecretFromLog(log *ethtypes.Log, eventTopic [32]byte) (*mcrypto.PrivateSpendKey, error) {
	if eventTopic != claimedTopic && eventTopic != refundedTopic {
		return nil, errors.New("invalid event, must be one of Claimed or Refunded")
	}

	// abiSF, err := abi.JSON(strings.NewReader(SwapFactoryMetaData.ABI))
	// if err != nil {
	// 	return nil, err
	// }

	// data := log.Data
	// res, err := abiSF.Unpack(event, data)
	// if err != nil {
	// 	return nil, err
	// }

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

	// abi, err := abi.JSON(strings.NewReader(SwapFactoryMetaData.ABI))
	// if err != nil {
	// 	return false, err
	// }

	// data := log.Data
	// res, err := abi.Unpack(event, data)
	// if err != nil {
	// 	return false, err
	// }

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
	abi, err := abi.JSON(strings.NewReader(SwapFactoryMetaData.ABI))
	if err != nil {
		return [32]byte{}, err
	}

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
	abi, err := abi.JSON(strings.NewReader(SwapFactoryMetaData.ABI))
	if err != nil {
		return nil, nil, err
	}

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

	t0 := res[3].(*big.Int)
	t1 := res[4].(*big.Int)
	return t0, t1, nil
}
