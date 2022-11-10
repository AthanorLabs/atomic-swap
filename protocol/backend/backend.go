package backend

import (
	"context"
	"sync"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/txsender"
)

var (
	defaultTimeoutDuration = time.Hour * 24
)

// Backend provides an interface for both the XMRTaker and XMRMaker into the Monero/Ethereum chains.
// It also interfaces with the network layer.
type Backend interface {
	XMR() monero.WalletClient
	ETH() extethclient.EthClient
	net.MessageSender

	// NewTxSender creates a new transaction sender, called per-swap
	NewTxSender(asset ethcommon.Address, erc20Contract *contracts.IERC20) (txsender.Sender, error)

	// helpers
	NewSwapFactory(addr ethcommon.Address) (*contracts.SwapFactory, error)

	// getters
	Ctx() context.Context
	Env() common.Environment
	SwapManager() swap.Manager
	Contract() *contracts.SwapFactory
	ContractAddr() ethcommon.Address
	Net() net.MessageSender
	SwapTimeout() time.Duration
	XMRDepositAddress(id *types.Hash) (mcrypto.Address, error)

	// setters
	SetSwapTimeout(timeout time.Duration)
	SetXMRDepositAddress(mcrypto.Address, types.Hash)
	ClearXMRDepositAddress(types.Hash)
	SetBaseXMRDepositAddress(mcrypto.Address)
}

type backend struct {
	ctx         context.Context
	env         common.Environment
	swapManager swap.Manager

	// monero endpoints
	moneroWallet monero.WalletClient
	ethClient    extethclient.EthClient

	// monero deposit address (used if xmrtaker has transferBack set to true)
	sync.RWMutex
	baseXMRDepositAddr *mcrypto.Address
	xmrDepositAddrs    map[types.Hash]mcrypto.Address

	// swap contract
	contract     *contracts.SwapFactory
	contractAddr ethcommon.Address
	swapTimeout  time.Duration

	// network interface
	net.MessageSender
}

// Config is the config for the Backend
type Config struct {
	Ctx            context.Context
	MoneroClient   monero.WalletClient
	EthereumClient extethclient.EthClient
	Environment    common.Environment

	SwapContract        *contracts.SwapFactory
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

	if cfg.SwapContract == nil || (cfg.SwapContractAddress == ethcommon.Address{}) {
		return nil, errNilSwapContractOrAddress
	}

	return &backend{
		ctx:             cfg.Ctx,
		env:             cfg.Environment,
		moneroWallet:    cfg.MoneroClient,
		ethClient:       cfg.EthereumClient,
		contract:        cfg.SwapContract,
		contractAddr:    cfg.SwapContractAddress,
		swapManager:     cfg.SwapManager,
		swapTimeout:     defaultTimeoutDuration,
		MessageSender:   cfg.Net,
		xmrDepositAddrs: make(map[types.Hash]mcrypto.Address),
	}, nil
}

func (b *backend) XMR() monero.WalletClient {
	return b.moneroWallet
}

func (b *backend) ETH() extethclient.EthClient {
	return b.ethClient
}

func (b *backend) NewTxSender(asset ethcommon.Address, erc20Contract *contracts.IERC20) (txsender.Sender, error) {
	ec := b.ethClient.Raw()

	if !b.ethClient.HasPrivateKey() {
		return txsender.NewExternalSender(b.ctx, b.env, ec, b.contractAddr, asset)
	}

	wrappedTxOpts, err := txsender.NewTxOpts(b.ethClient.PrivateKey(), b.ethClient.ChainID())
	if err != nil {
		return nil, err
	}
	return txsender.NewSenderWithPrivateKey(b.ctx, ec, b.contract, erc20Contract, wrappedTxOpts), nil
}

func (b *backend) Contract() *contracts.SwapFactory {
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

func (b *backend) Net() net.MessageSender {
	return b.MessageSender
}

func (b *backend) SwapManager() swap.Manager {
	return b.swapManager
}

func (b *backend) SwapTimeout() time.Duration {
	return b.swapTimeout
}

// SetSwapTimeout sets the duration between the swap being initiated on-chain and the timeout t0,
// and the duration between t0 and t1.
func (b *backend) SetSwapTimeout(timeout time.Duration) {
	b.swapTimeout = timeout
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

func (b *backend) NewSwapFactory(addr ethcommon.Address) (*contracts.SwapFactory, error) {
	return contracts.NewSwapFactory(addr, b.ethClient.Raw())
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
