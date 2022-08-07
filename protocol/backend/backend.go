package backend

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"sync"
	"time"

	eth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	"github.com/noot/atomic-swap/ethereum/block"
	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/protocol/swap"
	"github.com/noot/atomic-swap/protocol/txsender"
	"github.com/noot/atomic-swap/swapfactory"

	logging "github.com/ipfs/go-log"
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
	FilterLogs(ctx context.Context, q eth.FilterQuery) ([]ethtypes.Log, error)
	TransactionReceipt(ctx context.Context, txHash ethcommon.Hash) (*ethtypes.Receipt, error)
	WaitForTimestamp(ctx context.Context, ts time.Time) error
	LatestBlockTimestamp(ctx context.Context) (time.Time, error)

	// helpers
	WaitForReceipt(ctx context.Context, txHash ethcommon.Hash) (*ethtypes.Receipt, error)
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
	XMRDepositAddress(id *types.Hash) (mcrypto.Address, error)

	// setters
	SetSwapTimeout(timeout time.Duration)
	SetGasPrice(uint64)
	SetEthAddress(ethcommon.Address)
	SetXMRDepositAddress(mcrypto.Address, types.Hash)
	SetBaseXMRDepositAddress(mcrypto.Address)
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
	sync.RWMutex
	baseXMRDepositAddr *mcrypto.Address
	xmrDepositAddrs    map[types.Hash]mcrypto.Address

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
		sender, err = txsender.NewExternalSender(cfg.Ctx, cfg.Environment, cfg.EthereumClient, cfg.SwapContractAddress)
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
		Sender:          sender,
		ethAddress:      addr,
		chainID:         cfg.ChainID,
		gasPrice:        cfg.GasPrice,
		gasLimit:        cfg.GasLimit,
		contract:        cfg.SwapContract,
		contractAddr:    cfg.SwapContractAddress,
		swapManager:     cfg.SwapManager,
		swapTimeout:     defaultTimeoutDuration,
		MessageSender:   cfg.Net,
		xmrDepositAddrs: make(map[types.Hash]mcrypto.Address),
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

func (b *backend) FilterLogs(ctx context.Context, q eth.FilterQuery) ([]ethtypes.Log, error) {
	return b.ethClient.FilterLogs(ctx, q)
}

func (b *backend) TransactionReceipt(ctx context.Context, txHash ethcommon.Hash) (*ethtypes.Receipt, error) {
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

func (b *backend) XMRDepositAddress(id *types.Hash) (mcrypto.Address, error) {
	b.RLock()
	defer b.RUnlock()

	if id == nil && b.baseXMRDepositAddr == nil {
		return "", errNoXMRDepositAddress
	} else if id == nil {
		return *b.baseXMRDepositAddr, nil
	}

	addr, has := b.xmrDepositAddrs[*id]
	if !has && b.baseXMRDepositAddr == nil {
		return "", errNoXMRDepositAddress
	} else if !has {
		return *b.baseXMRDepositAddr, nil
	}

	return addr, nil
}

// WaitForReceipt waits for the receipt for the given transaction to be available and returns it.
func (b *backend) WaitForReceipt(ctx context.Context, txHash ethcommon.Hash) (*ethtypes.Receipt, error) {
	return block.WaitForReceipt(ctx, b.ethClient, txHash)
}

func (b *backend) WaitForTimestamp(ctx context.Context, ts time.Time) error {
	hdr, err := block.WaitForEthBlockAfterTimestamp(ctx, b.ethClient, ts.Unix())
	if err != nil {
		return err
	}
	log.Debug("Wait complete for block %d with ts=%s >= %s",
		hdr.Number.Uint64(),
		time.Unix(int64(hdr.Time), 0).Format(common.TimeFmtSecs),
		ts.Format(common.TimeFmtSecs),
	)
	return nil
}

func (b *backend) LatestBlockTimestamp(ctx context.Context) (time.Time, error) {
	hdr, err := b.EthClient().HeaderByNumber(ctx, nil)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(int64(hdr.Time), 0), nil
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

func (b *backend) SetBaseXMRDepositAddress(addr mcrypto.Address) {
	b.baseXMRDepositAddr = &addr
}

func (b *backend) SetXMRDepositAddress(addr mcrypto.Address, id types.Hash) {
	b.Lock()
	defer b.Unlock()
	// TODO: clear this out when swap is done, memory leak!!!
	b.xmrDepositAddrs[id] = addr
}

// TODO: these are kinda sus, maybe remove them? forces everyone to use
// the same contract though
func (b *backend) SetContract(contract *swapfactory.SwapFactory) {
	b.contract = contract
	b.Sender.SetContract(contract)
}

func (b *backend) SetContractAddress(addr ethcommon.Address) {
	b.contractAddr = addr
}
