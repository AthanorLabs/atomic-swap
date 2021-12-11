package alice

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero"
	mcrypto "github.com/noot/atomic-swap/monero/crypto"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/swap-contract"

	logging "github.com/ipfs/go-log"
)

var (
	log                    = logging.Logger("alice")
	defaultTimeoutDuration = big.NewInt(60 * 60 * 24) // 1 day = 60s * 60min * 24hr
)

// alice implements the functions that will be called by a user who owns ETH
// and wishes to swap for XMR.
type alice struct {
	ctx      context.Context
	env      common.Environment
	basepath string

	client monero.Client

	ethPrivKey *ecdsa.PrivateKey
	ethClient  *ethclient.Client
	callOpts   *bind.CallOpts
	chainID    *big.Int
	gasPrice   *big.Int
	gasLimit   uint64

	net net.MessageSender

	// non-nil if a swap is currently happening, nil otherwise
	swapMu    sync.Mutex
	swapState *swapState
}

// Config contains the configuration values for a new Alice instance.
type Config struct {
	Ctx                  context.Context
	Basepath             string
	MoneroWalletEndpoint string
	EthereumEndpoint     string
	EthereumPrivateKey   string
	Environment          common.Environment
	ChainID              int64
	GasPrice             *big.Int
	GasLimit             uint64
}

// NewAlice returns a new instance of Alice.
// It accepts an endpoint to a monero-wallet-rpc instance where Alice will generate
// the account in which the XMR will be deposited.
func NewAlice(cfg *Config) (*alice, error) { //nolint
	pk, err := crypto.HexToECDSA(cfg.EthereumPrivateKey)
	if err != nil {
		return nil, err
	}

	ec, err := ethclient.Dial(cfg.EthereumEndpoint)
	if err != nil {
		return nil, err
	}

	// TODO: add --gas-limit flag and default params for L2
	// auth.GasLimit = 35323600
	// auth.GasPrice = big.NewInt(2000000000)

	pub := pk.Public().(*ecdsa.PublicKey)

	// TODO: check that Alice's monero-wallet-cli endpoint has wallet-dir configured

	return &alice{
		ctx:        cfg.Ctx,
		basepath:   cfg.Basepath,
		env:        cfg.Environment,
		ethPrivKey: pk,
		ethClient:  ec,
		client:     monero.NewClient(cfg.MoneroWalletEndpoint),
		callOpts: &bind.CallOpts{
			From:    crypto.PubkeyToAddress(*pub),
			Context: cfg.Ctx,
		},
		chainID: big.NewInt(cfg.ChainID),
	}, nil
}

func (a *alice) SetMessageSender(n net.MessageSender) {
	a.net = n
}

func (a *alice) SetGasPrice(gasPrice uint64) {
	a.gasPrice = big.NewInt(0).SetUint64(gasPrice)
}

// generateKeys generates Alice's monero spend and view keys (S_b, V_b)
// It returns Alice's public spend key
func (s *swapState) generateKeys() (*mcrypto.PublicKeyPair, error) {
	if s.privkeys != nil {
		return s.pubkeys, nil
	}

	var err error
	s.privkeys, err = mcrypto.GenerateKeys()
	if err != nil {
		return nil, err
	}

	fp := fmt.Sprintf("%s/%d/alice-secret", s.alice.basepath, s.id)
	if err := mcrypto.WriteKeysToFile(fp, s.privkeys, s.alice.env); err != nil {
		return nil, err
	}

	s.pubkeys = s.privkeys.PublicKeyPair()
	return s.pubkeys, nil
}

// setBobKeys sets Bob's public spend key (to be stored in the contract) and Bob's
// private view key (used to check XMR balance before calling Ready())
func (s *swapState) setBobKeys(sk *mcrypto.PublicKey, vk *mcrypto.PrivateViewKey) {
	s.bobPublicSpendKey = sk
	s.bobPrivateViewKey = vk
}

// deployAndLockETH deploys an instance of the Swap contract and locks `amount` ether in it.
func (s *swapState) deployAndLockETH(amount common.EtherAmount) (ethcommon.Address, error) {
	if s.pubkeys == nil {
		return ethcommon.Address{}, errors.New("public keys aren't set")
	}

	if s.bobPublicSpendKey == nil || s.bobPrivateViewKey == nil {
		return ethcommon.Address{}, errors.New("bob's keys aren't set")
	}

	pkAlice := s.pubkeys.SpendKey().Bytes()
	pkBob := s.bobPublicSpendKey.Bytes()

	var pka, pkb [32]byte
	copy(pka[:], common.Reverse(pkAlice))
	copy(pkb[:], common.Reverse(pkBob))

	// TODO: put auth in swapState
	s.txOpts.Value = amount.BigInt()
	defer func() {
		s.txOpts.Value = nil
	}()

	address, tx, swap, err := swap.DeploySwap(s.txOpts, s.alice.ethClient, pkb, pka, s.bobAddress, defaultTimeoutDuration)
	if err != nil {
		return ethcommon.Address{}, fmt.Errorf("failed to deploy Swap.sol: %w", err)
	}

	log.Debugf("deploying Swap.sol, amount=%s txHash=%s", amount, tx.Hash())
	if _, ok := common.WaitForReceipt(s.ctx, s.alice.ethClient, tx.Hash()); !ok {
		return ethcommon.Address{}, errors.New("failed to deploy Swap.sol")
	}

	balance, err := s.alice.ethClient.BalanceAt(s.ctx, address, nil)
	if err != nil {
		return ethcommon.Address{}, err
	}

	log.Debug("contract balance: ", balance)

	s.contract = swap
	return address, nil
}

