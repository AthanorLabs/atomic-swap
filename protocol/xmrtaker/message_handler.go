// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package xmrtaker

import (
	"context"
	"fmt"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net/message"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
)

// HandleProtocolMessage is called by the network to handle an incoming message.
// If the message received is not the expected type for the point in the protocol we're at,
// this function will return an error.
func (s *swapState) HandleProtocolMessage(msg common.Message) error {
	switch msg := msg.(type) {
	case *message.SendKeysMessage:
		event := newEventKeysReceived(msg)
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
	s.updateStatus(status)
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

	log.Debugf("setting status to %s", status)
	s.updateStatus(status)
	return s.Backend.SwapManager().WriteSwapToDB(s.info)
}

func (s *swapState) handleSendKeysMessage(msg *message.SendKeysMessage) (common.Message, error) {
	if msg.ProvidedAmount == nil {
		return nil, errMissingProvidedAmount
	}

	if msg.ProvidedAmount.Cmp(s.info.ExpectedAmount) < 0 {
		return nil, fmt.Errorf("provided amount is not the same as expected: got %s, expected %s",
			msg.ProvidedAmount.Text('f'),
			s.info.ExpectedAmount.Text('f'),
		)
	}

	if msg.PublicSpendKey == nil || msg.PrivateViewKey == nil {
		return nil, errMissingKeys
	}

	if msg.EthAddress == (ethcommon.Address{}) {
		return nil, errMissingAddress
	}

	vk := msg.PrivateViewKey

	// verify counterparty's DLEq proof and ensure the resulting secp256k1 key is correct
	verificationRes, err := pcommon.VerifyKeysAndProof(msg.DLEqProof, msg.Secp256k1PublicKey, msg.PublicSpendKey)
	if err != nil {
		return nil, err
	}

	s.xmrmakerAddress = msg.EthAddress
	log.Debugf("got XMRMaker's keys and address: address=%s", s.xmrmakerAddress)

	symbol, err := pcommon.AssetSymbol(s.Backend, s.info.EthAsset)
	if err != nil {
		return nil, err
	}

	log.Infof(color.New(color.Bold).Sprintf("receiving %v XMR for %v %s",
		msg.ProvidedAmount,
		s.info.ProvidedAmount,
		symbol,
	))

	err = s.setXMRMakerKeys(verificationRes.Ed25519PublicKey, vk, verificationRes.Secp256k1PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to set xmrmaker keys: %w", err)
	}
	log.Debugf("stored XMR maker's keys, going to lock ETH")

	receipt, err := s.lockAsset()
	if err != nil {
		return nil, fmt.Errorf("failed to lock ethereum asset in contract: %w", err)
	}

	// start goroutine to check that XMRMaker locks before t_0
	go s.runT1ExpirationHandler()

	// start goroutine to check for xmr being locked
	go s.checkForXMRLock()

	out := &message.NotifyETHLocked{
		Address:        s.SwapCreatorAddr(),
		TxHash:         receipt.TxHash,
		ContractSwapID: s.contractSwapID,
		ContractSwap:   s.contractSwap,
	}

	return out, nil
}

func (s *swapState) checkForXMRLock() {
	var checkForXMRLockInterval time.Duration
	if s.Env() == common.Development {
		checkForXMRLockInterval = time.Second
	} else {
		// monero block time is >1 minute, so this should be fine
		checkForXMRLockInterval = time.Minute
	}

	// check that XMR was locked in expected account, and confirm amount
	lockedAddr, vk := s.expectedXMRLockAccount()

	conf := s.XMRClient().CreateWalletConf("xmrtaker-swap-wallet-verify-funds")
	abViewCli, err := monero.CreateViewOnlyWalletFromKeys(conf, vk, lockedAddr, s.walletScanHeight)
	if err != nil {
		log.Errorf("failed to generate view-only wallet to verify locked XMR: %s", err)
		return
	}
	defer abViewCli.CloseAndRemoveWallet()

	log.Debugf("generated view-only wallet to check funds: %s", abViewCli.WalletName())

	timer := time.NewTicker(checkForXMRLockInterval)
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-timer.C:
			balance, err := abViewCli.GetBalance(0)
			if err != nil {
				log.Errorf("failed to get balance: %s", err)
				continue
			}

			log.Debugf("checking locked wallet, address=%s balance=%d blocks-to-unlock=%d",
				lockedAddr, balance.Balance, balance.BlocksToUnlock)

			if s.expectedPiconeroAmount().CmpU64(balance.UnlockedBalance) <= 0 {
				event := newEventXMRLocked()
				s.eventCh <- event
				err := <-event.errCh
				if err != nil {
					log.Errorf("eventXMRLocked errored: %s", err)
				}

				return
			}
		}
	}
}

