package bob

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	eth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/fatih/color" //nolint:misspell

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero"
	mcrypto "github.com/noot/atomic-swap/monero/crypto"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/swap-contract"
	"github.com/noot/atomic-swap/types"
)

var nextID uint64

var (
	errMissingKeys       = errors.New("did not receive Alice's public spend or view key")
	errMissingAddress    = errors.New("got empty contract address")
	errNoRefundLogsFound = errors.New("no refund logs found")
	errPastClaimTime     = errors.New("past t1, can no longer claim")
)

var (
	// this is from the autogenerated swap.go
	// TODO: generate this ourselves instead of hard-coding
	refundedTopic = ethcommon.HexToHash("0xfe509803c09416b28ff3d8f690c8b0c61462a892c46d5430c8fb20abe472daf0")
)

type swapState struct {
	bob    *Instance
	ctx    context.Context
	cancel context.CancelFunc
	sync.Mutex

	id             uint64
	providesAmount common.MoneroAmount
	desiredAmount  common.EtherAmount
	offerID        types.Hash

	// our keys for this session
	privkeys *mcrypto.PrivateKeyPair
	pubkeys  *mcrypto.PublicKeyPair

	// swap contract and timeouts in it; set once contract is deployed
	contract     *swap.Swap
	contractAddr ethcommon.Address
	t0, t1       time.Time
	txOpts       *bind.TransactOpts

	// Alice's keys for this session
	alicePublicKeys *mcrypto.PublicKeyPair

	// next expected network message
	nextExpectedMessage net.Message

	// channels
	readyCh chan struct{}

	// set to true on claiming the ETH or reclaiming XMR
	completed            bool
	refunded             bool
	moneroReclaimAddress mcrypto.Address
}

func newSwapState(b *Instance, offerID types.Hash, providesAmount common.MoneroAmount, desiredAmount common.EtherAmount) (*swapState, error) { //nolint:lll
	txOpts, err := bind.NewKeyedTransactorWithChainID(b.ethPrivKey, b.chainID)
	if err != nil {
		return nil, err
	}

	txOpts.GasPrice = b.gasPrice
	txOpts.GasLimit = b.gasLimit

	ctx, cancel := context.WithCancel(b.ctx)

	s := &swapState{
		ctx:                 ctx,
		cancel:              cancel,
		bob:                 b,
		id:                  nextID,
		providesAmount:      providesAmount,
		desiredAmount:       desiredAmount,
		offerID:             offerID,
		nextExpectedMessage: &net.SendKeysMessage{},
		readyCh:             make(chan struct{}),
		txOpts:              txOpts,
	}

	nextID++
	return s, nil
}

// SendKeysMessage ...
func (s *swapState) SendKeysMessage() (*net.SendKeysMessage, error) {
	sk, vk, err := s.generateKeys()
	if err != nil {
		return nil, err
	}

	sig, err := s.privkeys.SpendKey().Sign(s.pubkeys.SpendKey().Bytes())
	if err != nil {
		return nil, err
	}

	return &net.SendKeysMessage{
		ProvidedAmount:  s.providesAmount.AsMonero(),
		PublicSpendKey:  sk.Hex(),
		PrivateViewKey:  vk.Hex(),
		PrivateKeyProof: sig.Hex(),
		EthAddress:      s.bob.ethAddress.String(),
	}, nil
}

// ReceivedAmount returns the amount received, or expected to be received, at the end of the swap
func (s *swapState) ReceivedAmount() float64 {
	return s.desiredAmount.AsEther()
}

