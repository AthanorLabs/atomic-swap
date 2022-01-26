package alice

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/protocol/swap"
	"github.com/noot/atomic-swap/swapfactory"

	logging "github.com/ipfs/go-log"
)

var (
	log                    = logging.Logger("alice")
	defaultTimeoutDuration = big.NewInt(60 * 60 * 24) // 1 day = 60s * 60min * 24hr
)

// Instance implements the functionality that will be used by a user who owns ETH
// and wishes to swap for XMR.
type Instance struct {
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

	swapManager *swap.Manager
	swapFactory *swapfactory.SwapFactory
}

// Config contains the configuration values for a new Alice instance.
type Config struct {
	Ctx                  context.Context
	Basepath             string
	MoneroWalletEndpoint string
	EthereumClient       *ethclient.Client
	EthereumPrivateKey   *ecdsa.PrivateKey
	SwapContract         *swapfactory.SwapFactory
	Environment          common.Environment
	ChainID              *big.Int
	GasPrice             *big.Int
	GasLimit             uint64
	SwapManager          *swap.Manager
}

// NewInstance returns a new instance of Alice.
// It accepts an endpoint to a monero-wallet-rpc instance where Alice will generate
// the account in which the XMR will be deposited.
func NewInstance(cfg *Config) (*Instance, error) {
	pub := cfg.EthereumPrivateKey.Public().(*ecdsa.PublicKey)

	// TODO: check that Alice's monero-wallet-cli endpoint has wallet-dir configured
	return &Instance{
		ctx:        cfg.Ctx,
		basepath:   cfg.Basepath,
		env:        cfg.Environment,
		ethPrivKey: cfg.EthereumPrivateKey,
		ethClient:  cfg.EthereumClient,
		client:     monero.NewClient(cfg.MoneroWalletEndpoint),
		callOpts: &bind.CallOpts{
			From:    crypto.PubkeyToAddress(*pub),
			Context: cfg.Ctx,
		},
		chainID:     cfg.ChainID,
		swapManager: cfg.SwapManager,
		swapFactory: cfg.SwapContract,
	}, nil
}

// SetMessageSender sets the Instance's net.MessageSender interface.
func (a *Instance) SetMessageSender(n net.MessageSender) {
	a.net = n
}

// SetGasPrice sets the ethereum gas price for the instance to use (in wei).
func (a *Instance) SetGasPrice(gasPrice uint64) {
	a.gasPrice = big.NewInt(0).SetUint64(gasPrice)
}
