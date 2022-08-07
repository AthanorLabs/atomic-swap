package xmrtaker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/net/message"
	pcommon "github.com/noot/atomic-swap/protocol"
	"github.com/noot/atomic-swap/swapfactory"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color" //nolint:misspell
)

// HandleProtocolMessage is called by the network to handle an incoming message.
// If the message received is not the expected type for the point in the protocol we're at,
// this function will return an error.
func (s *swapState) HandleProtocolMessage(msg net.Message) (net.Message, bool, error) {
	s.lockState()
	defer s.unlockState()

	if err := s.checkMessageType(msg); err != nil {
		return nil, true, err
	}

	switch msg := msg.(type) {
	case *net.SendKeysMessage:
		resp, err := s.handleSendKeysMessage(msg)
		if err != nil {
			return nil, true, err
		}

		return resp, false, nil
	case *message.NotifyXMRLock:
		out, err := s.handleNotifyXMRLock(msg)
		if err != nil {
			return nil, true, err
		}

		return out, false, nil
	case *message.NotifyClaimed:
		_, err := s.handleNotifyClaimed(msg.TxHash)
		if err != nil {
			log.Error("failed to create monero address: err=", err)
			return nil, true, err
		}

		s.clearNextExpectedMessage(types.CompletedSuccess)
		return nil, true, nil
	default:
		return nil, false, errUnexpectedMessageType
	}
}

func (s *swapState) clearNextExpectedMessage(status types.Status) {
	s.nextExpectedMessage = nil
	s.info.SetStatus(status)
	if s.statusCh != nil {
		s.statusCh <- status
	}
}

func (s *swapState) setNextExpectedMessage(msg net.Message) {
	if s == nil || s.nextExpectedMessage == nil {
		return
	}

	if msg.Type() == s.nextExpectedMessage.Type() {
		return
	}

	s.nextExpectedMessage = msg

	// TODO: check stage is not unknown (ie. swap completed)
	stage := pcommon.GetStatus(msg.Type())
	if s.statusCh != nil {
		s.statusCh <- stage
	}
}

func (s *swapState) checkMessageType(msg net.Message) error {
	if msg == nil {
		return errNilMessage
	}

	if s.nextExpectedMessage == nil {
		return nil
	}

	if msg.Type() != s.nextExpectedMessage.Type() {
		return errIncorrectMessageType
	}

	return nil
}

