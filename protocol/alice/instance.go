package alice

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/protocol/swap"
	"github.com/noot/atomic-swap/swapfactory"

	logging "github.com/ipfs/go-log"
)

var (
	log                    = logging.Logger("alice")
	defaultTimeoutDuration = time.Hour * 24
)

// Instance implements the functionality that will be used by a user who owns ETH
// and wishes to swap for XMR.
type Instance struct {
	ctx      context.Context
	env      common.Environment
	basepath string

	client                     monero.Client
	walletFile, walletPassword string
	walletAddress              mcrypto.Address

	ethPrivKey  *ecdsa.PrivateKey
	ethClient   *ethclient.Client
	callOpts    *bind.CallOpts
	chainID     *big.Int
	gasPrice    *big.Int
	gasLimit    uint64
	swapTimeout time.Duration

	net net.MessageSender

	// non-nil if a swap is currently happening, nil otherwise
	swapMu    sync.Mutex
	swapState *swapState

	swapManager  *swap.Manager
	contract     *swapfactory.SwapFactory
	contractAddr ethcommon.Address
}

// Config contains the configuration values for a new Alice instance.
type Config struct {
	Ctx                                    context.Context
	Basepath                               string
	MoneroWalletEndpoint                   string
	MoneroWalletFile, MoneroWalletPassword string
	EthereumClient                         *ethclient.Client
	EthereumPrivateKey                     *ecdsa.PrivateKey
	SwapContract                           *swapfactory.SwapFactory
	SwapContractAddress                    ethcommon.Address
	Environment                            common.Environment
	ChainID                                *big.Int
	GasPrice                               *big.Int
	GasLimit                               uint64
	SwapManager                            *swap.Manager
}

// NewInstance returns a new instance of Alice.
// It accepts an endpoint to a monero-wallet-rpc instance where Alice will generate
// the account in which the XMR will be deposited.
func NewInstance(cfg *Config) (*Instance, error) {
	if cfg.Environment == common.Development {
		defaultTimeoutDuration = time.Minute
	}

	pub := cfg.EthereumPrivateKey.Public().(*ecdsa.PublicKey)

	walletClient := monero.NewClient(cfg.MoneroWalletEndpoint)

	// open XMR wallet, if it exists
	if cfg.MoneroWalletFile != "" {
		if err := walletClient.OpenWallet(cfg.MoneroWalletFile, cfg.MoneroWalletPassword); err != nil {
			return nil, err
		}
	} else {
		log.Info("monero wallet file not set; creating wallet swap-deposit-wallet")
		err := walletClient.CreateWallet("swap-deposit-wallet", "")
		if err != nil {
			return nil, fmt.Errorf("failed to create swap deposit wallet: %w", err)
		}
	}

	// get wallet address to deposit funds into at end of swap
	address, err := walletClient.GetAddress(0)
	if err != nil {
		return nil, fmt.Errorf("failed to get monero wallet address: %w", err)
	}

	err = walletClient.CloseWallet()
	if err != nil {
		return nil, fmt.Errorf("failed to close wallet: %w", err)
	}

	// TODO: check that Alice's monero-wallet-cli endpoint has wallet-dir configured
	return &Instance{
		ctx:            cfg.Ctx,
		basepath:       cfg.Basepath,
		env:            cfg.Environment,
		ethPrivKey:     cfg.EthereumPrivateKey,
		ethClient:      cfg.EthereumClient,
		client:         walletClient,
		walletFile:     cfg.MoneroWalletFile,
		walletPassword: cfg.MoneroWalletPassword,
		walletAddress:  mcrypto.Address(address.Address),
		callOpts: &bind.CallOpts{
			From:    crypto.PubkeyToAddress(*pub),
			Context: cfg.Ctx,
		},
		chainID:      cfg.ChainID,
		swapManager:  cfg.SwapManager,
		contract:     cfg.SwapContract,
		contractAddr: cfg.SwapContractAddress,
		swapTimeout:  defaultTimeoutDuration,
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

// Refund is called by the RPC function swap_refund.
// If it's possible to refund the ongoing swap, it does that, then notifies the counterparty.
func (a *Instance) Refund() (ethcommon.Hash, error) {
	a.swapMu.Lock()
	defer a.swapMu.Unlock()

	if a.swapState == nil {
		return ethcommon.Hash{}, errNoOngoingSwap
	}

	return a.swapState.doRefund()
}

// GetOngoingSwapState ...
func (a *Instance) GetOngoingSwapState() common.SwapState {
	return a.swapState
}

// SetSwapTimeout sets the duration between the swap being initiated on-chain and the timeout t0,
// and the duration between t0 and t1.
func (a *Instance) SetSwapTimeout(timeout time.Duration) {
	a.swapTimeout = timeout
}
