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
	"github.com/noot/atomic-swap/net"

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