// ProtocolExited is called by the network when the protocol stream closes.
// If it closes prematurely, we need to perform recovery.
func (s *swapState) ProtocolExited() error {
	s.Lock()
	defer s.Unlock()

	defer func() {
		// stop all running goroutines
		s.cancel()
		s.bob.swapState = nil
	}()

	// TODO: defer this?
	if s.completed {
		str := color.New(color.Bold).Sprintf("**swap completed successfully! id=%d**", s.id)
		log.Info(str)

		if !s.refunded {
			// remove offer, as it's been taken
			s.bob.offerManager.deleteOffer(s.offerID)
		}

		return nil
	}

	switch s.nextExpectedMessage.(type) {
	case *net.SendKeysMessage:
		// we are fine, as we only just initiated the protocol.
		return errors.New("protocol exited before any funds were locked")
	case *net.NotifyContractDeployed:
		// we were waiting for the contract to be deployed, but haven't
		// locked out funds yet, so we're fine.
		return errors.New("protocol exited before any funds were locked")
	case *net.NotifyReady:
		// we already locked our funds - need to wait until we can claim
		// the funds (ie. wait until after t0)
		txHash, err := s.tryClaim()
		if err != nil {
			log.Errorf("failed to claim funds: err=%s", err)
		} else {
			log.Infof("claimed ether! transaction hash=%s", txHash)
			return nil
		}

		// we should check if Alice refunded, if so then check contract for secret
		address, err := s.tryReclaimMonero()
		if err != nil {
			log.Errorf("failed to check for refund: err=%s", err)
			// TODO: keep retrying until success
			return err
		}

		s.completed = true
		s.refunded = true // TODO: return this
		s.moneroReclaimAddress = address
		log.Infof("regained private key to monero wallet, address=%s", address)
		return nil
	default:
		log.Errorf("unexpected nextExpectedMessage in ProtocolExited: type=%T", s.nextExpectedMessage)
		return errors.New("unexpected message type")
	}
}

func (s *swapState) tryReclaimMonero() (mcrypto.Address, error) {
	skA, err := s.filterForRefund()
	if err != nil {
		return "", err
	}

	return s.reclaimMonero(skA)
}

func (s *swapState) reclaimMonero(skA *mcrypto.PrivateSpendKey) (mcrypto.Address, error) {
	vkA, err := skA.View()
	if err != nil {
		return "", err
	}

	skAB := mcrypto.SumPrivateSpendKeys(skA, s.privkeys.SpendKey())
	vkAB := mcrypto.SumPrivateViewKeys(vkA, s.privkeys.ViewKey())
	kpAB := mcrypto.NewPrivateKeyPair(skAB, vkAB)

	// write keys to file in case something goes wrong
	fp := fmt.Sprintf("%s/%d/swap-secret", s.bob.basepath, s.id)
	if err = mcrypto.WriteKeysToFile(fp, kpAB, s.bob.env); err != nil {
		return "", err
	}

	// TODO: check balance
	return monero.CreateMoneroWallet("bob-swap-wallet", s.bob.env, s.bob.client, kpAB)
}

