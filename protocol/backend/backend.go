package backend

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"time"

	eth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/protocol/swap"
	"github.com/noot/atomic-swap/protocol/txsender"
	"github.com/noot/atomic-swap/swapfactory"

	logging "github.com/ipfs/go-log"
)

const (
	// in total, we will wait up to 1 hour for a transaction to be included
	maxRetries           = 360
	receiptSleepDuration = time.Second * 10
)

var (
	log                    = logging.Logger("protocol/backend")
	defaultTimeoutDuration = time.Hour * 24
)

// Backend provides an interface for both the XMRTaker and XMRMaker into the Monero/Ethereum chains.
// It also interfaces with the network layer.
type Backend interface {
	monero.Client
	monero.DaemonClient
	net.MessageSender
	txsender.Sender

	// ethclient methods
	BalanceAt(ctx context.Context, account ethcommon.Address, blockNumber *big.Int) (*big.Int, error)
	CodeAt(ctx context.Context, account ethcommon.Address, blockNumber *big.Int) ([]byte, error)
	FilterLogs(ctx context.Context, q eth.FilterQuery) ([]types.Log, error)
	TransactionReceipt(ctx context.Context, txHash ethcommon.Hash) (*types.Receipt, error)

	// helpers
	WaitForReceipt(ctx context.Context, txHash ethcommon.Hash) (*types.Receipt, error)
	NewSwapFactory(addr ethcommon.Address) (*swapfactory.SwapFactory, error)

	// getters
	Ctx() context.Context
	Env() common.Environment
	ChainID() *big.Int
	CallOpts() *bind.CallOpts
	TxOpts() (*bind.TransactOpts, error)
	SwapManager() swap.Manager
	EthAddress() ethcommon.Address
	Contract() *swapfactory.SwapFactory
	ContractAddr() ethcommon.Address
	Net() net.MessageSender
	SwapTimeout() time.Duration
	ExternalSender() *txsender.ExternalSender
	XMRDepositAddress() mcrypto.Address

	// setters
	SetSwapTimeout(timeout time.Duration)
	SetGasPrice(uint64)
	SetEthAddress(ethcommon.Address)
	SetXMRDepositAddress(mcrypto.Address)
	SetContract(*swapfactory.SwapFactory)
	SetContractAddress(ethcommon.Address)
}

type backend struct {
	ctx         context.Context
	env         common.Environment
	swapManager swap.Manager

	// monero endpoints
	monero.Client
	monero.DaemonClient

	// monero deposit address (used if xmrtaker has transferBack set to true)
	xmrDepositAddr mcrypto.Address

	// ethereum endpoint and variables
	ethClient  *ethclient.Client
	ethPrivKey *ecdsa.PrivateKey
	callOpts   *bind.CallOpts
	ethAddress ethcommon.Address
	chainID    *big.Int
	gasPrice   *big.Int
	gasLimit   uint64
	txsender.Sender

	// swap contract
	contract     *swapfactory.SwapFactory
	contractAddr ethcommon.Address
	swapTimeout  time.Duration

	// network interface
	net.MessageSender
}

// Config is the config for the Backend
type Config struct {
	Ctx                  context.Context
	MoneroWalletEndpoint string
	MoneroDaemonEndpoint string // only needed for development

	EthereumClient     *ethclient.Client
	EthereumPrivateKey *ecdsa.PrivateKey
	Environment        common.Environment
	ChainID            *big.Int
	GasPrice           *big.Int
	GasLimit           uint64

	SwapContract        *swapfactory.SwapFactory
	SwapContractAddress ethcommon.Address

	SwapManager swap.Manager

	Net net.MessageSender
}

// NewBackend returns a new Backend
func NewBackend(cfg *Config) (Backend, error) {
	if cfg.Environment == common.Development && cfg.MoneroDaemonEndpoint == "" {
		return nil, errMustProvideDaemonEndpoint
	}

	if cfg.Environment == common.Development {
		defaultTimeoutDuration = time.Minute
	} else if cfg.Environment == common.Stagenet {
		defaultTimeoutDuration = time.Hour
	}

	var (
		addr   ethcommon.Address
		sender txsender.Sender
	)
	if cfg.EthereumPrivateKey != nil {
		txOpts, err := bind.NewKeyedTransactorWithChainID(cfg.EthereumPrivateKey, cfg.ChainID)
		if err != nil {
			return nil, err
		}

		addr = common.EthereumPrivateKeyToAddress(cfg.EthereumPrivateKey)
		sender = txsender.NewSenderWithPrivateKey(cfg.Ctx, cfg.EthereumClient, cfg.SwapContract, txOpts)
	} else {
		log.Debugf("instantiated backend with external sender")
		var err error
		sender, err = txsender.NewExternalSender(cfg.Ctx, cfg.EthereumClient, cfg.SwapContractAddress)
		if err != nil {
			return nil, err
		}
	}

	// monero-wallet-rpc client
	walletClient := monero.NewClient(cfg.MoneroWalletEndpoint)

	// this is only used in the monero development environment to generate new blocks
	var daemonClient monero.DaemonClient
	if cfg.Environment == common.Development {
		daemonClient = monero.NewClient(cfg.MoneroDaemonEndpoint)
	}

	if cfg.SwapContract == nil || (cfg.SwapContractAddress == ethcommon.Address{}) {
		return nil, errNilSwapContractOrAddress
	}

	return &backend{
		ctx:          cfg.Ctx,
		env:          cfg.Environment,
		Client:       walletClient,
		DaemonClient: daemonClient,
		ethClient:    cfg.EthereumClient,
		ethPrivKey:   cfg.EthereumPrivateKey,
		callOpts: &bind.CallOpts{
			From:    addr,
			Context: cfg.Ctx,
		},
		Sender:        sender,
		ethAddress:    addr,
		chainID:       cfg.ChainID,
		gasPrice:      cfg.GasPrice,
		gasLimit:      cfg.GasLimit,
		contract:      cfg.SwapContract,
		contractAddr:  cfg.SwapContractAddress,
		swapManager:   cfg.SwapManager,
		swapTimeout:   defaultTimeoutDuration,
		MessageSender: cfg.Net,
	}, nil
}

