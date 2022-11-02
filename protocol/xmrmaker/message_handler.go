package xmrmaker

import (
	"context"
	"fmt"
	"reflect"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/net/message"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
)

// HandleProtocolMessage is called by the network to handle an incoming message.
// If the message received is not the expected type for the point in the protocol we're at,
// this function will return an error.
func (s *swapState) HandleProtocolMessage(msg net.Message) error {
	if s == nil {
		return errNilSwapState
	}

	// s.lockState()
	// defer s.unlockState()

	if s.ctx.Err() != nil {
		return fmt.Errorf("protocol exited: %w", s.ctx.Err())
	}

	// if err := s.checkMessageType(msg); err != nil {
	// 	return err
	// }

	switch msg := msg.(type) {
	case *net.SendKeysMessage:
		event := newEventKeysSent(msg)
		s.eventCh <- event
		err := <-event.errCh
		if err != nil {
			return err
		}
	case *message.NotifyETHLocked:
		event := newEventETHLocked(msg)
		s.eventCh <- event
		err := <-event.errCh
		if err != nil {
			return err
		}

		// TODO we can actually close the network stream after sending the
		// XMRLocked message

	// case *message.NotifyReady:
	// 	log.Debug("contract ready, attempting to claim funds...")
	// 	close(s.readyCh)

	// 	// contract ready, let's claim our ether
	// 	txHash, err := s.claimFunds()
	// 	if err != nil {
	// 		return fmt.Errorf("failed to redeem ether: %w", err)
	// 	}

	// 	log.Debug("funds claimed!!")
	// 	resp := &message.NotifyClaimed{
	// 		TxHash: txHash.String(),
	// 	}

	// 	s.clearNextExpectedMessage(types.CompletedSuccess)
	// 	return s.SendSwapMessage(resp, s.ID())
	// case *message.NotifyRefund:
	// 	// generate monero wallet, regaining control over locked funds
	// 	addr, err := s.handleRefund(msg.TxHash)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	s.clearNextExpectedMessage(types.CompletedRefund)
	// 	log.Infof("regained control over monero account %s", addr)
	// 	s.CloseProtocolStream(s.ID())
	// 	return nil
	default:
		return errUnexpectedMessageType
	}

	return nil
}

func (s *swapState) clearNextExpectedEvent(status types.Status) {
	s.nextExpectedEvent = nil
	s.info.SetStatus(status)
	if s.offerExtra.StatusCh != nil {
		s.offerExtra.StatusCh <- status
	}
}

func (s *swapState) setNextExpectedEvent(event Event) {
	// if s == nil {
	// 	return
	// }

	// TODO is event ever nil?
	if event == nil || s.nextExpectedEvent == nil {
		return
	}

	// TODO test this!!!
	// alternatively make a Type() method for the Event interface
	// can also change nextExpectedEvent to EventType
	if reflect.TypeOf(event) == reflect.TypeOf(s.nextExpectedEvent) {
		return
	}

	s.nextExpectedEvent = event
	status := getStatus(event)
	if s.offerExtra.StatusCh != nil && status != types.UnknownStatus {
		s.offerExtra.StatusCh <- status
	}
}

// func (s *swapState) checkMessageType(msg net.Message) error {
// 	if msg == nil {
// 		return errNilMessage
// 	}

// 	if s == nil || s.nextExpectedEvent == nil {
// 		return nil
// 	}

// 	// TODO
// 	// // XMRTaker might refund anytime before t0 or after t1, so we should allow this.
// 	// if _, ok := msg.(*message.NotifyRefund); ok {
// 	// 	return nil
// 	// }

// 	if msg.Type() != s.nextExpectedMessage.Type() {
// 		return errIncorrectMessageType
// 	}

// 	return nil
// }

func (s *swapState) handleNotifyETHLocked(msg *message.NotifyETHLocked) (net.Message, error) {
	if msg.Address == "" {
		return nil, errMissingAddress
	}

	if msg.ContractSwapID.IsZero() {
		return nil, errNilContractSwapID
	}

	log.Infof("got NotifyETHLocked; address=%s contract swap ID=%s", msg.Address, msg.ContractSwapID)

	// validate that swap ID == keccak256(swap struct)
	if err := checkContractSwapID(msg); err != nil {
		return nil, err
	}

	s.contractSwapID = msg.ContractSwapID
	s.contractSwap = convertContractSwap(msg.ContractSwap)

	if err := pcommon.WriteContractSwapToFile(s.offerExtra.InfoFile, s.contractSwapID, s.contractSwap); err != nil {
		return nil, err
	}

	contractAddr := ethcommon.HexToAddress(msg.Address)
	if _, err := contracts.CheckSwapFactoryContractCode(s.ctx, s.Backend.EthClient(), contractAddr); err != nil {
		return nil, err
	}

	if err := s.setContract(contractAddr); err != nil {
		return nil, fmt.Errorf("failed to instantiate contract instance: %w", err)
	}

	if err := pcommon.WriteContractAddressToFile(s.offerExtra.InfoFile, msg.Address); err != nil {
		return nil, fmt.Errorf("failed to write contract address to file: %w", err)
	}

	if err := s.checkContract(ethcommon.HexToHash(msg.TxHash)); err != nil {
		return nil, err
	}

	// TODO: check these (in checkContract) (#161)
	s.setTimeouts(msg.ContractSwap.Timeout0, msg.ContractSwap.Timeout1)

	notifyXMRLocked, err := s.lockFunds(common.MoneroToPiconero(s.info.ProvidedAmount()))
	if err != nil {
		return nil, fmt.Errorf("failed to lock funds: %w", err)
	}

	go s.runT0ExpirationHandler()

	// TODO: set next expected event
	//s.setNextExpectedMessage(&message.NotifyReady{})
	return notifyXMRLocked, nil
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
			// TODO: Do we propagate this error? If we retry, the logic should probably be inside
			// WaitForTimestamp. (#162)
			log.Errorf("Failure waiting for T0 timeout: err=%s", err)
			return
		}
		s.handleT0Expired()
	}
}

func (s *swapState) handleT0Expired() {
	// s.lockState()
	// defer s.unlockState()

	// TODO this probably shouldn't happen anymore since we're event-driven
	if !s.info.Status().IsOngoing() {
		// swap was already completed, just return
		return
	}

	event := newEventContractReady()
	s.eventCh <- event
	err := <-event.errCh
	if err != nil {
		// TODO this is quite bad?
		log.Errorf("failed to handle t0 expiration: %s", err)
	}

	// // we can now call Claim()
	// txHash, err := s.claimFunds()
	// if err != nil {
	// 	log.Errorf("failed to claim: err=%s", err)
	// 	// TODO: retry claim, depending on error (#162)
	// 	if err = s.exit(); err != nil {
	// 		log.Errorf("exit failed: err=%s", err)
	// 	}
	// 	return
	// }

	// log.Debug("funds claimed!")
	// s.clearNextExpectedEvent(types.CompletedSuccess)

	// // send *message.NotifyClaimed
	// notifyClaimed := &message.NotifyClaimed{TxHash: txHash.String()}
	// if err := s.SendSwapMessage(notifyClaimed, s.ID()); err != nil {
	// 	log.Errorf("failed to send NotifyClaimed message: err=%s", err)
	// }
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
	s.setNextExpectedEvent(&EventETHLocked{})
	return nil
}
