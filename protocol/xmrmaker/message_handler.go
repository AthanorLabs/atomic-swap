package xmrmaker

import (
	"context"
	"fmt"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/db"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
)

// HandleProtocolMessage is called by the network to handle an incoming message.
// If the message received is not the expected type for the point in the protocol we're at,
// this function will return an error.
func (s *swapState) HandleProtocolMessage(msg message.Message) error {
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

		// TODO: we can actually close the network stream after
		// sending the XMRLocked message, but since the network
		// calls Exit() when the stream closes, it needs to not
		// do that in this case.
	default:
		return errUnexpectedMessageType
	}

	return nil
}

func (s *swapState) clearNextExpectedEvent(status types.Status) {
	s.nextExpectedEvent = EventNoneType
	s.info.SetStatus(status)
	if s.offerExtra.StatusCh != nil {
		s.offerExtra.StatusCh <- status
	}
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

	if s.offerExtra.StatusCh != nil {
		s.offerExtra.StatusCh <- status
	}

	return nil
}

func (s *swapState) handleNotifyETHLocked(msg *message.NotifyETHLocked) (message.Message, error) {
	if msg.Address == "" {
		return nil, errMissingAddress
	}

	if types.IsHashZero(msg.ContractSwapID) {
		return nil, errNilContractSwapID
	}

	log.Infof("got NotifyETHLocked; address=%s contract swap ID=%s", msg.Address, msg.ContractSwapID)

	// validate that swap ID == keccak256(swap struct)
	if err := checkContractSwapID(msg); err != nil {
		return nil, err
	}

	s.contractSwapID = msg.ContractSwapID
	s.contractSwap = convertContractSwap(msg.ContractSwap)

	receipt, err := s.Backend.ETHClient().Raw().TransactionReceipt(s.ctx, ethcommon.HexToHash(msg.TxHash))
	if err != nil {
		return nil, err
	}

	contractAddr := ethcommon.HexToAddress(msg.Address)
	// note: this function verifies the forwarder code as well, even if we aren't using a relayer,
	// in which case it's not relevant to us and we don't need to verify it.
	// doesn't hurt though I suppose.
	_, err = contracts.CheckSwapFactoryContractCode(s.ctx, s.Backend.ETHClient().Raw(), contractAddr)
	if err != nil {
		return nil, err
	}

	if err = s.setContract(contractAddr); err != nil {
		return nil, fmt.Errorf("failed to instantiate contract instance: %w", err)
	}

	ethInfo := &db.EthereumSwapInfo{
		StartNumber:     receipt.BlockNumber,
		SwapID:          s.contractSwapID,
		Swap:            s.contractSwap,
		ContractAddress: contractAddr,
	}

	if err = s.Backend.RecoveryDB().PutContractSwapInfo(s.ID(), ethInfo); err != nil {
		return nil, err
	}

	if err = s.checkContract(ethcommon.HexToHash(msg.TxHash)); err != nil {
		return nil, err
	}

	err = s.checkAndSetTimeouts(msg.ContractSwap.Timeout0, msg.ContractSwap.Timeout1)
	if err != nil {
		return nil, err
	}

	notifyXMRLocked, err := s.lockFunds(coins.MoneroToPiconero(s.info.ProvidedAmount))
	if err != nil {
		return nil, fmt.Errorf("failed to lock funds: %w", err)
	}

	go s.runT0ExpirationHandler()
	return notifyXMRLocked, nil
}

func (s *swapState) runT0ExpirationHandler() {
	log.Debugf("time until t0 (%s): %vs",
		s.t0.Format(common.TimeFmtSecs),
		time.Until(s.t0).Seconds(),
	)

	waitCtx, waitCtxCancel := context.WithCancel(context.Background())
	defer waitCtxCancel() // Unblock WaitForTimestamp if still running when we exit

	// note: this will cause unit tests to hang if not running ganache
	// with --miner.blockTime!!!
	waitCh := make(chan error)
	go func() {
		waitCh <- s.ETHClient().WaitForTimestamp(waitCtx, s.t0)
		close(waitCh)
	}()

	select {
	case <-s.ctx.Done():
		return
	case <-s.readyCh:
		log.Debugf("returning from runT0ExpirationHandler as contract was set to ready")
		return
	case err := <-waitCh:
		if err != nil {
			// TODO: Do we propagate this error? If we retry, the logic should probably be inside
			// WaitForTimestamp. (#162)
			log.Errorf("Failure waiting for T0 timeout: err=%s", err)
			return
		}
		log.Debugf("reached t0, time to claim")
		s.handleT0Expired()
	}
}

func (s *swapState) handleT0Expired() {
	event := newEventContractReady()
	s.eventCh <- event
	err := <-event.errCh
	if err != nil {
		// TODO: this is quite bad, how should this be handled? (#162)
		log.Errorf("failed to handle t0 expiration: %s", err)
	}
}

func (s *swapState) handleSendKeysMessage(msg *message.SendKeysMessage) error {
	if msg.PublicSpendKey == "" || msg.PublicViewKey == "" {
		return errMissingKeys
	}

	// verify counterparty's DLEq proof and ensure the resulting secp256k1 key is correct
	verifyResult, err := pcommon.VerifyKeysAndProof(msg.DLEqProof, msg.Secp256k1PublicKey, msg.PublicSpendKey)
	if err != nil {
		return err
	}

	kp, err := mcrypto.NewPublicKeyPairFromHex(msg.PublicSpendKey, msg.PublicViewKey)
	if err != nil {
		return fmt.Errorf("failed to generate XMRTaker's public keys: %w", err)
	}

	return s.setXMRTakerPublicKeys(kp, verifyResult.Secp256k1PublicKey)
}