func (b *backend) CallOpts() *bind.CallOpts {
	return b.callOpts
}

func (b *backend) ChainID() *big.Int {
	return b.chainID
}

func (b *backend) Contract() *swapfactory.SwapFactory {
	return b.contract
}

func (b *backend) ContractAddr() ethcommon.Address {
	return b.contractAddr
}

func (b *backend) Ctx() context.Context {
	return b.ctx
}

func (b *backend) Env() common.Environment {
	return b.env
}

func (b *backend) EthAddress() ethcommon.Address {
	return b.ethAddress
}

func (b *backend) EthClient() *ethclient.Client {
	return b.ethClient
}

func (b *backend) ExternalSender() *txsender.ExternalSender {
	s, ok := b.Sender.(*txsender.ExternalSender)
	if !ok {
		return nil
	}

	return s
}

func (b *backend) Net() net.MessageSender {
	return b.MessageSender
}

func (b *backend) SwapManager() swap.Manager {
	return b.swapManager
}

func (b *backend) SwapTimeout() time.Duration {
	return b.swapTimeout
}

// SetGasPrice sets the ethereum gas price for the instance to use (in wei).
func (b *backend) SetGasPrice(gasPrice uint64) {
	b.gasPrice = big.NewInt(0).SetUint64(gasPrice)
}

// SetSwapTimeout sets the duration between the swap being initiated on-chain and the timeout t0,
// and the duration between t0 and t1.
func (b *backend) SetSwapTimeout(timeout time.Duration) {
	b.swapTimeout = timeout
}

func (b *backend) BalanceAt(ctx context.Context, account ethcommon.Address, blockNumber *big.Int) (*big.Int, error) {
	return b.ethClient.BalanceAt(ctx, account, blockNumber)
}

func (b *backend) CodeAt(ctx context.Context, account ethcommon.Address, blockNumber *big.Int) ([]byte, error) {
	return b.ethClient.CodeAt(ctx, account, blockNumber)
}

func (b *backend) FilterLogs(ctx context.Context, q eth.FilterQuery) ([]types.Log, error) {
	return b.ethClient.FilterLogs(ctx, q)
}

func (b *backend) TransactionReceipt(ctx context.Context, txHash ethcommon.Hash) (*types.Receipt, error) {
	return b.ethClient.TransactionReceipt(ctx, txHash)
}

func (b *backend) TxOpts() (*bind.TransactOpts, error) {
	txOpts, err := bind.NewKeyedTransactorWithChainID(b.ethPrivKey, b.chainID)
	if err != nil {
		return nil, err
	}

	txOpts.GasPrice = b.gasPrice
	txOpts.GasLimit = b.gasLimit
	return txOpts, nil
}

func (b *backend) XMRDepositAddress() mcrypto.Address {
	return b.xmrDepositAddr
}

// WaitForReceipt waits for the receipt for the given transaction to be available and returns it.
func (b *backend) WaitForReceipt(ctx context.Context, txHash ethcommon.Hash) (*types.Receipt, error) {
	for i := 0; i < maxRetries; i++ {
		receipt, err := b.ethClient.TransactionReceipt(ctx, txHash)
		if err != nil {
			log.Infof("waiting for transaction to be included in chain: txHash=%s", txHash)
			time.Sleep(receiptSleepDuration)
			continue
		}

		log.Infof("transaction %s included in chain, block hash=%s, block number=%d, gas used=%d",
			txHash,
			receipt.BlockHash,
			receipt.BlockNumber,
			receipt.CumulativeGasUsed,
		)
		return receipt, nil
	}

	return nil, errReceiptTimeOut
}

func (b *backend) NewSwapFactory(addr ethcommon.Address) (*swapfactory.SwapFactory, error) {
	return swapfactory.NewSwapFactory(addr, b.ethClient)
}

func (b *backend) SetEthAddress(addr ethcommon.Address) {
	if b.ExternalSender() == nil {
		return
	}

	b.ethAddress = addr
}

func (b *backend) SetXMRDepositAddress(addr mcrypto.Address) {
	b.xmrDepositAddr = addr
}

func (b *backend) SetContract(contract *swapfactory.SwapFactory) {
	b.contract = contract
	b.Sender.SetContract(contract)
}

func (b *backend) SetContractAddress(addr ethcommon.Address) {
	b.contractAddr = addr
}
