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

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/txsender"
	"github.com/athanorlabs/atomic-swap/swapfactory"

	logging "github.com/ipfs/go-log"
)

var (
	log                    = logging.Logger("protocol/backend")
	defaultTimeoutDuration = time.Hour * 24
)

// Backend provides an interface for both the XMRTaker and XMRMaker into the Monero/Ethereum chains.
// It also interfaces with the network layer.
type Backend interface {
	monero.WalletClient
	net.MessageSender

	// NewTxSender creates a new transaction sender, called per-swap
	NewTxSender(asset ethcommon.Address,
		erc20Contract *swapfactory.IERC20) (txsender.Sender, error)

	// ethclient methods
	BalanceAt(ctx context.Context, account ethcommon.Address, blockNumber *big.Int) (*big.Int, error)
	ERC20Info(ctx context.Context, token ethcommon.Address) (name string, symbol string, decimals uint8, err error)
	ERC20BalanceAt(ctx context.Context, token ethcommon.Address, account ethcommon.Address,
		blockNumber *big.Int) (*big.Int, error)
	CodeAt(ctx context.Context, account ethcommon.Address, blockNumber *big.Int) ([]byte, error)
	FilterLogs(ctx context.Context, q eth.FilterQuery) ([]ethtypes.Log, error)
	TransactionByHash(ctx context.Context, hash ethcommon.Hash) (tx *ethtypes.Transaction, isPending bool, err error)
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
	XMRDepositAddress(id *types.Hash) (mcrypto.Address, error)
	HasEthereumPrivateKey() bool
	EthClient() *ethclient.Client

	// setters
	SetSwapTimeout(timeout time.Duration)
	SetGasPrice(uint64)
	SetEthAddress(ethcommon.Address)
	SetXMRDepositAddress(mcrypto.Address, types.Hash)
	ClearXMRDepositAddress(types.Hash)
	SetBaseXMRDepositAddress(mcrypto.Address)
	SetContract(*swapfactory.SwapFactory)
	SetContractAddress(ethcommon.Address)
}

type backend struct {
	ctx         context.Context
	env         common.Environment
	swapManager swap.Manager

	// monero endpoints
	monero.WalletClient

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

	txOpts *txsender.TxOpts

	// swap contract
	contract     *swapfactory.SwapFactory
	contractAddr ethcommon.Address
	swapTimeout  time.Duration

	// network interface
	net.MessageSender
}

// Config is the config for the Backend
type Config struct {
	Ctx                context.Context
	MoneroClient       monero.WalletClient
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
	if cfg.Environment == common.Development {
		defaultTimeoutDuration = 90 * time.Second
	} else if cfg.Environment == common.Stagenet {
		defaultTimeoutDuration = time.Hour
	}

	var (
		addr   ethcommon.Address
		txOpts *txsender.TxOpts
		err    error
	)
	if cfg.EthereumPrivateKey != nil {
		addr = common.EthereumPrivateKeyToAddress(cfg.EthereumPrivateKey)

		// TODO: set gas limit + price based on network (#153)
		txOpts, err = txsender.NewTxOpts(cfg.EthereumPrivateKey, cfg.ChainID)
		if err != nil {
			return nil, err
		}
	}

	if cfg.SwapContract == nil || (cfg.SwapContractAddress == ethcommon.Address{}) {
		return nil, errNilSwapContractOrAddress
	}

	return &backend{
		ctx:          cfg.Ctx,
		env:          cfg.Environment,
		WalletClient: cfg.MoneroClient,
		ethClient:    cfg.EthereumClient,
		ethPrivKey:   cfg.EthereumPrivateKey,
		callOpts: &bind.CallOpts{
			From:    addr,
			Context: cfg.Ctx,
		},
		ethAddress:      addr,
		chainID:         cfg.ChainID,
		gasPrice:        cfg.GasPrice,
		gasLimit:        cfg.GasLimit,
		txOpts:          txOpts,
		contract:        cfg.SwapContract,
		contractAddr:    cfg.SwapContractAddress,
		swapManager:     cfg.SwapManager,
		swapTimeout:     defaultTimeoutDuration,
		MessageSender:   cfg.Net,
		xmrDepositAddrs: make(map[types.Hash]mcrypto.Address),
	}, nil
}

func (b *backend) NewTxSender(asset ethcommon.Address, erc20Contract *swapfactory.IERC20) (txsender.Sender, error) {
	if b.ethPrivKey == nil {
		return txsender.NewExternalSender(b.ctx, b.env, b.ethClient, b.contractAddr, asset)
	}

	sender := txsender.NewSenderWithPrivateKey(b.ctx, b.ethClient, b.contract, erc20Contract, b.txOpts)
	return sender, nil
}

func (b *backend) HasEthereumPrivateKey() bool {
	return b.ethPrivKey != nil
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

func (b *backend) ERC20BalanceAt(ctx context.Context, token ethcommon.Address, account ethcommon.Address,
	blockNumber *big.Int) (*big.Int, error) {
	tokenContract, err := swapfactory.NewIERC20(token, b.ethClient)
	if err != nil {
		return big.NewInt(0), err
	}
	return tokenContract.BalanceOf(b.callOpts, account)
}

func (b *backend) ERC20Info(ctx context.Context, token ethcommon.Address) (name string, symbol string,
	decimals uint8, err error) {
	tokenContract, err := swapfactory.NewIERC20(token, b.ethClient)
	if err != nil {
		return "", "", 18, err
	}
	name, err = tokenContract.Name(b.callOpts)
	if err != nil {
		return "", "", 18, err
	}
	symbol, err = tokenContract.Symbol(b.callOpts)
	if err != nil {
		return "", "", 18, err
	}
	decimals, err = tokenContract.Decimals(b.callOpts)
	if err != nil {
		return "", "", 18, err
	}
	return name, symbol, decimals, nil
}

func (b *backend) CodeAt(ctx context.Context, account ethcommon.Address, blockNumber *big.Int) ([]byte, error) {
	return b.ethClient.CodeAt(ctx, account, blockNumber)
}

func (b *backend) FilterLogs(ctx context.Context, q eth.FilterQuery) ([]ethtypes.Log, error) {
	return b.ethClient.FilterLogs(ctx, q)
}

func (b *backend) TransactionByHash(ctx context.Context, hash ethcommon.Hash) (tx *ethtypes.Transaction, isPending bool, err error) { //nolint:lll
	return b.ethClient.TransactionByHash(ctx, hash)
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
	// only allowed if using external signer
	if b.ethPrivKey != nil {
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
	b.xmrDepositAddrs[id] = addr
}

func (b *backend) ClearXMRDepositAddress(id types.Hash) {
	b.Lock()
	defer b.Unlock()
	delete(b.xmrDepositAddrs, id)
}

// NOTE: this is called when a swap is initiated and the XMR-taker specifies the contract
// address they will be using.
// the contract bytecode is validated in the calling code, but this should never be called
// for unvalidated contracts.
func (b *backend) SetContract(contract *swapfactory.SwapFactory) {
	b.contract = contract
}

func (b *backend) SetContractAddress(addr ethcommon.Address) {
	b.contractAddr = addr
}