func (s *swapState) runT1ExpirationHandler() {
	defer log.Debugf("returning from runT1ExpirationHandler")

	if time.Until(s.t1) <= 0 {
		log.Debugf("T1 already passed, not starting T1 expiration handler")
		return
	}

	// TODO: this variable is so that we definitely refund before t1.
	// Current algorithm is to trigger the timeout when only 15% of the allotted
	// time is remaining. If the block interval is 1 second on a test network and
	// and T1 is 7 seconds after swap creation, we need the refund to trigger more
	// than one second before the block with a timestamp exactly equal to T1 to
	// satisfy the strictly less than requirement. 7s * 15% = 1.05s. 15% remaining
	// may be reasonable even with large timeouts on production networks, but more
	// research is needed.
	t1Delta := s.t2.Sub(s.t1) // time between swap start and T1 is equal to T2-T1
	deltaBeforeT1ToGiveUp := time.Duration(float64(t1Delta) * 0.15)
	deltaUntilGiveUp := time.Until(s.t1) - deltaBeforeT1ToGiveUp
	giveUpAndRefundTimer := time.NewTimer(deltaUntilGiveUp)
	defer giveUpAndRefundTimer.Stop() // don't wait for the timeout to garbage collect
	log.Debugf("time until refund: %vs", deltaUntilGiveUp.Seconds())

	select {
	case <-s.ctx.Done():
		return
	case <-s.xmrLockedCh:
		return
	case <-giveUpAndRefundTimer.C:
		log.Infof("approaching T1, attempting to refund ETH")
		event := newEventShouldRefund()
		s.eventCh <- event
		err := <-event.errCh
		if err != nil {
			// TODO: what should we do here? this would be bad. (#162)
			log.Errorf("failed to refund: %s", err)
		}
	}
}

func (s *swapState) expectedXMRLockAccount() (*mcrypto.Address, *mcrypto.PrivateViewKey) {
	vk := mcrypto.SumPrivateViewKeys(s.xmrmakerPrivateViewKey, s.privkeys.ViewKey())
	sk := mcrypto.SumPublicKeys(s.xmrmakerPublicSpendKey, s.pubkeys.SpendKey())
	return mcrypto.NewPublicKeyPair(sk, vk.Public()).Address(s.Env()), vk
}

func (s *swapState) handleNotifyXMRLock() error {
	close(s.xmrLockedCh)
	log.Info("XMR was locked successfully, setting contract to ready...")

	if err := s.setReady(); err != nil {
		return fmt.Errorf("failed to call Ready: %w", err)
	}

	go s.runT2ExpirationHandler()
	return nil
}

func (s *swapState) runT2ExpirationHandler() {
	log.Debugf("time until t2 (%s): %vs",
		s.t2.Format(common.TimeFmtSecs),
		time.Until(s.t2).Seconds(),
	)

	defer log.Debugf("returning from runT2ExpirationHandler")

	waitCtx, waitCtxCancel := context.WithCancel(context.Background())
	defer waitCtxCancel() // Unblock WaitForTimestamp if still running when we exit

	waitCh := make(chan error)
	go func() {
		waitCh <- s.ETHClient().WaitForTimestamp(waitCtx, s.t2)
		close(waitCh)
	}()

	select {
	case <-s.ctx.Done():
		return
	case <-s.claimedCh:
		return
	case err := <-waitCh:
		if err != nil {
			// TODO: Do we propagate this error? If we retry, the logic should probably be inside
			// WaitForTimestamp. (#162)
			log.Errorf("failure waiting for T2 timeout: %s", err)
			return
		}
		s.handleT2Expired()
	}
}

func (s *swapState) handleT2Expired() {
	log.Debugf("handling T2")
	event := newEventShouldRefund()
	s.eventCh <- event
	err := <-event.errCh
	if err != nil {
		// TODO: what should we do here? this would be bad. (#162)
		log.Errorf("failed to refund: %s", err)
	}
}
