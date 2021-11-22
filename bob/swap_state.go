package bob

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/swap-contract"
)

var nextID uint64 = 0

var (
	errMissingKeys    = errors.New("did not receive Alice's public spend or view key")
	errMissingAddress = errors.New("got empty contract address")
)

type swapState struct {
	*bob
	ctx    context.Context
	cancel context.CancelFunc

	id                            uint64
	providesAmount, desiredAmount uint64

	// our keys for this session
	privkeys *monero.PrivateKeyPair
	pubkeys  *monero.PublicKeyPair

	// swap contract and timeouts in it; set once contract is deployed
	contract     *swap.Swap
	contractAddr ethcommon.Address
	t0, t1       time.Time

	// Alice's keys for this session
	alicePublicKeys     *monero.PublicKeyPair
	alicePrivateViewKey *monero.PrivateViewKey

	// next expected network message
	nextExpectedMessage net.Message

	// channels
	readyCh chan struct{}

	// set to true on claiming the ETH
	success bool
}

func newSwapState(b *bob, providesAmount, desiredAmount uint64) *swapState {
	ctx, cancel := context.WithCancel(b.ctx)

	s := &swapState{
		ctx:                 ctx,
		cancel:              cancel,
		bob:                 b,
		id:                  nextID,
		providesAmount:      providesAmount,
		desiredAmount:       desiredAmount,
		nextExpectedMessage: &net.SendKeysMessage{},
		readyCh:             make(chan struct{}),
	}

	nextID++
	return s
}

func (s *swapState) SendKeysMessage() (*net.SendKeysMessage, error) {
	sk, vk, err := s.generateKeys()
	if err != nil {
		return nil, err
	}

	sh := s.privkeys.SpendKey().Hash()

	return &net.SendKeysMessage{
		PublicSpendKey: sk.Hex(),
		PrivateViewKey: vk.Hex(),
		SpendKeyHash:   hex.EncodeToString(sh[:]),
		EthAddress:     s.bob.ethAddress.String(),
	}, nil
}

// ProtocolComplete is called by the network when the protocol stream closes.
// If it closes prematurely, we need to perform recovery.
func (s *swapState) ProtocolComplete() {
	// stop all running goroutines
	s.cancel()

	defer func() {
		s.bob.swapState = nil
	}()

	if s.success {
		return
	}

	switch s.nextExpectedMessage.(type) {
	case *net.SendKeysMessage:
		// we are fine, as we only just initiated the protocol.
	case *net.NotifyContractDeployed:
		// we were waiting for the contract to be deployed, but haven't
		// locked out funds yet, so we're fine.
	case *net.NotifyReady:
		// we already locked our funds - need to wait until we can claim
		// the funds (ie. wait until after t0)
		if err := s.tryClaim(); err != nil {
			log.Errorf("failed to claim funds: err=%s", err)
		}
	default:
		log.Errorf("unexpected nextExpectedMessage in ProtocolComplete: type=%T", s.nextExpectedMessage)
	}
}

func (s *swapState) tryClaim() error {
	untilT0 := time.Until(s.t0)
	untilT1 := time.Until(s.t1)

	if untilT0 < 0 {
		// we need to wait until t0 to claim
		log.Infof("waiting until time %s to refund", s.t0)
		<-time.After(untilT0)
	}

	if untilT1 > 0 { //nolint
		// we've passed t1, our only option now is for Alice to refund
		// and we can regain control of the locked XMR.
		// TODO: watch contract for Refund() to be called.
	}

	_, err := s.claimFunds()
	return err
}

// HandleProtocolMessage is called by the network to handle an incoming message.
// If the message received is not the expected type for the point in the protocol we're at,
// this function will return an error.
func (s *swapState) HandleProtocolMessage(msg net.Message) (net.Message, bool, error) {
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

		s.nextExpectedMessage = &net.NotifyReady{}
		log.Infof("got Swap contract address! address=%s", msg.Address)

		if err := s.setContract(ethcommon.HexToAddress(msg.Address)); err != nil {
			return nil, true, fmt.Errorf("failed to instantiate contract instance: %w", err)
		}

		addrAB, err := s.lockFunds(s.providesAmount)
		if err != nil {
			return nil, true, fmt.Errorf("failed to lock funds: %w", err)
		}

		out := &net.NotifyXMRLock{
			Address: string(addrAB),
		}

		// set t0 and t1
		st0, err := s.contract.Timeout0(s.bob.callOpts)
		if err != nil {
			return nil, true, fmt.Errorf("failed to get timeout0 from contract: err=%w", err)
		}

		s.t0 = time.Unix(st0.Int64(), 0)

		st1, err := s.contract.Timeout1(s.bob.callOpts)
		if err != nil {
			return nil, true, fmt.Errorf("failed to get timeout1 from contract: err=%w", err)
		}

		s.t1 = time.Unix(st1.Int64(), 0)

		go func() {
			until := time.Until(s.t0)

			select {
			case <-s.ctx.Done():
				return
			case <-time.After(until):
				// we can now call Claim()
				txHash, err := s.claimFunds()
				if err != nil {
					log.Errorf("failed to claim: err=%s", err)
					return
				}

				log.Debug("funds claimed!")

				// send *net.NotifyClaimed
				if err := s.net.SendSwapMessage(&net.NotifyClaimed{
					TxHash: txHash,
				}); err != nil {
					log.Errorf("failed to send NotifyClaimed message: err=%s", err)
				}
			case <-s.readyCh:
				return
			}
		}()

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
			TxHash: txHash,
		}

		return out, true, nil
	case *net.NotifyRefund:
		// generate monero wallet, regaining control over locked funds
		addr, err := s.handleRefund(msg.TxHash)
		if err != nil {
			return nil, true, err
		}

		log.Infof("regained control over monero account %s", addr)
		return nil, true, nil
	default:
		return nil, true, errors.New("unexpected message type")
	}
}

