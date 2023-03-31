package xmrmaker

import (
	"bytes"
	"fmt"
	"math/big"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
)

// checkContractSwapID checks that the `Swap` type sent matches the swap ID when hashed
func checkContractSwapID(msg *message.NotifyETHLocked) error {
	if msg.ContractSwap.SwapID() != msg.ContractSwapID {
		return errSwapIDMismatch
	}

	return nil
}

// checkContract checks the contract's balance and Claim/Refund keys.
// if the balance doesn't match what we're expecting to receive, or the public keys in the contract
// aren't what we expect, we error and abort the swap.
func (s *swapState) checkContract(txHash ethcommon.Hash) error {
	tx, _, err := s.ETHClient().Raw().TransactionByHash(s.ctx, txHash)
	if err != nil {
		return err
	}

	if tx.To() == nil || *(tx.To()) != s.contractAddr {
		return errInvalidETHLockedTransaction
	}

	receipt, err := s.ETHClient().WaitForReceipt(s.ctx, txHash)
	if err != nil {
		return fmt.Errorf("failed to get receipt for New transaction: %w", err)
	}

	if receipt.Status == 0 {
		// swap transaction reverted
		return errLockTxReverted
	}

	// check that New log was emitted
	if len(receipt.Logs) == 0 {
		return errCannotFindNewLog
	}

	var event *contracts.SwapFactoryNew
	for _, log := range receipt.Logs {
		event, err = s.Contract().ParseNew(*log)
		if err == nil {
			break
		}
	}
	if err != nil {
		return errCannotFindNewLog
	}

	if !bytes.Equal(event.SwapID[:], s.contractSwapID[:]) {
		return errUnexpectedSwapID
	}

	// check that contract was constructed with correct secp256k1 keys
	skOurs := s.secp256k1Pub.Keccak256()
	if !bytes.Equal(event.ClaimKey[:], skOurs[:]) {
		return fmt.Errorf("contract claim key is not expected: got 0x%x, expected 0x%x", event.ClaimKey, skOurs)
	}

	skTheirs := s.xmrtakerSecp256K1PublicKey.Keccak256()
	if !bytes.Equal(event.RefundKey[:], skTheirs[:]) {
		return fmt.Errorf("contract refund key is not expected: got 0x%x, expected 0x%x", event.RefundKey, skTheirs)
	}

	// check asset of created swap
	if types.EthAsset(s.contractSwap.Asset) != types.EthAsset(event.Asset) {
		return fmt.Errorf("swap asset is not expected: got %v, expected %v", event.Asset, s.contractSwap.Asset)
	}

	// check value of created swap
	if s.contractSwap.Value.Cmp(event.Value) != 0 {
		// this should never happen
		return fmt.Errorf("swap value and event value don't match: got %v, expected %v", event.Value, s.contractSwap.Value)
	}

	expectedAmount, err := pcommon.GetEthereumAssetAmount(
		s.ctx,
		s.ETHClient(),
		s.info.ExpectedAmount,
		types.EthAsset(s.contractSwap.Asset),
	)
	if err != nil {
		return err
	}

	if s.contractSwap.Value.Cmp(expectedAmount.BigInt()) != 0 {
		return fmt.Errorf("swap value is not expected: got %v, expected %v",
			s.contractSwap.Value,
			expectedAmount.BigInt(),
		)
	}

	return nil
}

// checkAndSetTimeouts checks that the timeouts set by the counterparty when initiating the swap
// are not too short or too long.
// we expect the timeout to be of a certain length (1 hour for mainnet/stagenet), and allow a 3 minute
// variation between now and the expected time until the first timeout t0, to allow for block confirmations.
// the time between t0 and t1 should always be the exact length we expect.
func (s *swapState) checkAndSetTimeouts(t0, t1 *big.Int) error {
	s.setTimeouts(t0, t1)

	// we ignore the timeout for development, as unit tests and integration tests
	// often set different timeouts.
	if s.Backend.Env() == common.Development {
		return nil
	}

	expectedTimeout := common.SwapTimeoutFromEnv(s.Backend.Env())
	allowableTimeDiff := expectedTimeout / 20

	if s.t1.Sub(s.t0) != expectedTimeout {
		return errInvalidT1
	}

	if time.Now().Add(expectedTimeout).Sub(s.t0).Abs() > allowableTimeDiff {
		return errInvalidT0
	}

	return nil
}

func (s *swapState) setTimeouts(t0, t1 *big.Int) {
	s.t0 = time.Unix(t0.Int64(), 0)
	s.t1 = time.Unix(t1.Int64(), 0)
	s.info.Timeout0 = &s.t0
	s.info.Timeout1 = &s.t1
}
