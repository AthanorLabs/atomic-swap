// Package backend provides the portion of top-level swapd instance
// management that is shared by both the maker and the taker.
package backend

import (
	"context"
	"sync"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p/core/peer"

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

// NetSender consists of Host methods invoked by the Maker/Taker
type NetSender interface {
	SendSwapMessage(common.Message, types.Hash) error
	CloseProtocolStream(id types.Hash)
	DiscoverRelayers() ([]peer.ID, error)                                                          // Only used by Maker
	SubmitClaimToRelayer(peer.ID, *message.RelayClaimRequest) (*message.RelayClaimResponse, error) // Only used by Taker
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
	PutCounterpartySwapKeys(id types.Hash, sk *mcrypto.PublicKey, vk *mcrypto.PrivateViewKey) error
	GetCounterpartySwapKeys(id types.Hash) (*mcrypto.PublicKey, *mcrypto.PrivateViewKey, error)
	DeleteSwap(id types.Hash) error
}

// Backend provides an interface for both the XMRTaker and XMRMaker into the Monero/Ethereum chains.
// It also interfaces with the network layer.
type Backend interface {
	XMRClient() monero.WalletClient
	ETHClient() extethclient.EthClient
	NetSender

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
	XMRDepositAddress(offerID *types.Hash) *mcrypto.Address

	// setters
	SetSwapTimeout(timeout time.Duration)
	SetXMRDepositAddress(*mcrypto.Address, types.Hash)
	ClearXMRDepositAddress(types.Hash)
}

type backend struct {
	ctx         context.Context
	env         common.Environment
	swapManager swap.Manager
	recoveryDB  RecoveryDB

	// wallet/node endpoints
	moneroWallet monero.WalletClient
	ethClient    extethclient.EthClient

	// Monero deposit address. When the XMR maker has noTransferBack set to
	// false (default), claimed funds are swept into the primary XMR wallet
	// address used by swapd. This sweep destination address can be overridden
	// on a per-swap basis, by setting an address indexed by the offerID/swapID
	// in the map below.
	perSwapXMRDepositAddrRWMu sync.RWMutex
	perSwapXMRDepositAddr     map[types.Hash]*mcrypto.Address

	// swap contract
	contract     *contracts.SwapFactory
	contractAddr ethcommon.Address
	swapTimeout  time.Duration

	// network interface
	NetSender
}

// Config is the config for the Backend
type Config struct {
	Ctx                context.Context
	MoneroClient       monero.WalletClient
	EthereumClient     extethclient.EthClient
	Environment        common.Environment
	SwapFactoryAddress ethcommon.Address
	SwapManager        swap.Manager
	RecoveryDB         RecoveryDB
	Net                NetSender
}

// NewBackend returns a new Backend
func NewBackend(cfg *Config) (Backend, error) {
	if (cfg.SwapFactoryAddress == ethcommon.Address{}) {
		return nil, errNilSwapContractOrAddress
	}

	swapFactory, err := contracts.NewSwapFactory(cfg.SwapFactoryAddress, cfg.EthereumClient.Raw())
	if err != nil {
		return nil, err
	}

	return &backend{
		ctx:                   cfg.Ctx,
		env:                   cfg.Environment,
		moneroWallet:          cfg.MoneroClient,
		ethClient:             cfg.EthereumClient,
		contract:              swapFactory,
		contractAddr:          cfg.SwapFactoryAddress,
		swapManager:           cfg.SwapManager,
		swapTimeout:           common.SwapTimeoutFromEnv(cfg.Environment),
		NetSender:             cfg.Net,
		perSwapXMRDepositAddr: make(map[types.Hash]*mcrypto.Address),
		recoveryDB:            cfg.RecoveryDB,
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

func (b *backend) NewSwapFactory(addr ethcommon.Address) (*contracts.SwapFactory, error) {
	return contracts.NewSwapFactory(addr, b.ethClient.Raw())
}

// XMRDepositAddress returns the per-swap override deposit address, if a
// per-swap address was set. Otherwise the primary swapd Monero wallet address
// is returned.
func (b *backend) XMRDepositAddress(offerID *types.Hash) *mcrypto.Address {
	b.perSwapXMRDepositAddrRWMu.RLock()
	defer b.perSwapXMRDepositAddrRWMu.RUnlock()

	if offerID != nil {
		addr, ok := b.perSwapXMRDepositAddr[*offerID]
		if ok {
			return addr
		}
	}

	return b.XMRClient().PrimaryAddress()
}

// SetXMRDepositAddress sets a per-swap override deposit address to use when
// sweeping funds out of the shared swap wallet. When noTransferBack is unset
// (default), funds will be swept to this override address instead of to swap's
// primary monero wallet.
func (b *backend) SetXMRDepositAddress(addr *mcrypto.Address, offerID types.Hash) {
	b.perSwapXMRDepositAddrRWMu.Lock()
	defer b.perSwapXMRDepositAddrRWMu.Unlock()
	b.perSwapXMRDepositAddr[offerID] = addr
}

// ClearXMRDepositAddress clears the per-swap, override deposit address from the
// map if a value was set.
func (b *backend) ClearXMRDepositAddress(offerID types.Hash) {
	b.perSwapXMRDepositAddrRWMu.Lock()
	defer b.perSwapXMRDepositAddrRWMu.Unlock()
	delete(b.perSwapXMRDepositAddr, offerID)
}