func (s *swapState) handleSendKeysMessage(msg *net.SendKeysMessage) error {
	if msg.PublicSpendKey == "" || msg.PrivateViewKey == "" {
		return errMissingKeys
	}

	if msg.SpendKeyHash == "" {
		return errors.New("did not receive SpendKeyHash")
	}

	// TODO: verify hash derives view key, and that view only wallet can be generated

	log.Debug("got Alice's public keys")
	s.nextExpectedMessage = &net.NotifyContractDeployed{}

	vk, err := monero.NewPrivateViewKeyFromHex(msg.PrivateViewKey)
	if err != nil {
		return fmt.Errorf("failed to generate Alice's private view key: %w", err)
	}

	s.alicePrivateViewKey = vk

	kp, err := monero.NewPublicKeyPairFromHex(msg.PublicSpendKey, vk.Public().Hex())
	if err != nil {
		return fmt.Errorf("failed to generate Alice's public keys: %w", err)
	}

	s.setAlicePublicKeys(kp)
	return nil
}

func (s *swapState) handleRefund(txHash string) (monero.Address, error) {
	receipt, err := s.bob.ethClient.TransactionReceipt(s.ctx, ethcommon.HexToHash(txHash))
	if err != nil {
		return "", err
	}

	if len(receipt.Logs) == 0 {
		return "", errors.New("claim transaction has no logs")
	}

	abi, err := abi.JSON(strings.NewReader(swap.SwapABI))
	if err != nil {
		return "", err
	}

	data := receipt.Logs[0].Data
	res, err := abi.Unpack("Refunded", data)
	if err != nil {
		return "", err
	}

	log.Debug("got Alice's secret: ", hex.EncodeToString(res[0].(*big.Int).Bytes()))

	// got Alice's secret
	sa := res[0].([32]byte)
	skA, err := monero.NewPrivateSpendKey(sa[:])
	if err != nil {
		log.Errorf("failed to convert Alice's secret into a key: %s", err)
		return "", err
	}

	vkA, err := skA.View()
	if err != nil {
		log.Errorf("failed to convert Alice's spend key into a view key: %s", err)
		return "", err
	}

	skAB := monero.SumPrivateSpendKeys(skA, s.privkeys.SpendKey())
	vkAB := monero.SumPrivateViewKeys(vkA, s.privkeys.ViewKey())
	kpAB := monero.NewPrivateKeyPair(skAB, vkAB)

	// write keys to file in case something goes wrong
	// TODO: configure basepath
	if err = common.WriteKeysToFile("/tmp/swap-xmr", kpAB); err != nil {
		return "", err
	}

	pkAB := kpAB.PublicKeyPair()
	log.Info("public spend keys: ", pkAB.SpendKey().Hex())
	log.Info("public view keys: ", pkAB.ViewKey().Hex())

	return s.createMoneroWallet(kpAB)
}

// createMoneroWallet creates Alice's monero wallet after Bob calls Claim().
func (s *swapState) createMoneroWallet(kpAB *monero.PrivateKeyPair) (monero.Address, error) {
	t := time.Now().Format("2006-Jan-2-15:04:05")
	walletName := fmt.Sprintf("bob-swap-wallet-%s", t)
	if err := s.bob.client.GenerateFromKeys(kpAB, walletName, ""); err != nil {
		// TODO: save the keypair on disk!!! otherwise we lose the keys
		return "", err
	}

	log.Info("created wallet: ", walletName)

	if err := s.bob.client.Refresh(); err != nil {
		return "", err
	}

	balance, err := s.bob.client.GetBalance(0)
	if err != nil {
		return "", err
	}

	log.Info("wallet balance: ", balance.Balance)
	s.success = true
	return kpAB.Address(), nil
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
