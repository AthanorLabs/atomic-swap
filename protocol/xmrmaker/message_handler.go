// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package xmrmaker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/db"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
)

// HandleProtocolMessage is called by the network to handle an incoming message.
// If the message received is not the expected type for the point in the protocol we're at,
// this function will return an error.
func (s *swapState) HandleProtocolMessage(msg common.Message) error {
	if s == nil {
		return errNilSwapState
	}

	if s.ctx.Err() != nil {
		return fmt.Errorf("protocol exited: %w", s.ctx.Err())
	}

	switch msg := msg.(type) {
	case *message.NotifyETHLocked:
		event := newEventETHLocked(msg)
		s.eventCh <- event
		err := <-event.errCh
		if err != nil {
			return err
		}
	default:
		return errUnexpectedMessageType
	}

	return nil
}

func (s *swapState) clearNextExpectedEvent(status types.Status) {
	s.nextExpectedEvent = EventNoneType
	s.info.SetStatus(status)
}

func (s *swapState) setNextExpectedEvent(event EventType) error {
	if s.nextExpectedEvent == EventNoneType {
		// should have called clearNextExpectedEvent instead
		panic("cannot set next expected event to EventNoneType")
	}

	if event == s.nextExpectedEvent {
		panic("cannot set next expected event to same as current")
	}

	s.nextExpectedEvent = event
	status := event.getStatus()

	if status == types.UnknownStatus {
		panic("status corresponding to event cannot be UnknownStatus")
	}

	s.info.SetStatus(status)
	err := s.Backend.SwapManager().WriteSwapToDB(s.info)
	if err != nil {
		return err
	}

	return nil
}

// waitForNewSwapReceipt waits for the newSwap transaction, that locks the
// taker's ETH, to be seen as included in a block by our endpoint. This is a
// pre-requirement for validating the newSwap transaction, which should be done
// after calling this method.
func waitForNewSwapReceipt(
	ctx context.Context,
	ec *ethclient.Client,
	txHash ethcommon.Hash,
) (*ethtypes.Receipt, error) {
	const loopPause = 1500 * time.Millisecond // 1.5 seconds

	// In mainnet testing, when the maker and taker are using different ETH
	// endpoints, we've seen cases where the taker receives a TX receipt and
	// transmits the hash to the maker before the maker's side thinks the TX has
	// been included in a block. We wait for up to 15 seconds if our attempts at
	// getting the transaction receipt return NotFound.
	for i := 0; i < 10; i++ {
		receipt, err := ec.TransactionReceipt(ctx, txHash)
		if err != nil && !errors.Is(err, ethereum.NotFound) {
			return nil, err
		}
		// If err is still set, the error was ethereum.NotFound, which is returned
		// even if our endpoint sees the TX as pending.
		if err != nil {
			if err = common.SleepWithContext(ctx, loopPause); err != nil {
				return nil, err // context expired
			}
			continue
		}

		if receipt.Status != ethtypes.ReceiptStatusSuccessful {
			return nil, fmt.Errorf("received newSwap tx=%s was reverted", txHash.Hex())
		}

		return receipt, nil
	}

	return nil, ethereum.NotFound
}

func (s *swapState) handleNotifyETHLocked(msg *message.NotifyETHLocked) error {
	if msg.Address == (ethcommon.Address{}) {
		return errMissingAddress
	}

	if types.IsHashZero(msg.ContractSwapID) {
		return errNilContractSwapID
	}

	log.Infof("got NotifyETHLocked; address=%s contract swap ID=%s", msg.Address, msg.ContractSwapID)

	// validate that swap ID == keccak256(swap struct)
	if msg.ContractSwap.SwapID() != msg.ContractSwapID {
		return errSwapIDMismatch
	}

	s.contractSwapID = msg.ContractSwapID
	s.contractSwap = msg.ContractSwap

	receipt, err := waitForNewSwapReceipt(s.ctx, s.Backend.ETHClient().Raw(), msg.TxHash)
	if err != nil {
		return err
	}

	contractAddr := msg.Address
	err = contracts.CheckSwapCreatorContractCode(s.ctx, s.Backend.ETHClient().Raw(), contractAddr)
	if err != nil {
		return err
	}

	if err = s.setContract(contractAddr); err != nil {
		return fmt.Errorf("failed to instantiate contract instance: %w", err)
	}

	ethInfo := &db.EthereumSwapInfo{
		StartNumber:     receipt.BlockNumber,
		SwapID:          s.contractSwapID,
		Swap:            s.contractSwap,
		SwapCreatorAddr: contractAddr,
	}

	if err = s.Backend.RecoveryDB().PutContractSwapInfo(s.OfferID(), ethInfo); err != nil {
		return err
	}

	log.Infof("stored ContractSwapInfo: id=%s", s.OfferID())

	if err = s.checkContract(msg.TxHash); err != nil {
		return err
	}

	err = s.checkAndSetTimeouts(msg.ContractSwap.Timeout1, msg.ContractSwap.Timeout2)
	if err != nil {
		return err
	}

	err = s.lockFunds(coins.MoneroToPiconero(s.info.ProvidedAmount))
	if err != nil {
		return fmt.Errorf("failed to lock funds: %w", err)
	}

	go s.runT1ExpirationHandler()
	return nil
}

func (s *swapState) runT1ExpirationHandler() {
	log.Debugf("time until t1 (%s): %vs",
		s.t1.Format(common.TimeFmtSecs),
		time.Until(s.t1).Seconds(),
	)

	if time.Until(s.t2) < 0 {
		log.Debugf("t2 (%s) has already passed; not starting t1 expiration handler",
			s.t2.Format(common.TimeFmtSecs),
		)
		return
	}

	waitCtx, waitCtxCancel := context.WithCancel(context.Background())
	defer waitCtxCancel() // Unblock WaitForTimestamp if still running when we exit

	// note: this will cause unit tests to hang if not running ganache
	// with --miner.blockTime!!!
	waitCh := make(chan error)
	go func() {
		waitCh <- s.ETHClient().WaitForTimestamp(waitCtx, s.t1)
		close(waitCh)
	}()

	select {
	case <-s.ctx.Done():
		return
	case <-s.readyCh:
		log.Debugf("returning from runT1ExpirationHandler as contract was set to ready")
		return
	case err := <-waitCh:
		if err != nil {
			// TODO: Do we propagate this error? If we retry, the logic should probably be inside
			// WaitForTimestamp. (#162)
			log.Errorf("Failure waiting for T1 timeout: err=%s", err)
			return
		}
		log.Debugf("reached t1, time to claim")
		s.handleT1Expired()
	}
}

func (s *swapState) handleT1Expired() {
	event := newEventContractReady()
	s.eventCh <- event
	err := <-event.errCh
	if err != nil {
		// TODO: this is quite bad, how should this be handled? (#162)
		log.Errorf("failed to handle t1 expiration: %s", err)
	}
}

func (s *swapState) handleSendKeysMessage(msg *message.SendKeysMessage) error {
	if msg.PublicSpendKey == nil || msg.PrivateViewKey == nil {
		return errMissingKeys
	}

	// verify counterparty's DLEq proof and ensure the resulting secp256k1 key is correct
	verifyResult, err := pcommon.VerifyKeysAndProof(msg.DLEqProof, msg.Secp256k1PublicKey, msg.PublicSpendKey)
	if err != nil {
		return err
	}

	return s.setXMRTakerKeys(msg.PublicSpendKey, msg.PrivateViewKey, verifyResult.Secp256k1PublicKey)
}