func (s *swapState) handleSendKeysMessage(msg *net.SendKeysMessage) (net.Message, error) {
	if msg.ProvidedAmount < s.info.ReceivedAmount() {
		return nil, fmt.Errorf("receiving amount is not the same as expected: got %v, expected %v",
			msg.ProvidedAmount,
			s.info.ReceivedAmount(),
		)
	}

	if msg.PublicSpendKey == "" || msg.PrivateViewKey == "" {
		return nil, errMissingKeys
	}

	if msg.EthAddress == "" {
		return nil, errMissingAddress
	}

	vk, err := mcrypto.NewPrivateViewKeyFromHex(msg.PrivateViewKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate XMRMaker's private view keys: %w", err)
	}

	s.xmrmakerAddress = ethcommon.HexToAddress(msg.EthAddress)

	log.Debugf("got XMRMaker's keys and address: address=%s", s.xmrmakerAddress)

	sk, err := mcrypto.NewPublicKeyFromHex(msg.PublicSpendKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate XMRMaker's public spend key: %w", err)
	}

	// verify counterparty's DLEq proof and ensure the resulting secp256k1 key is correct
	secp256k1Pub, err := pcommon.VerifyKeysAndProof(msg.DLEqProof, msg.Secp256k1PublicKey)
	if err != nil {
		return nil, err
	}

	log.Infof(color.New(color.Bold).Sprintf("receiving %v XMR for %v ETH", msg.ProvidedAmount, s.info.ProvidedAmount()))

	s.setXMRMakerKeys(sk, vk, secp256k1Pub)
	txHash, err := s.lockETH(s.providedAmountInWei())
	if err != nil {
		return nil, fmt.Errorf("failed to lock ETH in contract: %w", err)
	}

	log.Info("locked ether in swap contract, waiting for XMR to be locked")

	// start goroutine to check that XMRMaker locks before t_0
	go func() {
		// TODO: this variable is so that we definitely refund before t0.
		// Current algorithm is to trigger the timeout when only 15% of the allotted
		// time is remaining. If the block interval is 1 second on a test network and
		// and T0 is 7 seconds after swap creation, we need the refund to trigger more
		// than one second before the block with a timestamp exactly equal to T0 to
		// satisfy the strictly less than requirement. 7s * 15% = 1.05s. 15% remaining
		// may be reasonable even with large timeouts on production networks, but more
		// research is needed.
		t0Delta := s.t1.Sub(s.t0) // time between swap start and T0 is equal to T1-T0
		deltaBeforeT0ToGiveUp := time.Duration(float64(t0Delta) * 0.15)
		deltaUntilGiveUp := time.Until(s.t0) - deltaBeforeT0ToGiveUp
		giveUpAndRefundTimer := time.NewTimer(deltaUntilGiveUp)
		defer giveUpAndRefundTimer.Stop() // don't wait for the timeout to garbage collect
		log.Debugf("time until refund: %vs", deltaUntilGiveUp.Seconds())

		select {
		case <-s.ctx.Done():
			return
		case <-s.xmrLockedCh:
			return
		case <-giveUpAndRefundTimer.C:
			s.lockState()
			defer s.unlockState()

			if !s.info.Status().IsOngoing() {
				return
			}

			// XMRMaker hasn't locked yet, let's call refund
			txhash, err := s.refund()
			if err != nil {
				if !strings.Contains(err.Error(), revertSwapCompleted) {
					log.Errorf("failed to refund: err=%s", err)
				} else {
					log.Debugf("failed to refund (okay): err=%s", err)
				}
				return
			}

			log.Infof("got our ETH back: tx hash=%s", txhash)

			// send NotifyRefund msg
			if err := s.SendSwapMessage(&message.NotifyRefund{
				TxHash: txhash.String(),
			}, s.ID()); err != nil {
				log.Errorf("failed to send refund message: err=%s", err)
			}
		}
	}()

	s.setNextExpectedMessage(&message.NotifyXMRLock{})

	out := &message.NotifyETHLocked{
		Address:        s.ContractAddr().String(),
		TxHash:         txHash.String(),
		ContractSwapID: s.contractSwapID,
		ContractSwap:   pcommon.ConvertContractSwapToMsg(s.contractSwap),
	}

	return out, nil
}

func (s *swapState) handleNotifyXMRLock(msg *message.NotifyXMRLock) (net.Message, error) {
	if msg.Address == "" {
		return nil, errNoLockedXMRAddress
	}

	// check that XMR was locked in expected account, and confirm amount
	vk := mcrypto.SumPrivateViewKeys(s.xmrmakerPrivateViewKey, s.privkeys.ViewKey())
	sk := mcrypto.SumPublicKeys(s.xmrmakerPublicSpendKey, s.pubkeys.SpendKey())
	kp := mcrypto.NewPublicKeyPair(sk, vk.Public())

	if msg.Address != string(kp.Address(s.Env())) {
		return nil, fmt.Errorf("address received in message does not match expected address")
	}

	s.LockClient()
	defer s.UnlockClient()

	t := time.Now().Format(common.TimeFmtNSecs)
	walletName := fmt.Sprintf("xmrtaker-viewonly-wallet-%s", t)
	if err := s.GenerateViewOnlyWalletFromKeys(vk, kp.Address(s.Env()), walletName, ""); err != nil {
		return nil, fmt.Errorf("failed to generate view-only wallet to verify locked XMR: %w", err)
	}

	log.Debugf("generated view-only wallet to check funds: %s", walletName)

	if s.Env() != common.Development {
		log.Infof("waiting for new blocks...")
		// wait for 2 new blocks, otherwise balance might be 0
		// TODO: check transaction hash
		height, err := monero.WaitForBlocks(s.Backend, 2)
		if err != nil {
			return nil, err
		}

		log.Infof("monero block height: %d", height)
	}

	log.Debug("refreshing client...")

	if err := s.Refresh(); err != nil {
		return nil, fmt.Errorf("failed to refresh client: %w", err)
	}

	accounts, err := s.GetAccounts()
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	var (
		balance *monero.GetBalanceResponse
	)

	for i, acc := range accounts.SubaddressAccounts {
		addr, ok := acc["base_address"].(string)
		if !ok {
			panic("address is not a string!")
		}

		if mcrypto.Address(addr) == kp.Address(s.Env()) {
			balance, err = s.GetBalance(uint(i))
			if err != nil {
				return nil, fmt.Errorf("failed to get balance: %w", err)
			}

			break
		}
	}

	if balance == nil {
		return nil, fmt.Errorf("failed to find account with address %s", kp.Address(s.Env()))
	}

	log.Debugf("checking locked wallet, address=%s balance=%v", kp.Address(s.Env()), balance.Balance)

	// TODO: also check that the balance isn't unlocked only after an unreasonable amount of blocks
	if balance.Balance < float64(s.receivedAmountInPiconero()) {
		return nil, fmt.Errorf("locked XMR amount is less than expected: got %v, expected %v",
			balance.Balance, float64(s.receivedAmountInPiconero()))
	}

	if err := s.CloseWallet(); err != nil {
		return nil, fmt.Errorf("failed to close wallet: %w", err)
	}

	close(s.xmrLockedCh)
	log.Info("XMR was locked successfully, setting contract to ready...")

	if err := s.ready(); err != nil {
		return nil, fmt.Errorf("failed to call Ready: %w", err)
	}

	go s.runT1ExpirationHandler()

	s.setNextExpectedMessage(&message.NotifyClaimed{})
	return &message.NotifyReady{}, nil
}

