package alice

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero"
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
	ctx context.Context

	client monero.Client

	ethPrivKey *ecdsa.PrivateKey
	ethClient  *ethclient.Client
	auth       *bind.TransactOpts
	callOpts   *bind.CallOpts

	net net.MessageSender

	// non-nil if a swap is currently happening, nil otherwise
	swapMu    sync.Mutex
	swapState *swapState
}

// NewAlice returns a new instance of Alice.
// It accepts an endpoint to a monero-wallet-rpc instance where Alice will generate
// the account in which the XMR will be deposited.
func NewAlice(ctx context.Context, moneroEndpoint, ethEndpoint, ethPrivKey string) (*alice, error) {
	pk, err := crypto.HexToECDSA(ethPrivKey)
	if err != nil {
		return nil, err
	}

	ec, err := ethclient.Dial(ethEndpoint)
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(1337)) // ganache chainID
	if err != nil {
		return nil, err
	}

	pub := pk.Public().(*ecdsa.PublicKey)

	// TODO: check that Alice's monero-wallet-cli endpoint has wallet-dir configured

	return &alice{
		ctx:        ctx,
		ethPrivKey: pk,
		ethClient:  ec,
		client:     monero.NewClient(moneroEndpoint),
		auth:       auth,
		callOpts: &bind.CallOpts{
			From:    crypto.PubkeyToAddress(*pub),
			Context: ctx,
		},
	}, nil
}

func (a *alice) SetMessageSender(n net.MessageSender) {
	a.net = n
}

// generateKeys generates Alice's monero spend and view keys (S_b, V_b)
// It returns Alice's public spend key
func (s *swapState) generateKeys() (*monero.PublicKeyPair, error) {
	if s.privkeys != nil {
		return s.pubkeys, nil
	}

	var err error
	s.privkeys, err = monero.GenerateKeys()
	if err != nil {
		return nil, err
	}

	// TODO: configure basepath
	// TODO: add swap ID
	if err := common.WriteKeysToFile("/tmp/alice-xmr", s.privkeys); err != nil {
		return nil, err
	}

	s.pubkeys = s.privkeys.PublicKeyPair()
	return s.pubkeys, nil
}

// setBobKeys sets Bob's public spend key (to be stored in the contract) and Bob's
// private view key (used to check XMR balance before calling Ready())
func (s *swapState) setBobKeys(sk *monero.PublicKey, vk *monero.PrivateViewKey) {
	s.bobSpendKey = sk
	s.bobViewKey = vk
}

// deployAndLockETH deploys an instance of the Swap contract and locks `amount` ether in it.
func (s *swapState) deployAndLockETH(amount uint64) (ethcommon.Address, error) {
	pkAlice := s.pubkeys.SpendKey().Bytes()
	pkBob := s.bobSpendKey.Bytes()

	var pka, pkb [32]byte
	copy(pka[:], pkAlice)
	copy(pkb[:], pkBob)

	log.Debug("locking amount: ", amount)
	// TODO: put auth in swapState
	s.alice.auth.Value = big.NewInt(int64(amount))
	defer func() {
		s.alice.auth.Value = nil
	}()

	address, _, swap, err := swap.DeploySwap(s.alice.auth, s.alice.ethClient, pkb, pka, defaultTimeoutDuration)
	if err != nil {
		return ethcommon.Address{}, err
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
	_, err := s.contract.SetReady(s.alice.auth)
	return err
}

// watchForClaim watches for Bob to call Claim() on the Swap contract.
// When Claim() is called, revealing Bob's secret s_b, the secret key corresponding
// to (s_a + s_b) will be sent over this channel, allowing Alice to claim the XMR it contains.
func (s *swapState) watchForClaim() (<-chan *monero.PrivateKeyPair, error) { //nolint:unused
	watchOpts := &bind.WatchOpts{
		Context: s.ctx,
	}

	out := make(chan *monero.PrivateKeyPair)
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
				if claim == nil || claim.S == nil {
					continue
				}

				// got Bob's secret
				sbBytes := claim.S.Bytes()
				var sb [32]byte
				copy(sb[:], sbBytes)

				skB, err := monero.NewPrivateSpendKey(sb[:])
				if err != nil {
					log.Error("failed to convert Bob's secret into a key: ", err)
					return
				}

				vkA, err := skB.View()
				if err != nil {
					log.Error("failed to get view key from Bob's secret spend key: ", err)
					return
				}

				skAB := monero.SumPrivateSpendKeys(skB, s.privkeys.SpendKey())
				vkAB := monero.SumPrivateViewKeys(vkA, s.privkeys.ViewKey())
				kpAB := monero.NewPrivateKeyPair(skAB, vkAB)

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
	sc := big.NewInt(0).SetBytes(secret)

	log.Infof("attempting to call Refund()...")
	tx, err := s.contract.Refund(s.alice.auth, sc)
	if err != nil {
		return "", err
	}

	return tx.Hash().String(), nil
}

// createMoneroWallet creates Alice's monero wallet after Bob calls Claim().
func (s *swapState) createMoneroWallet(kpAB *monero.PrivateKeyPair) (monero.Address, error) {
	t := time.Now().Format("2006-Jan-2-15:04:05")
	walletName := fmt.Sprintf("alice-swap-wallet-%s", t)
	if err := s.alice.client.GenerateFromKeys(kpAB, walletName, ""); err != nil {
		// TODO: save the keypair on disk!!! otherwise we lose the keys
		return "", err
	}

	log.Info("created wallet: ", walletName)

	if err := s.alice.client.Refresh(); err != nil {
		return "", err
	}

	balance, err := s.alice.client.GetBalance(0)
	if err != nil {
		return "", err
	}

	log.Info("wallet balance: ", balance.Balance)
	s.success = true
	return kpAB.Address(), nil
}

// handleNotifyClaimed handles Bob's reveal after he calls Claim().
// it calls `createMoneroWallet` to create Alice's wallet, allowing her to own the XMR.
func (s *swapState) handleNotifyClaimed(txHash string) (monero.Address, error) {
	receipt, err := s.alice.ethClient.TransactionReceipt(s.ctx, ethcommon.HexToHash(txHash))
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
	res, err := abi.Unpack("Claimed", data)
	if err != nil {
		return "", err
	}

	log.Debug("got Bob's secret: ", hex.EncodeToString(res[0].(*big.Int).Bytes()))

	// got Bob's secret
	sbBytes := res[0].(*big.Int).Bytes()
	var sb [32]byte
	copy(sb[:], sbBytes)

	skB, err := monero.NewPrivateSpendKey(sb[:])
	if err != nil {
		log.Errorf("failed to convert Bob's secret into a key: %s", err)
		return "", err
	}

	skAB := monero.SumPrivateSpendKeys(skB, s.privkeys.SpendKey())
	vkAB := monero.SumPrivateViewKeys(s.bobViewKey, s.privkeys.ViewKey())
	kpAB := monero.NewPrivateKeyPair(skAB, vkAB)

	pkAB := kpAB.PublicKeyPair()
	log.Info("public spend keys: ", pkAB.SpendKey().Hex())
	log.Info("public view keys: ", pkAB.ViewKey().Hex())

	return s.createMoneroWallet(kpAB)
}
