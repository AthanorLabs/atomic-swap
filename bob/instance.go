package bob

import (
	"context"
	"crypto/ecdsa"
	"errors"
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
	log = logging.Logger("bob")
)

// Instance implements the functionality that will be needed by a user who owns XMR
// and wishes to swap for ETH.
type Instance struct {
	ctx      context.Context
	env      common.Environment
	basepath string

	client                     monero.Client
	daemonClient               monero.DaemonClient
	walletFile, walletPassword string

	ethClient  *ethclient.Client
	ethPrivKey *ecdsa.PrivateKey
	callOpts   *bind.CallOpts
	ethAddress ethcommon.Address
	chainID    *big.Int
	gasPrice   *big.Int
	gasLimit   uint64

	net net.MessageSender

	offerManager *offerManager

	swapMu    sync.Mutex
	swapState *swapState
}

// Config contains the configuration values for a new Bob instance.
type Config struct {
	Ctx                        context.Context
	Basepath                   string
	MoneroWalletEndpoint       string
	MoneroDaemonEndpoint       string // only needed for development
	WalletFile, WalletPassword string
	EthereumEndpoint           string
	EthereumPrivateKey         string
	Environment                common.Environment
	ChainID                    int64
	GasPrice                   *big.Int
	GasLimit                   uint64
}

// NewInstance returns a new *bob.Instance.
// It accepts an endpoint to a monero-wallet-rpc instance where account 0 contains Bob's XMR.
func NewInstance(cfg *Config) (*Instance, error) {
	if cfg.Environment == common.Development && cfg.MoneroDaemonEndpoint == "" {
		return nil, errors.New("environment is development, must provide monero daemon endpoint")
	}

	pk, err := crypto.HexToECDSA(cfg.EthereumPrivateKey)
	if err != nil {
		return nil, err
	}

	ec, err := ethclient.Dial(cfg.EthereumEndpoint)
	if err != nil {
		return nil, err
	}

	// auth.GasLimit = 3027733
	// auth.GasPrice = big.NewInt(2000000000)

	pub := pk.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)

	// monero-wallet-rpc client
	walletClient := monero.NewClient(cfg.MoneroWalletEndpoint)

	// open Bob's XMR wallet
	if cfg.WalletFile != "" {
		if err = walletClient.OpenWallet(cfg.WalletFile, cfg.WalletPassword); err != nil {
			return nil, err
		}
	} else {
		log.Warn("monero wallet-file not set; must be set via RPC call personal_setMoneroWalletFile before making an offer")
	}

	// this is only used in the monero development environment to generate new blocks
	var daemonClient monero.DaemonClient
	if cfg.Environment == common.Development {
		daemonClient = monero.NewClient(cfg.MoneroDaemonEndpoint)
	}

	return &Instance{
		ctx:            cfg.Ctx,
		basepath:       cfg.Basepath,
		env:            cfg.Environment,
		client:         walletClient,
		daemonClient:   daemonClient,
		walletFile:     cfg.WalletFile,
		walletPassword: cfg.WalletPassword,
		ethClient:      ec,
		ethPrivKey:     pk,
		callOpts: &bind.CallOpts{
			From:    addr,
			Context: cfg.Ctx,
		},
		ethAddress:   addr,
		chainID:      big.NewInt(cfg.ChainID),
		offerManager: newOfferManager(),
	}, nil
}

// SetMessageSender sets the Instance's net.MessageSender interface.
func (b *Instance) SetMessageSender(n net.MessageSender) {
	b.net = n
}

// SetMoneroWalletFile sets the Instance's current monero wallet file.
func (b *Instance) SetMoneroWalletFile(file, password string) error {
	_ = b.client.CloseWallet()
	return b.client.OpenWallet(file, password)
}

// SetGasPrice sets the ethereum gas price for the instance to use (in wei).
func (b *Instance) SetGasPrice(gasPrice uint64) {
	b.gasPrice = big.NewInt(0).SetUint64(gasPrice)
}

func (b *Instance) openWallet() error { //nolint
	return b.client.OpenWallet(b.walletFile, b.walletPassword)
}

// watchForReady watches for Alice to call Ready() on the swap contract, allowing
// Bob to call Claim().
func (s *swapState) watchForReady() (<-chan struct{}, error) { //nolint:unused
	watchOpts := &bind.WatchOpts{
		Context: s.ctx,
	}

	done := make(chan struct{})
	ch := make(chan *swap.SwapIsReady)
	defer close(done)

	// watch for Refund() event on chain, calculate unlock key as result
	sub, err := s.contract.WatchIsReady(watchOpts, ch)
	if err != nil {
		return nil, err
	}

	defer sub.Unsubscribe()

	go func() {
		for {
			select {
			case event := <-ch:
				if !event.B {
					continue
				}

				// contract is ready!!
				close(done)
			case <-s.ctx.Done():
				return
			}
		}
	}()

	return done, nil
}

// watchForRefund watches for the Refund event in the contract.
// This should be called before LockFunds.
// If a keypair is sent over this channel, the rest of the protocol should be aborted.
//
// If Alice chooses to refund and thus reveals s_a,
// the private spend and view keys that contain the previously locked monero
// ((s_a + s_b), (v_a + v_b)) are sent over the channel.
// Bob can then use these keys to move his funds if he wishes.
func (s *swapState) watchForRefund() (<-chan *mcrypto.PrivateKeyPair, error) { //nolint:unused
	watchOpts := &bind.WatchOpts{
		Context: s.ctx,
	}

	out := make(chan *mcrypto.PrivateKeyPair)
	ch := make(chan *swap.SwapRefunded)
	defer close(out)

	// watch for Refund() event on chain, calculate unlock key as result
	sub, err := s.contract.WatchRefunded(watchOpts, ch)
	if err != nil {
		return nil, err
	}

	defer sub.Unsubscribe()

	go func() {
		for {
			select {
			case refund := <-ch:
				if refund == nil {
					continue
				}

				// got Alice's secret
				sa := refund.S
				skA, err := mcrypto.NewPrivateSpendKey(sa[:])
				if err != nil {
					log.Info("failed to convert Alice's secret into a key: %w", err)
					return
				}

				vkA, err := skA.View()
				if err != nil {
					log.Info("failed to get view key from Alice's secret spend key: %w", err)
					return
				}

				skAB := mcrypto.SumPrivateSpendKeys(skA, s.privkeys.SpendKey())
				vkAB := mcrypto.SumPrivateViewKeys(vkA, s.privkeys.ViewKey())
				kpAB := mcrypto.NewPrivateKeyPair(skAB, vkAB)
				out <- kpAB
			case <-s.ctx.Done():
				return
			}
		}
	}()

	return out, nil
}
