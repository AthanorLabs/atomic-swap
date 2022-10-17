package xmrmaker

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strings"
	"sync"
	"time"

	eth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/fatih/color" //nolint:misspell

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/crypto/secp256k1"
	"github.com/athanorlabs/atomic-swap/dleq"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/net/message"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	pswap "github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/txsender"
	"github.com/athanorlabs/atomic-swap/protocol/xmrmaker/offers"
	"github.com/athanorlabs/atomic-swap/relayer"
)

const revertSwapCompleted = "swap is already completed"

var refundedTopic = common.GetTopic(common.RefundedEventSignature)

type swapState struct {
	backend.Backend
	sender txsender.Sender

	ctx     context.Context
	cancel  context.CancelFunc
	stateMu sync.Mutex

	info         *pswap.Info
	offer        *types.Offer
	offerExtra   *types.OfferExtra
	offerManager *offers.Manager

	// our keys for this session
	dleqProof    *dleq.Proof
	secp256k1Pub *secp256k1.PublicKey
	privkeys     *mcrypto.PrivateKeyPair
	pubkeys      *mcrypto.PublicKeyPair

	// swap contract and timeouts in it
	contract       *contracts.SwapFactory
	contractAddr   ethcommon.Address
	contractSwapID [32]byte
	contractSwap   contracts.SwapFactorySwap
	t0, t1         time.Time

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

func newSwapState(
	b backend.Backend,
	offer *types.Offer,
	offerExtra *types.OfferExtra,
	om *offers.Manager,
	providesAmount common.MoneroAmount,
	desiredAmount common.EtherAmount,
) (*swapState, error) {
	exchangeRate := types.ExchangeRate(providesAmount.AsMonero() / desiredAmount.AsEther())
	stage := types.ExpectingKeys
	if offerExtra.StatusCh == nil {
		offerExtra.StatusCh = make(chan types.Status, 7)
	}
	offerExtra.StatusCh <- stage
	info := pswap.NewInfo(offer.GetID(), types.ProvidesXMR, providesAmount.AsMonero(), desiredAmount.AsEther(),
		exchangeRate, offer.EthAsset, stage, offerExtra.StatusCh)
	if err := b.SwapManager().AddSwap(info); err != nil {
		return nil, err
	}

	var sender txsender.Sender
	if offer.EthAsset != types.EthAssetETH {
		erc20Contract, err := contracts.NewIERC20(offer.EthAsset.Address(), b.EthClient())
		if err != nil {
			return nil, err
		}

		sender, err = b.NewTxSender(offer.EthAsset.Address(), erc20Contract)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		sender, err = b.NewTxSender(offer.EthAsset.Address(), nil)
		if err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithCancel(b.Ctx())
	s := &swapState{
		ctx:                 ctx,
		cancel:              cancel,
		Backend:             b,
		sender:              sender,
		offer:               offer,
		offerExtra:          offerExtra,
		offerManager:        om,
		nextExpectedMessage: &net.SendKeysMessage{},
		readyCh:             make(chan struct{}),
		info:                info,
		done:                make(chan struct{}),
	}

	return s, nil
}

func (s *swapState) lockState() {
	s.stateMu.Lock()
}

func (s *swapState) unlockState() {
	s.stateMu.Unlock()
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

// InfoFile returns the swap's infoFile path
func (s *swapState) InfoFile() string {
	return s.offerExtra.InfoFile
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
// If the swap already completed successfully, this function does not do anything regarding the protocol.
func (s *swapState) Exit() error {
	if s == nil {
		return errNilSwapState
	}

	s.lockState()
	defer s.unlockState()
	return s.exit()
}

// exit is the same as Exit, but assumes the calling code block already holds the swapState lock.
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
			_, err := s.offerManager.AddOffer(s.offer, s.offerExtra.RelayerEndpoint, s.offerExtra.RelayerFee)
			if err != nil {
				log.Warnf("failed to re-add offer %s: %s", s.offer.GetID(), err)
			}
		}

		close(s.done)
	}()

	if s.info.Status() == types.CompletedSuccess {
		str := color.New(color.Bold).Sprintf("**swap completed successfully: id=%s**", s.ID())
		log.Info(str)
		return nil
	}

	if s.info.Status() == types.CompletedRefund {
		str := color.New(color.Bold).Sprintf("**swap refunded successfully: id=%s**", s.ID())
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
			// we should wait for both events in the contract and proceed accordingly. (#162)
			//
			// we already locked our funds - need to wait until we can claim
			// the funds (ie. wait until after t0)
			txHash, err := s.tryClaim()
			if err != nil {
				// note: this shouldn't happen, as it means we had a race condition somewhere
				if strings.Contains(err.Error(), revertSwapCompleted) && !s.info.Status().IsOngoing() {
					return nil
				}

				log.Errorf("failed to claim funds: err=%s", err)
			} else {
				log.Infof("claimed ether! transaction hash=%s", txHash)
				s.clearNextExpectedMessage(types.CompletedSuccess)
				return nil
			}

			// TODO: keep retrying until success (#162)
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
	if err = pcommon.WriteSharedSwapKeyPairToFile(s.offerExtra.InfoFile, kpAB, s.Env()); err != nil {
		return "", err
	}

	s.LockClient()
	defer s.UnlockClient()
	return monero.CreateWallet("xmrmaker-swap-wallet", s.Env(), s, kpAB)
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
		matches, err := contracts.CheckIfLogIDMatches(log, refundedEvent, s.contractSwapID) //nolint:govet
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

	sa, err := contracts.GetSecretFromLog(&foundLog, refundedEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret from log: %w", err)
	}

	return sa, nil
}

func (s *swapState) tryClaim() (ethcommon.Hash, error) {
	stage, err := s.Contract().Swaps(s.CallOpts(), s.contractSwapID)
	if err != nil {
		return ethcommon.Hash{}, err
	}
	switch stage {
	case contracts.StageInvalid:
		return ethcommon.Hash{}, errClaimInvalid
	case contracts.StageCompleted:
		return ethcommon.Hash{}, errClaimSwapComplete
	case contracts.StagePending, contracts.StageReady:
		// do nothing
	default:
		panic("Unhandled stage value")
	}

	ts, err := s.LatestBlockTimestamp(s.ctx)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	// The block that our claim transaction goes into needs a timestamp that is strictly less
	// than T1. Since the minimum interval between blocks is 1 second, the current block must
	// be at least 2 seconds before T1 for a non-zero chance of the next block having a
	// timestamp that is strictly less than T1.
	if ts.After(s.t1.Add(-2 * time.Second)) {
		// We've passed t1, so the only way we can regain control of the locked XMR is for
		// XMRTaker to call refund on the contract.
		return ethcommon.Hash{}, errClaimPastTime
	}

	if ts.Before(s.t0) && stage != contracts.StageReady {
		// TODO: t0 could be 24 hours from now. Don't we want to poll the stage periodically? (#163)
		// we need to wait until t0 to claim
		log.Infof("waiting until time %s to claim, time now=%s", s.t0, time.Now())
		err = s.WaitForTimestamp(s.ctx, s.t0)
		if err != nil {
			return ethcommon.Hash{}, err
		}
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

	return pcommon.WriteKeysToFile(s.offerExtra.InfoFile, s.privkeys, s.Env())
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
	s.contractAddr = address

	var err error
	s.contract, err = s.NewSwapFactory(address)
	if err != nil {
		return err
	}

	s.sender.SetContractAddress(address)
	s.sender.SetContract(s.contract)
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
	tx, _, err := s.TransactionByHash(s.ctx, txHash)
	if err != nil {
		return err
	}

	if tx.To() == nil || *(tx.To()) != s.ContractAddr() {
		return errInvalidETHLockedTransaction
	}

	receipt, err := s.WaitForReceipt(s.ctx, txHash)
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

	// TODO: check timeouts (#161)

	// check asset of created swap
	if types.EthAsset(s.contractSwap.Asset) != types.EthAsset(event.Asset) {
		return fmt.Errorf("swap asset is not expected: got %v, expected %v", event.Asset, s.contractSwap.Asset)
	}

	// check value of created swap
	if s.contractSwap.Value.Cmp(event.Value) != 0 {
		return fmt.Errorf("swap value is not expected: got %v, expected %v", event.Value, s.contractSwap.Value)
	}

	return nil
}

// lockFunds locks XMRMaker's funds in the monero account specified by public key
// (S_a + S_b), viewable with (V_a + V_b)
// It accepts the amount to lock as the input
func (s *swapState) lockFunds(amount common.MoneroAmount) (mcrypto.Address, error) {
	kp := mcrypto.SumSpendAndViewKeys(s.xmrtakerPublicKeys, s.pubkeys)
	log.Infof("going to lock XMR funds, amount(piconero)=%d", amount)

	s.LockClient()
	defer s.UnlockClient()

	balance, err := s.GetBalance(0)
	if err != nil {
		return "", err
	}

	log.Debug("total XMR balance: ", balance.Balance)
	log.Info("unlocked XMR balance: ", balance.UnlockedBalance)

	address := kp.Address(s.Env())
	txResp, err := s.Transfer(address, 0, uint64(amount))
	if err != nil {
		return "", err
	}

	log.Infof("locked XMR, txHash=%s fee=%d", txResp.TxHash, txResp.Fee)

	// wait for a new block
	height, err := monero.WaitForBlocks(s, 1)
	if err != nil {
		return "", err
	}
	log.Infof("monero block height: %d", height)

	if err := s.Refresh(); err != nil {
		return "", err
	}

	log.Infof("successfully locked XMR funds: address=%s", address)
	return address, nil
}

// claimFunds redeems XMRMaker's ETH funds by calling Claim() on the contract
func (s *swapState) claimFunds() (ethcommon.Hash, error) {
	addr := s.EthAddress()

	var (
		symbol   string
		decimals uint8
		err      error
	)
	if types.EthAsset(s.contractSwap.Asset) != types.EthAssetETH {
		_, symbol, decimals, err = s.ERC20Info(s.ctx, s.contractSwap.Asset)
		if err != nil {
			return ethcommon.Hash{}, fmt.Errorf("failed to get ERC20 info: %w", err)
		}
	}

	if types.EthAsset(s.contractSwap.Asset) == types.EthAssetETH {
		balance, err := s.BalanceAt(s.ctx, addr, nil) //nolint:govet
		if err != nil {
			return ethcommon.Hash{}, err
		}
		log.Infof("balance before claim: %v ETH", common.EtherAmount(*balance).AsEther())
	} else {
		balance, err := s.ERC20BalanceAt(s.ctx, s.contractSwap.Asset, addr, nil) //nolint:govet
		if err != nil {
			return ethcommon.Hash{}, err
		}
		log.Infof("balance before claim: %v %s", common.EtherAmount(*balance).ToDecimals(decimals), symbol)
	}

	var (
		txHash ethcommon.Hash
	)

	// call swap.Swap.Claim() w/ b.privkeys.sk, revealing XMRMaker's secret spend key
	if s.offerExtra.RelayerEndpoint != "" {
		// relayer endpoint is set, claim using relayer
		// TODO: eventually update when relayer discovery is implemented
		txHash, err = s.claimRelayer()
		if err != nil {
			return ethcommon.Hash{}, err
		}
	} else {
		// claim and wait for tx to be included
		sc := s.getSecret()
		txHash, _, err = s.sender.Claim(s.contractSwap, sc)
		if err != nil {
			return ethcommon.Hash{}, err
		}

		log.Infof("sent claim tx, tx hash=%s", txHash)
	}

	if types.EthAsset(s.contractSwap.Asset) == types.EthAssetETH {
		balance, err := s.BalanceAt(s.ctx, addr, nil)
		if err != nil {
			return ethcommon.Hash{}, err
		}
		log.Infof("balance after claim: %v ETH", common.EtherAmount(*balance).AsEther())
	} else {
		balance, err := s.ERC20BalanceAt(s.ctx, s.contractSwap.Asset, addr, nil)
		if err != nil {
			return ethcommon.Hash{}, err
		}

		log.Infof("balance after claim: %v %s", common.EtherAmount(*balance).ToDecimals(decimals), symbol)
	}

	return txHash, nil
}

// claimRelayer claims the ETH funds via relayer.
func (s *swapState) claimRelayer() (ethcommon.Hash, error) {
	sk := s.EthPrivateKey()
	forwarderAddress, err := s.Contract().TrustedForwarder(&bind.CallOpts{})
	if err != nil {
		return ethcommon.Hash{}, err
	}

	rc, err := relayer.NewClient(sk, s.EthClient(), s.offerExtra.RelayerEndpoint, forwarderAddress)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	abi, err := abi.JSON(strings.NewReader(contracts.SwapFactoryABI))
	if err != nil {
		return ethcommon.Hash{}, err
	}

	sc := s.getSecret()
	feeValue := calculateRelayerFeeValue(s.contractSwap.Value, s.offerExtra.RelayerFee)
	calldata, err := abi.Pack("claimRelayer", s.contractSwap, sc, feeValue)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	txHash, err := rc.SubmitTransaction(s.ContractAddr(), calldata)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	// wait for inclusion
	_, err = s.WaitForReceipt(s.Ctx(), txHash)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	return txHash, nil
}

// relayerFee is a percentage (not yet divided by 100!!)
func calculateRelayerFeeValue(swapValue *big.Int, relayerFee float64) *big.Int {
	swapValueF := big.NewFloat(0).SetInt(swapValue)
	relayerFeeF := big.NewFloat(relayerFee)
	relayerFeeF = big.NewFloat(0).Quo(relayerFeeF, big.NewFloat(100))
	feeValue := big.NewFloat(0).Mul(swapValueF, relayerFeeF)

	numEtherUnits := big.NewFloat(math.Pow(10, 18))
	feeValueWei := big.NewFloat(0).Mul(feeValue, numEtherUnits)
	wei, _ := feeValueWei.Int(nil)
	return wei
}
