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
	SwapManager          *swap.Manager
}

// NewInstance returns a new instance of Alice.
// It accepts an endpoint to a monero-wallet-rpc instance where Alice will generate
// the account in which the XMR will be deposited.
func NewInstance(cfg *Config) (*Instance, error) {
	pk, err := crypto.HexToECDSA(cfg.EthereumPrivateKey)
	if err != nil {
		return nil, err
	}

	ec, err := ethclient.Dial(cfg.EthereumEndpoint)
	if err != nil {
		return nil, err
	}

	pub := pk.Public().(*ecdsa.PublicKey)

	// TODO: check that Alice's monero-wallet-cli endpoint has wallet-dir configured
	return &Instance{
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
		chainID:     big.NewInt(cfg.ChainID),
		swapManager: cfg.SwapManager,
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
