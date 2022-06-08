package xmrmaker

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	eth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/fatih/color" //nolint:misspell

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	"github.com/noot/atomic-swap/crypto/secp256k1"
	"github.com/noot/atomic-swap/dleq"
	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/net/message"
	pcommon "github.com/noot/atomic-swap/protocol"
	"github.com/noot/atomic-swap/protocol/backend"
	pswap "github.com/noot/atomic-swap/protocol/swap"
	"github.com/noot/atomic-swap/swapfactory"
)

const revertSwapCompleted = "swap is already completed"

var (
	// this is from the autogenerated swap.go
	// TODO: generate this ourselves instead of hard-coding
	refundedTopic = ethcommon.HexToHash("0x007c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f")
)

type swapState struct {
	backend.Backend

	ctx    context.Context
	cancel context.CancelFunc
	sync.Mutex
	infofile string

	info         *pswap.Info
	offer        *types.Offer
	offerManager *offerManager
	statusCh     chan types.Status

	// our keys for this session
	dleqProof    *dleq.Proof
	secp256k1Pub *secp256k1.PublicKey
	privkeys     *mcrypto.PrivateKeyPair
	pubkeys      *mcrypto.PublicKeyPair

	// swap contract and timeouts in it; set once contract is deployed
	contractSwapID [32]byte
	contractSwap   swapfactory.SwapFactorySwap
	t0, t1         time.Time
	txOpts         *bind.TransactOpts

	// XMRTaker's keys for this session
	xmrtakerPublicKeys         *mcrypto.PublicKeyPair
	xmrtakerSecp256K1PublicKey *secp256k1.PublicKey

	// next expected network message
	nextExpectedMessage net.Message

	// channels
	readyCh chan struct{}
	done    chan struct{}
	exited  bool

	// address of reclaimed monero wallet, if the swap is refunded77
	moneroReclaimAddress mcrypto.Address
}

func newSwapState(b backend.Backend, offer *types.Offer, om *offerManager, statusCh chan types.Status, infofile string,
	providesAmount common.MoneroAmount, desiredAmount common.EtherAmount) (*swapState, error) {
	txOpts, err := b.TxOpts()
	if err != nil {
		return nil, err
	}

	exchangeRate := types.ExchangeRate(providesAmount.AsMonero() / desiredAmount.AsEther())
	stage := types.ExpectingKeys
	if statusCh == nil {
		statusCh = make(chan types.Status, 7)
	}
	statusCh <- stage
	info := pswap.NewInfo(types.ProvidesXMR, providesAmount.AsMonero(), desiredAmount.AsEther(),
		exchangeRate, stage, statusCh)
	if err := b.SwapManager().AddSwap(info); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(b.Ctx())
	s := &swapState{
		ctx:                 ctx,
		cancel:              cancel,
		Backend:             b,
		offer:               offer,
		offerManager:        om,
		infofile:            infofile,
		nextExpectedMessage: &net.SendKeysMessage{},
		readyCh:             make(chan struct{}),
		txOpts:              txOpts,
		info:                info,
		statusCh:            statusCh,
		done:                make(chan struct{}),
	}

	return s, nil
}

// SendKeysMessage ...
func (s *swapState) SendKeysMessage() (*net.SendKeysMessage, error) {
	if err := s.generateAndSetKeys(); err != nil {
		return nil, err
	}

	return &net.SendKeysMessage{
		ProvidedAmount:     s.info.ProvidedAmount(),
		PublicSpendKey:     s.pubkeys.SpendKey().Hex(),
		PrivateViewKey:     s.privkeys.ViewKey().Hex(),
		DLEqProof:          hex.EncodeToString(s.dleqProof.Proof()),
		Secp256k1PublicKey: s.secp256k1Pub.String(),
		EthAddress:         s.EthAddress().String(),
	}, nil
}

// InfoFile returns the swap's infofile path
func (s *swapState) InfoFile() string {
	return s.infofile
}

// ReceivedAmount returns the amount received, or expected to be received, at the end of the swap
func (s *swapState) ReceivedAmount() float64 {
	return s.info.ReceivedAmount()
}

// ID returns the ID of the swap
func (s *swapState) ID() types.Hash {
	return s.info.ID()
}

// Exit is called by the network when the protocol stream closes, or if the swap_refund RPC endpoint is called.
// It exists the swap by refunding if necessary. If no locking has been done, it simply aborts the swap.
// If the swap already completed successfully, this function does not doing anything in regards to the protoco.
func (s *swapState) Exit() error {
	if s == nil {
		return errNilSwapState
	}

	s.Lock()
	defer s.Unlock()
	return s.exit()
}