func (s *swapState) runT1ExpirationHandler() {
	log.Debugf("time until t1 (%s): %vs",
		s.t0.Format(common.TimeFmtSecs),
		time.Until(s.t1).Seconds(),
	)

	waitCtx, waitCtxCancel := context.WithCancel(context.Background())
	defer waitCtxCancel() // Unblock WaitForTimestamp if still running when we exit

	waitCh := make(chan error)
	go func() {
		waitCh <- s.WaitForTimestamp(waitCtx, s.t1)
		close(waitCh)
	}()

	select {
	case <-s.ctx.Done():
		return
	case <-s.claimedCh:
		return
	case err := <-waitCh:
		if err != nil {
			// TODO: Do we propagate this error? If we retry, the logic should probably be inside WaitForTimestamp.
			log.Errorf("Failure waiting for T1 timeout: err=%s", err)
			return
		}
		s.handleT1Expired()
	}
}

func (s *swapState) handleT1Expired() {
	s.lockState()
	defer s.unlockState()

	if !s.info.Status().IsOngoing() {
		return
	}

	// XMRMaker hasn't claimed, and we're after t_1. let's call Refund
	txhash, err := s.refund()
	if err != nil {
		log.Errorf("failed to refund: err=%s", err)
		return
	}

	log.Infof("got our ETH back: tx hash=%s", txhash)
	s.clearNextExpectedMessage(types.CompletedRefund) // TODO: duplicate?

	// send NotifyRefund msg
	if err = s.SendSwapMessage(&message.NotifyRefund{
		TxHash: txhash.String(),
	}, s.ID()); err != nil {
		log.Errorf("failed to send refund message: err=%s", err)
	}

	if err = s.exit(); err != nil {
		log.Errorf("exit failed: err=%s", err)
	}
}

// handleNotifyClaimed handles XMRMaker's reveal after he calls Claim().
// it calls `createMoneroWallet` to create XMRTaker's wallet, allowing her to own the XMR.
func (s *swapState) handleNotifyClaimed(txHash string) (mcrypto.Address, error) {
	log.Debugf("got NotifyClaimed, txHash=%s", txHash)
	receipt, err := s.WaitForReceipt(s.ctx, ethcommon.HexToHash(txHash))
	if err != nil {
		return "", fmt.Errorf("failed check claim transaction receipt: %w", err)
	}

	if len(receipt.Logs) == 0 {
		return "", errClaimTxHasNoLogs
	}

	log.Infof("counterparty claimed ETH; tx hash=%s", txHash)

	skB, err := swapfactory.GetSecretFromLog(receipt.Logs[0], "Claimed")
	if err != nil {
		return "", fmt.Errorf("failed to get secret from log: %w", err)
	}

	return s.claimMonero(skB)
}
