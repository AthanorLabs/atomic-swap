package xmrmaker

import (
	"context"
	"fmt"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/net/message"
	pcommon "github.com/noot/atomic-swap/protocol"
	"github.com/noot/atomic-swap/swapfactory"
)

// HandleProtocolMessage is called by the network to handle an incoming message.
// If the message received is not the expected type for the point in the protocol we're at,
// this function will return an error.
func (s *swapState) HandleProtocolMessage(msg net.Message) (net.Message, bool, error) {
	if s == nil {
		return nil, true, errNilSwapState
	}

	s.lockState()
	defer s.unlockState()

	if s.ctx.Err() != nil {
		return nil, true, fmt.Errorf("protocol exited: %w", s.ctx.Err())
	}

	if err := s.checkMessageType(msg); err != nil {
		return nil, true, err
	}

	switch msg := msg.(type) {
	case *net.SendKeysMessage:
		if err := s.handleSendKeysMessage(msg); err != nil {
			return nil, true, err
		}

		return nil, false, nil
	case *message.NotifyETHLocked:
		out, err := s.handleNotifyETHLocked(msg)
		if err != nil {
			return nil, true, err
		}

		return out, false, nil
	case *message.NotifyReady:
		log.Debug("contract ready, attempting to claim funds...")
		close(s.readyCh)

		// contract ready, let's claim our ether
		txHash, err := s.claimFunds()
		if err != nil {
			return nil, true, fmt.Errorf("failed to redeem ether: %w", err)
		}

		log.Debug("funds claimed!!")
		out := &message.NotifyClaimed{
			TxHash: txHash.String(),
		}

		s.clearNextExpectedMessage(types.CompletedSuccess)
		return out, true, nil
	case *message.NotifyRefund:
		// generate monero wallet, regaining control over locked funds
		addr, err := s.handleRefund(msg.TxHash)
		if err != nil {
			return nil, false, err
		}

		s.clearNextExpectedMessage(types.CompletedRefund)
		log.Infof("regained control over monero account %s", addr)
		return nil, true, nil
	default:
		return nil, true, errUnexpectedMessageType
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
	if s == nil {
		return
	}

	if msg == nil || s.nextExpectedMessage == nil {
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

	if s == nil || s.nextExpectedMessage == nil {
		return nil
	}

	// XMRTaker might refund anytime before t0 or after t1, so we should allow this.
	if _, ok := msg.(*message.NotifyRefund); ok {
		return nil
	}

	if msg.Type() != s.nextExpectedMessage.Type() {
		return errIncorrectMessageType
	}

	return nil
}

func (s *swapState) handleNotifyETHLocked(msg *message.NotifyETHLocked) (net.Message, error) {
	if msg.Address == "" {
		return nil, errMissingAddress
	}

	if msg.ContractSwapID == [32]byte{} {
		return nil, errNilContractSwapID
	}

	log.Infof("got NotifyETHLocked; address=%s contract swap ID=%x", msg.Address, msg.ContractSwapID)

	// validate that swap ID == keccak256(swap struct)
	if err := checkContractSwapID(msg); err != nil {
		return nil, err
	}

	s.contractSwapID = msg.ContractSwapID
	s.contractSwap = convertContractSwap(msg.ContractSwap)

	if err := pcommon.WriteContractSwapToFile(s.infoFile, s.contractSwapID, s.contractSwap); err != nil {
		return nil, err
	}

	contractAddr := ethcommon.HexToAddress(msg.Address)
	if err := checkContractCode(s.ctx, s, contractAddr); err != nil {
		return nil, err
	}

	if err := s.setContract(contractAddr); err != nil {
		return nil, fmt.Errorf("failed to instantiate contract instance: %w", err)
	}

	if err := pcommon.WriteContractAddressToFile(s.infoFile, msg.Address); err != nil {
		return nil, fmt.Errorf("failed to write contract address to file: %w", err)
	}

	if err := s.checkContract(ethcommon.HexToHash(msg.TxHash)); err != nil {
		return nil, err
	}

	// TODO: check these (in checkContract)
	s.setTimeouts(msg.ContractSwap.Timeout0, msg.ContractSwap.Timeout1)

	addrAB, err := s.lockFunds(common.MoneroToPiconero(s.info.ProvidedAmount()))
	if err != nil {
		return nil, fmt.Errorf("failed to lock funds: %w", err)
	}

	out := &message.NotifyXMRLock{
		Address: string(addrAB),
	}

	go s.runT0ExpirationHandler()

	s.setNextExpectedMessage(&message.NotifyReady{})
	return out, nil
}

func (s *swapState) runT0ExpirationHandler() {
	log.Debugf("time until t0 (%s): %vs",
		s.t0.Format(common.TimeFmtSecs),
		time.Until(s.t0).Seconds(),
	)

	waitCtx, waitCtxCancel := context.WithCancel(context.Background())
	defer waitCtxCancel() // Unblock WaitForTimestamp if still running when we exit

	waitCh := make(chan error)
	go func() {
		waitCh <- s.WaitForTimestamp(waitCtx, s.t0)
		close(waitCh)
	}()

	select {
	case <-s.ctx.Done():
		return
	case <-s.readyCh:
		return
	case err := <-waitCh:
		if err != nil {
			// TODO: Do we propagate this error? If we retry, the logic should probably be inside WaitForTimestamp.
			log.Errorf("Failure waiting for T0 timeout: err=%s", err)
			return
		}
		s.handleT0Expired()
	}
}

func (s *swapState) handleT0Expired() {
	s.lockState()
	defer s.unlockState()

	if !s.info.Status().IsOngoing() {
		// swap was already completed, just return
		return
	}

	// we can now call Claim()
	txHash, err := s.claimFunds()
	if err != nil {
		log.Errorf("failed to claim: err=%s", err)
		// TODO: retry claim, depending on error
		if err = s.exit(); err != nil {
			log.Errorf("exit failed: err=%s", err)
		}
		return
	}

	log.Debug("funds claimed!")
	s.clearNextExpectedMessage(types.CompletedSuccess)

	// send *message.NotifyClaimed
	notifyClaimed := &message.NotifyClaimed{TxHash: txHash.String()}
	if err := s.SendSwapMessage(notifyClaimed, s.ID()); err != nil {
		log.Errorf("failed to send NotifyClaimed message: err=%s", err)
	}
}

func (s *swapState) handleSendKeysMessage(msg *net.SendKeysMessage) error {
	if msg.PublicSpendKey == "" || msg.PublicViewKey == "" {
		return errMissingKeys
	}

	kp, err := mcrypto.NewPublicKeyPairFromHex(msg.PublicSpendKey, msg.PublicViewKey)
	if err != nil {
		return fmt.Errorf("failed to generate XMRTaker's public keys: %w", err)
	}

	// verify counterparty's DLEq proof and ensure the resulting secp256k1 key is correct
	secp256k1Pub, err := pcommon.VerifyKeysAndProof(msg.DLEqProof, msg.Secp256k1PublicKey)
	if err != nil {
		return err
	}

	s.setXMRTakerPublicKeys(kp, secp256k1Pub)
	s.setNextExpectedMessage(&message.NotifyETHLocked{})
	return nil
}

func (s *swapState) handleRefund(txHash string) (mcrypto.Address, error) {
	receipt, err := s.TransactionReceipt(s.ctx, ethcommon.HexToHash(txHash))
	if err != nil {
		return "", err
	}

	if len(receipt.Logs) == 0 {
		return "", errClaimTxHasNoLogs
	}

	sa, err := swapfactory.GetSecretFromLog(receipt.Logs[0], "Refunded")
	if err != nil {
		return "", err
	}

	return s.reclaimMonero(sa)
}
