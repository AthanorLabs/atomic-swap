// Package backend provides the portion of top-level swapd instance
// management that is shared by both the maker and the taker.
package backend

import (
	"context"
	"sync"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p/core/peer"

	rcommon "github.com/athanorlabs/go-relayer/common"
	rnet "github.com/athanorlabs/go-relayer/net"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/db"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net/message"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/txsender"
)

// MessageSender is implemented by a Host
type MessageSender interface {
	SendSwapMessage(message.Message, types.Hash) error
	CloseProtocolStream(id types.Hash)
}

// RelayerHost contains required network functionality for discovering
// and messaging relayers.
type RelayerHost interface {
	Discover(time.Duration) ([]peer.ID, error)
	SubmitTransaction(who peer.ID, msg *rnet.TransactionRequest) (*rcommon.SubmitTransactionResponse, error)
}

// RecoveryDB is implemented by *db.RecoveryDB
type RecoveryDB interface {
	PutContractSwapInfo(id types.Hash, info *db.EthereumSwapInfo) error
	GetContractSwapInfo(id types.Hash) (*db.EthereumSwapInfo, error)
	PutSwapPrivateKey(id types.Hash, keys *mcrypto.PrivateSpendKey) error
	GetSwapPrivateKey(id types.Hash) (*mcrypto.PrivateSpendKey, error)
	PutCounterpartySwapPrivateKey(id types.Hash, keys *mcrypto.PrivateSpendKey) error
	GetCounterpartySwapPrivateKey(id types.Hash) (*mcrypto.PrivateSpendKey, error)
	PutSwapRelayerInfo(id types.Hash, info *types.OfferExtra) error
	GetSwapRelayerInfo(id types.Hash) (*types.OfferExtra, error)
	PutXMRMakerSwapKeys(id types.Hash, sk *mcrypto.PublicKey, vk *mcrypto.PrivateViewKey) error
	GetXMRMakerSwapKeys(id types.Hash) (*mcrypto.PublicKey, *mcrypto.PrivateViewKey, error)
	DeleteSwap(id types.Hash) error
}

// Backend provides an interface for both the XMRTaker and XMRMaker into the Monero/Ethereum chains.
// It also interfaces with the network layer.
type Backend interface {
	XMRClient() monero.WalletClient
	ETHClient() extethclient.EthClient
	MessageSender

	RecoveryDB() RecoveryDB

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
	SwapTimeout() time.Duration
	XMRDepositAddress(id *types.Hash) (mcrypto.Address, error)

	// setters
	SetSwapTimeout(timeout time.Duration)
	SetXMRDepositAddress(mcrypto.Address, types.Hash)
	ClearXMRDepositAddress(types.Hash)
	SetBaseXMRDepositAddress(mcrypto.Address)

	// relayer functions
	DiscoverRelayers() ([]peer.ID, error)
	SubmitTransactionToRelayer(peer.ID, *rcommon.SubmitTransactionRequest) (*rcommon.SubmitTransactionResponse, error)
}

type backend struct {
	ctx         context.Context
	env         common.Environment
	swapManager swap.Manager
	recoveryDB  RecoveryDB

	// wallet/node endpoints
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
	MessageSender

	// relayer network interface
	rnet RelayerHost
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

	RecoveryDB RecoveryDB

	Net         MessageSender
	RelayerHost RelayerHost
}

// NewBackend returns a new Backend
func NewBackend(cfg *Config) (Backend, error) {
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
		swapTimeout:     common.SwapTimeoutFromEnv(cfg.Environment),
		MessageSender:   cfg.Net,
		xmrDepositAddrs: make(map[types.Hash]mcrypto.Address),
		recoveryDB:      cfg.RecoveryDB,
		rnet:            cfg.RelayerHost,
	}, nil
}

func (b *backend) XMRClient() monero.WalletClient {
	return b.moneroWallet
}

func (b *backend) ETHClient() extethclient.EthClient {
	return b.ethClient
}

func (b *backend) NewTxSender(asset ethcommon.Address, erc20Contract *contracts.IERC20) (txsender.Sender, error) {
	if !b.ethClient.HasPrivateKey() {
		return txsender.NewExternalSender(b.ctx, b.env, b.ethClient.Raw(), b.contractAddr, asset)
	}

	return txsender.NewSenderWithPrivateKey(b.ctx, b.ETHClient(), b.contract, erc20Contract), nil
}

func (b *backend) RecoveryDB() RecoveryDB {
	return b.recoveryDB
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

func (b *backend) DiscoverRelayers() ([]peer.ID, error) {
	const defaultDiscoverTime = time.Second * 3
	return b.rnet.Discover(defaultDiscoverTime)
}

func (b *backend) SubmitTransactionToRelayer(
	to peer.ID,
	req *rcommon.SubmitTransactionRequest,
) (*rcommon.SubmitTransactionResponse, error) {
	msg := &rnet.TransactionRequest{
		SubmitTransactionRequest: *req,
	}

	return b.rnet.SubmitTransaction(to, msg)
}