func (s *swapState) filterForRefund() (*mcrypto.PrivateSpendKey, error) {
	logs, err := s.bob.ethClient.FilterLogs(s.ctx, eth.FilterQuery{
		Addresses: []ethcommon.Address{s.contractAddr},
		Topics:    [][]ethcommon.Hash{{refundedTopic}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %w", err)
	}

	if len(logs) == 0 {
		return nil, errNoRefundLogsFound
	}

	sa, err := swap.GetSecretFromLog(&logs[0], "Refunded")
	if err != nil {
		return nil, fmt.Errorf("failed to get secret from log: %w", err)
	}

	return sa, nil
}

func (s *swapState) tryClaim() (ethcommon.Hash, error) {
	untilT0 := time.Until(s.t0)
	untilT1 := time.Until(s.t1)

	if untilT0 < 0 {
		// we need to wait until t0 to claim
		log.Infof("waiting until time %s to claim", s.t0)
		<-time.After(untilT0 + time.Second)
	}

	if untilT1 > 0 {
		// we've passed t1, our only option now is for Alice to refund
		// and we can regain control of the locked XMR.
		return ethcommon.Hash{}, errPastClaimTime
	}

	return s.claimFunds()
}

// HandleProtocolMessage is called by the network to handle an incoming message.
// If the message received is not the expected type for the point in the protocol we're at,
// this function will return an error.
func (s *swapState) HandleProtocolMessage(msg net.Message) (net.Message, bool, error) {
	s.Lock()
	defer s.Unlock()

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

		// we initiated, so we're now waiting for Alice to deploy the contract.
		return nil, false, nil
	case *net.NotifyContractDeployed:
		if msg.Address == "" {
			return nil, true, errMissingAddress
		}

		log.Infof("got Swap contract address! address=%s", msg.Address)

		if err := s.setContract(ethcommon.HexToAddress(msg.Address)); err != nil {
			return nil, true, fmt.Errorf("failed to instantiate contract instance: %w", err)
		}

		fp := fmt.Sprintf("%s/%d/contractaddress", s.bob.basepath, s.id)
		if err := common.WriteContractAddressToFile(fp, msg.Address); err != nil {
			return nil, true, fmt.Errorf("failed to write contract address to file: %w", err)
		}

		if err := s.checkContract(); err != nil {
			return nil, true, err
		}

		addrAB, err := s.lockFunds(s.providesAmount)
		if err != nil {
			return nil, true, fmt.Errorf("failed to lock funds: %w", err)
		}

		out := &net.NotifyXMRLock{
			Address: string(addrAB),
		}

		// set t0 and t1
		if err := s.setTimeouts(); err != nil {
			return nil, true, err
		}

		go func() {
			until := time.Until(s.t0)

			log.Debugf("time until t0: %vs", until.Seconds())

			select {
			case <-s.ctx.Done():
				return
			case <-time.After(until + time.Second):
				// we can now call Claim()
				txHash, err := s.claimFunds()
				if err != nil {
					log.Errorf("failed to claim: err=%s", err)
					// TODO: retry claim, depending on error
					return
				}

				log.Debug("funds claimed!")
				s.completed = true

				// send *net.NotifyClaimed
				if err := s.bob.net.SendSwapMessage(&net.NotifyClaimed{
					TxHash: txHash.String(),
				}); err != nil {
					log.Errorf("failed to send NotifyClaimed message: err=%s", err)
				}
			case <-s.readyCh:
				return
			}
		}()

		s.nextExpectedMessage = &net.NotifyReady{}
		return out, false, nil
	case *net.NotifyReady:
		log.Debug("Alice called Ready(), attempting to claim funds...")
		close(s.readyCh)

		// contract ready, let's claim our ether
		txHash, err := s.claimFunds()
		if err != nil {
			return nil, true, fmt.Errorf("failed to redeem ether: %w", err)
		}

		log.Debug("funds claimed!!")
		out := &net.NotifyClaimed{
			TxHash: txHash.String(),
		}

		s.completed = true
		return out, true, nil
	case *net.NotifyRefund:
		// generate monero wallet, regaining control over locked funds
		addr, err := s.handleRefund(msg.TxHash)
		if err != nil {
			return nil, false, err
		}

		s.completed = true
		s.refunded = true
		log.Infof("regained control over monero account %s", addr)
		return nil, true, nil
	default:
		return nil, true, errors.New("unexpected message type")
	}
}

// generateKeys generates Bob's spend and view keys (s_b, v_b)
// It returns Bob's public spend key and his private view key, so that Alice can see
// if the funds are locked.
func (s *swapState) generateKeys() (*mcrypto.PublicKey, *mcrypto.PrivateViewKey, error) {
	if s.privkeys != nil {
		return s.pubkeys.SpendKey(), s.privkeys.ViewKey(), nil
	}

	var err error
	s.privkeys, err = mcrypto.GenerateKeys()
	if err != nil {
		return nil, nil, err
	}

	fp := fmt.Sprintf("%s/%d/bob-secret", s.bob.basepath, s.id)
	if err := mcrypto.WriteKeysToFile(fp, s.privkeys, s.bob.env); err != nil {
		return nil, nil, err
	}

	s.pubkeys = s.privkeys.PublicKeyPair()
	return s.pubkeys.SpendKey(), s.privkeys.ViewKey(), nil
}

// setAlicePublicKeys sets Alice's public spend and view keys
func (s *swapState) setAlicePublicKeys(sk *mcrypto.PublicKeyPair) {
	s.alicePublicKeys = sk
}

// setContract sets the contract in which Alice has locked her ETH.
func (s *swapState) setContract(address ethcommon.Address) error {
	var err error
	s.contractAddr = address
	s.contract, err = swap.NewSwap(address, s.bob.ethClient)
	return err
}

func (s *swapState) setTimeouts() error {
	st0, err := s.contract.Timeout0(s.bob.callOpts)
	if err != nil {
		return fmt.Errorf("failed to get timeout0 from contract: err=%w", err)
	}

	s.t0 = time.Unix(st0.Int64(), 0)

	st1, err := s.contract.Timeout1(s.bob.callOpts)
	if err != nil {
		return fmt.Errorf("failed to get timeout1 from contract: err=%w", err)
	}

	s.t1 = time.Unix(st1.Int64(), 0)
	return nil
}

// checkContract checks the contract's balance and Claim/Refund keys.
// if the balance doesn't match what we're expecting to receive, or the public keys in the contract
// aren't what we expect, we error and abort the swap.
func (s *swapState) checkContract() error {
	balance, err := s.bob.ethClient.BalanceAt(s.ctx, s.contractAddr, nil)
	if err != nil {
		return err
	}

	if balance.Cmp(s.desiredAmount.BigInt()) < 0 {
		return fmt.Errorf("contract does not have expected balance: got %s, expected %s", balance, s.desiredAmount)
	}

	constructedTopic := ethcommon.HexToHash("0x8d36aa70807342c3036697a846281194626fd4afa892356ad5979e03831ab080")
	logs, err := s.bob.ethClient.FilterLogs(s.ctx, eth.FilterQuery{
		Addresses: []ethcommon.Address{s.contractAddr},
		Topics:    [][]ethcommon.Hash{{constructedTopic}},
	})
	if err != nil {
		return fmt.Errorf("failed to filter logs: %w", err)
	}

	if len(logs) == 0 {
		return errors.New("cannot find Constructed log")
	}

	abi, err := abi.JSON(strings.NewReader(swap.SwapABI))
	if err != nil {
		return err
	}

	data := logs[0].Data
	res, err := abi.Unpack("Constructed", data)
	if err != nil {
		return err
	}

	if len(res) < 2 {
		return errors.New("constructed event was missing parameters")
	}

	pkClaim := res[0].([32]byte)
	pkRefund := res[0].([32]byte)

	skOurs := common.Reverse(s.pubkeys.SpendKey().Bytes())
	if !bytes.Equal(pkClaim[:], skOurs) {
		return fmt.Errorf("contract claim key is not expected: got 0x%x, expected 0x%x", pkClaim, skOurs)
	}

	skTheirs := common.Reverse(s.alicePublicKeys.SpendKey().Bytes())
	if !bytes.Equal(pkRefund[:], skOurs) {
		return fmt.Errorf("contract claim key is not expected: got 0x%x, expected 0x%x", pkRefund, skTheirs)
	}

	return nil
}

func (s *swapState) handleSendKeysMessage(msg *net.SendKeysMessage) error {
	if msg.PublicSpendKey == "" || msg.PublicViewKey == "" {
		return errMissingKeys
	}

	log.Debug("got Alice's public keys")

	kp, err := mcrypto.NewPublicKeyPairFromHex(msg.PublicSpendKey, msg.PublicViewKey)
	if err != nil {
		return fmt.Errorf("failed to generate Alice's public keys: %w", err)
	}

	// verify that counterparty really has the private key to the public key
	sig, err := mcrypto.NewSignatureFromHex(msg.PrivateKeyProof)
	if err != nil {
		return err
	}

	ok := kp.SpendKey().Verify(kp.SpendKey().Bytes(), sig)
	if !ok {
		return errors.New("failed to verify proof of private key")
	}

	s.setAlicePublicKeys(kp)
	s.nextExpectedMessage = &net.NotifyContractDeployed{}
	return nil
}

func (s *swapState) handleRefund(txHash string) (mcrypto.Address, error) {
	receipt, err := s.bob.ethClient.TransactionReceipt(s.ctx, ethcommon.HexToHash(txHash))
	if err != nil {
		return "", err
	}

	if len(receipt.Logs) == 0 {
		return "", errors.New("claim transaction has no logs")
	}

	sa, err := swap.GetSecretFromLog(receipt.Logs[0], "Refunded")
	if err != nil {
		return "", err
	}

	return s.reclaimMonero(sa)
}

func (s *swapState) checkMessageType(msg net.Message) error {
	// Alice might refund anytime before t0 or after t1, so we should allow this.
	if _, ok := msg.(*net.NotifyRefund); ok {
		return nil
	}

	if msg.Type() != s.nextExpectedMessage.Type() {
		return errors.New("received unexpected message")
	}

	return nil
}

// lockFunds locks Bob's funds in the monero account specified by public key
// (S_a + S_b), viewable with (V_a + V_b)
// It accepts the amount to lock as the input
// TODO: units
func (s *swapState) lockFunds(amount common.MoneroAmount) (mcrypto.Address, error) {
	kp := mcrypto.SumSpendAndViewKeys(s.alicePublicKeys, s.pubkeys)
	log.Infof("going to lock XMR funds, amount(piconero)=%d", amount)

	balance, err := s.bob.client.GetBalance(0)
	if err != nil {
		return "", err
	}

	log.Debug("total XMR balance: ", balance.Balance)
	log.Info("unlocked XMR balance: ", balance.UnlockedBalance)

	address := kp.Address(s.bob.env)
	txResp, err := s.bob.client.Transfer(address, 0, uint(amount))
	if err != nil {
		return "", err
	}

	log.Infof("locked XMR, txHash=%s fee=%d", txResp.TxHash, txResp.Fee)

	bobAddr, err := s.bob.client.GetAddress(0)
	if err != nil {
		return "", err
	}

	// if we're on a development --regtest node, generate some blocks
	if s.bob.env == common.Development {
		_ = s.bob.daemonClient.GenerateBlocks(bobAddr.Address, 2)
	} else {
		// otherwise, wait for new blocks
		if err := monero.WaitForBlocks(s.bob.client); err != nil {
			return "", err
		}
	}

	if err := s.bob.client.Refresh(); err != nil {
		return "", err
	}

	log.Infof("successfully locked XMR funds: address=%s", address)
	return address, nil
}

// claimFunds redeems Bob's ETH funds by calling Claim() on the contract
func (s *swapState) claimFunds() (ethcommon.Hash, error) {
	pub := s.bob.ethPrivKey.Public().(*ecdsa.PublicKey)
	addr := ethcrypto.PubkeyToAddress(*pub)

	balance, err := s.bob.ethClient.BalanceAt(s.ctx, addr, nil)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	log.Info("Bob's balance before claim: ", balance)

	// call swap.Swap.Claim() w/ b.privkeys.sk, revealing Bob's secret spend key
	secret := s.privkeys.SpendKeyBytes()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret))

	tx, err := s.contract.Claim(s.txOpts, sc)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	log.Infof("sent Claim tx, tx hash=%s", tx.Hash())

	if _, ok := common.WaitForReceipt(s.ctx, s.bob.ethClient, tx.Hash()); !ok {
		return ethcommon.Hash{}, errors.New("failed to check Claim transaction receipt")
	}

	balance, err = s.bob.ethClient.BalanceAt(s.ctx, addr, nil)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	log.Info("Bob's balance after claim: ", balance)
	return tx.Hash(), nil
}