// ready calls the Ready() method on the Swap contract, indicating to Bob he has until time t_1 to
// call Claim(). Ready() should only be called once Alice sees Bob lock his XMR.
// If time t_0 has passed, there is no point of calling Ready().
func (s *swapState) ready() error {
	tx, err := s.contract.SetReady(s.txOpts)
	if err != nil {
		return err
	}

	if _, ok := common.WaitForReceipt(s.ctx, s.alice.ethClient, tx.Hash()); !ok {
		return errors.New("failed to set IsReady to true in Swap.sol")
	}

	return nil
}

// watchForClaim watches for Bob to call Claim() on the Swap contract.
// When Claim() is called, revealing Bob's secret s_b, the secret key corresponding
// to (s_a + s_b) will be sent over this channel, allowing Alice to claim the XMR it contains.
func (s *swapState) watchForClaim() (<-chan *mcrypto.PrivateKeyPair, error) { //nolint:unused
	watchOpts := &bind.WatchOpts{
		Context: s.ctx,
	}

	out := make(chan *mcrypto.PrivateKeyPair)
	ch := make(chan *swap.SwapClaimed)
	defer close(out)

	// watch for Refund() event on chain, calculate unlock key as result
	sub, err := s.contract.WatchClaimed(watchOpts, ch)
	if err != nil {
		return nil, err
	}

	defer sub.Unsubscribe()

	go func() {
		log.Debug("watching for claim...")
		for {
			select {
			case claim := <-ch:
				if claim == nil {
					continue
				}

				// got Bob's secret
				sb := claim.S
				skB, err := mcrypto.NewPrivateSpendKey(sb[:])
				if err != nil {
					log.Error("failed to convert Bob's secret into a key: ", err)
					return
				}

				vkA, err := skB.View()
				if err != nil {
					log.Error("failed to get view key from Bob's secret spend key: ", err)
					return
				}

				skAB := mcrypto.SumPrivateSpendKeys(skB, s.privkeys.SpendKey())
				vkAB := mcrypto.SumPrivateViewKeys(vkA, s.privkeys.ViewKey())
				kpAB := mcrypto.NewPrivateKeyPair(skAB, vkAB)

				out <- kpAB
				return
			case <-s.ctx.Done():
				return
			}
		}
	}()

	return out, nil
}

// refund calls the Refund() method in the Swap contract, revealing Alice's secret
// and returns to her the ether in the contract.
// If time t_1 passes and Claim() has not been called, Alice should call Refund().
func (s *swapState) refund() (string, error) {
	secret := s.privkeys.SpendKeyBytes()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret))

	if s.contract == nil {
		return "", errors.New("contract is nil")
	}

	log.Infof("attempting to call Refund()...")
	tx, err := s.contract.Refund(s.txOpts, sc)
	if err != nil {
		return "", err
	}

	if _, ok := common.WaitForReceipt(s.ctx, s.alice.ethClient, tx.Hash()); !ok {
		return "", errors.New("failed to call Refund in Swap.sol")
	}

	s.success = true
	return tx.Hash().String(), nil
}

// handleNotifyClaimed handles Bob's reveal after he calls Claim().
// it calls `createMoneroWallet` to create Alice's wallet, allowing her to own the XMR.
func (s *swapState) handleNotifyClaimed(txHash string) (mcrypto.Address, error) {
	receipt, ok := common.WaitForReceipt(s.ctx, s.alice.ethClient, ethcommon.HexToHash(txHash))
	if !ok {
		return "", errors.New("failed check Claim transaction receipt")
	}

	if len(receipt.Logs) == 0 {
		return "", errors.New("claim transaction has no logs")
	}

	skB, err := swap.GetSecretFromLog(receipt.Logs[0], "Claimed")
	if err != nil {
		return "", fmt.Errorf("failed to get secret from log: %w", err)
	}

	skAB := mcrypto.SumPrivateSpendKeys(skB, s.privkeys.SpendKey())
	vkAB := mcrypto.SumPrivateViewKeys(s.bobPrivateViewKey, s.privkeys.ViewKey())
	kpAB := mcrypto.NewPrivateKeyPair(skAB, vkAB)

	// write keys to file in case something goes wrong
	fp := fmt.Sprintf("%s/%d/swap-secret", s.alice.basepath, s.id)
	if err = mcrypto.WriteKeysToFile(fp, kpAB, s.alice.env); err != nil {
		return "", err
	}

	s.success = true
	return monero.CreateMoneroWallet("alice-swap-wallet", s.alice.env, s.alice.client, kpAB)
}