func (s *swapState) exit() error {
	if s == nil {
		return errNilSwapState
	}

	if s.exited {
		return nil
	}

	s.exited = true

	log.Debugf("attempting to exit swap: nextExpectedMessage=%v", s.nextExpectedMessage)

	defer func() {
		// stop all running goroutines
		s.cancel()
		s.SwapManager().CompleteOngoingSwap(s.offer.GetID())

		if s.info.Status() != types.CompletedSuccess {
			// re-add offer, as it wasn't taken successfully
			s.offerManager.putOffer(s.offer)
		}

		close(s.done)
	}()

	if s.info.Status() == types.CompletedSuccess {
		str := color.New(color.Bold).Sprintf("**swap completed successfully: id=%d**", s.ID())
		log.Info(str)
		return nil
	}

	if s.info.Status() == types.CompletedRefund {
		str := color.New(color.Bold).Sprintf("**swap refunded successfully: id=%d**", s.ID())
		log.Info(str)
		return nil
	}

	switch s.nextExpectedMessage.(type) {
	case *net.SendKeysMessage:
		// we are fine, as we only just initiated the protocol.
		s.clearNextExpectedMessage(types.CompletedAbort)
		return nil
	case *message.NotifyETHLocked:
		// we were waiting for the contract to be deployed, but haven't
		// locked out funds yet, so we're fine.
		s.clearNextExpectedMessage(types.CompletedAbort)
		return nil
	case *message.NotifyReady:
		// we should check if XMRTaker refunded, if so then check contract for secret
		address, err := s.tryReclaimMonero()
		if err != nil {
			log.Errorf("failed to check for refund: err=%s", err)

			// TODO: depending on the error, we should either retry to refund or try to claim.
			// we should wait for both events in the contract and proceed accordingly.
			//
			// we already locked our funds - need to wait until we can claim
			// the funds (ie. wait until after t0)
			txHash, err := s.tryClaim()
			if err != nil {
				// TODO: this shouldn't happen, as it means we had a race condition somewhere
				if strings.Contains(err.Error(), revertSwapCompleted) && !s.info.Status().IsOngoing() {
					return nil
				}

				log.Errorf("failed to claim funds: err=%s", err)
			} else {
				log.Infof("claimed ether! transaction hash=%s", txHash)
				s.clearNextExpectedMessage(types.CompletedSuccess)
				return nil
			}

			// TODO: keep retrying until success
			return err
		}

		s.clearNextExpectedMessage(types.CompletedRefund)
		s.moneroReclaimAddress = address
		log.Infof("regained private key to monero wallet, address=%s", address)
		return nil
	default:
		s.clearNextExpectedMessage(types.CompletedAbort)
		log.Errorf("unexpected nextExpectedMessage in Exit: type=%T", s.nextExpectedMessage)
		return errUnexpectedMessageType
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
	if err = pcommon.WriteSharedSwapKeyPairToFile(s.infofile, kpAB, s.Env()); err != nil {
		return "", err
	}

	// TODO: check balance
	return monero.CreateMoneroWallet("xmrmaker-swap-wallet", s.Env(), s, kpAB)
}

func (s *swapState) filterForRefund() (*mcrypto.PrivateSpendKey, error) {
	const refundedEvent = "Refunded"

	logs, err := s.FilterLogs(s.ctx, eth.FilterQuery{
		Addresses: []ethcommon.Address{s.ContractAddr()},
		Topics:    [][]ethcommon.Hash{{refundedTopic}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %w", err)
	}

	if len(logs) == 0 {
		return nil, errNoRefundLogsFound
	}

	var (
		foundLog ethtypes.Log
		found    bool
	)

	for _, log := range logs {
		matches, err := swapfactory.CheckIfLogIDMatches(log, refundedEvent, s.contractSwapID) //nolint:govet
		if err != nil {
			continue
		}

		if matches {
			foundLog = log
			found = true
			break
		}
	}

	if !found {
		return nil, errNoRefundLogsFound
	}

	sa, err := swapfactory.GetSecretFromLog(&foundLog, refundedEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret from log: %w", err)
	}

	return sa, nil
}

func (s *swapState) tryClaim() (ethcommon.Hash, error) {
	untilT0 := time.Until(s.t0)
	untilT1 := time.Until(s.t1)
	stage, err := s.Contract().Swaps(s.CallOpts(), s.contractSwapID)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	if untilT0 > 0 && stage != swapfactory.StageReady {
		// we need to wait until t0 to claim
		log.Infof("waiting until time %s to claim, time now=%s", s.t0, time.Now())
		<-time.After(untilT0 + time.Second)
	}

	if untilT1 < 0 {
		// we've passed t1, our only option now is for XMRTaker to refund
		// and we can regain control of the locked XMR.
		return ethcommon.Hash{}, errPastClaimTime
	}

	return s.claimFunds()
}

// generateKeys generates XMRMaker's spend and view keys (s_b, v_b)
// It returns XMRMaker's public spend key and his private view key, so that XMRTaker can see
// if the funds are locked.
func (s *swapState) generateAndSetKeys() error {
	if s == nil {
		return errNilSwapState
	}

	if s.privkeys != nil {
		return nil
	}

	keysAndProof, err := generateKeys()
	if err != nil {
		return err
	}

	s.dleqProof = keysAndProof.DLEqProof
	s.secp256k1Pub = keysAndProof.Secp256k1PublicKey
	s.privkeys = keysAndProof.PrivateKeyPair
	s.pubkeys = keysAndProof.PublicKeyPair

	return pcommon.WriteKeysToFile(s.infofile, s.privkeys, s.Env())
}

func generateKeys() (*pcommon.KeysAndProof, error) {
	return pcommon.GenerateKeysAndProof()
}

// getSecret secrets returns the current secret scalar used to unlock funds from the contract.
func (s *swapState) getSecret() [32]byte {
	secret := s.dleqProof.Secret()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret[:]))
	return sc
}

// setXMRTakerPublicKeys sets XMRTaker's public spend and view keys
func (s *swapState) setXMRTakerPublicKeys(sk *mcrypto.PublicKeyPair, secp256k1Pub *secp256k1.PublicKey) {
	s.xmrtakerPublicKeys = sk
	s.xmrtakerSecp256K1PublicKey = secp256k1Pub
}

// setContract sets the contract in which XMRTaker has locked her ETH.
func (s *swapState) setContract(address ethcommon.Address) error {
	var err error
	// note: this overrides the backend contract
	s.SetContractAddress(address)
	contract, err := s.NewSwapFactory(address)
	if err != nil {
		return err
	}

	s.SetContract(contract)
	return nil
}

func (s *swapState) setTimeouts(t0, t1 *big.Int) {
	s.t0 = time.Unix(t0.Int64(), 0)
	s.t1 = time.Unix(t1.Int64(), 0)
}

// checkContract checks the contract's balance and Claim/Refund keys.
// if the balance doesn't match what we're expecting to receive, or the public keys in the contract
// aren't what we expect, we error and abort the swap.
func (s *swapState) checkContract(txHash ethcommon.Hash) error {
	receipt, err := s.WaitForReceipt(s.ctx, txHash)
	if err != nil {
		return fmt.Errorf("failed to get receipt for New transaction: %w", err)
	}

	// check that New log was emitted
	if len(receipt.Logs) == 0 {
		return errCannotFindNewLog
	}

	event, err := s.Contract().ParseNew(*receipt.Logs[0])
	if err != nil {
		return err
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

	// TODO: check timeouts

	// check value of created swap
	value := s.contractSwap.Value
	expected := common.EtherToWei(s.info.ReceivedAmount()).BigInt()
	if value.Cmp(expected) < 0 {
		return fmt.Errorf("contract does not have expected balance: got %s, expected %s", value, expected)
	}

	return nil
}

// lockFunds locks XMRMaker's funds in the monero account specified by public key
// (S_a + S_b), viewable with (V_a + V_b)
// It accepts the amount to lock as the input
// TODO: units
func (s *swapState) lockFunds(amount common.MoneroAmount) (mcrypto.Address, error) {
	kp := mcrypto.SumSpendAndViewKeys(s.xmrtakerPublicKeys, s.pubkeys)
	log.Infof("going to lock XMR funds, amount(piconero)=%d", amount)

	balance, err := s.GetBalance(0)
	if err != nil {
		return "", err
	}

	log.Debug("total XMR balance: ", balance.Balance)
	log.Info("unlocked XMR balance: ", balance.UnlockedBalance)

	address := kp.Address(s.Env())
	txResp, err := s.Transfer(address, 0, uint(amount))
	if err != nil {
		return "", err
	}

	log.Infof("locked XMR, txHash=%s fee=%d", txResp.TxHash, txResp.Fee)

	xmrmakerAddr, err := s.GetAddress(0)
	if err != nil {
		return "", err
	}

	// if we're on a development --regtest node, generate some blocks
	if s.Env() == common.Development {
		_ = s.GenerateBlocks(xmrmakerAddr.Address, 2)
	} else {
		// otherwise, wait for new blocks
		height, err := monero.WaitForBlocks(s, 1)
		if err != nil {
			return "", err
		}

		log.Infof("monero block height: %d", height)
	}

	if err := s.Refresh(); err != nil {
		return "", err
	}

	log.Infof("successfully locked XMR funds: address=%s", address)
	return address, nil
}

// claimFunds redeems XMRMaker's ETH funds by calling Claim() on the contract
func (s *swapState) claimFunds() (ethcommon.Hash, error) {
	addr := s.EthAddress()

	balance, err := s.BalanceAt(s.ctx, addr, nil)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	log.Infof("balance before claim: %v ETH", common.EtherAmount(*balance).AsEther())

	// call swap.Swap.Claim() w/ b.privkeys.sk, revealing XMRMaker's secret spend key
	sc := s.getSecret()
	txHash, _, err := s.Claim(s.contractSwap, sc)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	log.Infof("sent claim tx, tx hash=%s", txHash)

	balance, err = s.BalanceAt(s.ctx, addr, nil)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	log.Infof("balance after claim: %v ETH", common.EtherAmount(*balance).AsEther())
	return txHash, nil
}
