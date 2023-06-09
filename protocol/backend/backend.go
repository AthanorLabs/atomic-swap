// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package backend provides the portion of top-level swapd instance
// management that is shared by both the maker and the taker.
package backend

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/coins"
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
	"github.com/athanorlabs/atomic-swap/relayer"
)

// NetSender consists of Host methods invoked by the Maker/Taker
type NetSender interface {
	SendSwapMessage(common.Message, types.Hash) error
	DeleteOngoingSwap(offerID types.Hash)
	CloseProtocolStream(id types.Hash)
	DiscoverRelayers() ([]peer.ID, error)                                                        // Only used by Maker
	QueryRelayerAddress(peer.ID) (types.Hash, error)                                             // only used by taker
	SubmitRelayRequest(peer.ID, *message.RelayClaimRequest) (*message.RelayClaimResponse, error) // only used by taker
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
	PutNewSwapTxHash(id types.Hash, txHash types.Hash) error
	GetNewSwapTxHash(id types.Hash) (types.Hash, error)
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
	NewSwapCreator(addr ethcommon.Address) (*contracts.SwapCreator, error)
	HandleRelayClaimRequest(remotePeer peer.ID, request *message.RelayClaimRequest) (*message.RelayClaimResponse, error)
	GetRelayerAddressHash() (types.Hash, error)
	HasOngoingSwapAsTaker(peer.ID) error
	SubmitClaimToRelayer(
		peer.ID,
		*types.Hash,
		*contracts.SwapCreatorRelaySwap,
		[32]byte,
	) (*message.RelayClaimResponse, error) // Only used by Taker

	// getters
	Ctx() context.Context
	Env() common.Environment
	SwapManager() swap.Manager
	SwapCreator() *contracts.SwapCreator
	SwapCreatorAddr() ethcommon.Address
	SwapTimeout() time.Duration
	XMRDepositAddress(offerID *types.Hash) *mcrypto.Address

	// setters
	SetSwapTimeout(timeout time.Duration)
	SetXMRDepositAddress(*mcrypto.Address, types.Hash)
	ClearXMRDepositAddress(types.Hash)

	// transfer helpers
	TransferXMR(to *mcrypto.Address, amount *coins.PiconeroAmount) (string, error)
	SweepXMR(to *mcrypto.Address) ([]string, error)
	TransferETH(to ethcommon.Address, amount *coins.WeiAmount, gasLimit *uint64) (*ethtypes.Receipt, error)
	SweepETH(to ethcommon.Address) (*ethtypes.Receipt, error)
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
	swapCreator     *contracts.SwapCreator
	swapCreatorAddr ethcommon.Address
	swapTimeout     time.Duration

	// network interface
	NetSender

	// map of hash(relayer address || salt) -> salt
	relayerHashMu sync.RWMutex
	relayerHash   map[types.Hash][4]byte
}

// Config is the config for the Backend
type Config struct {
	Ctx             context.Context
	MoneroClient    monero.WalletClient
	EthereumClient  extethclient.EthClient
	Environment     common.Environment
	SwapCreatorAddr ethcommon.Address
	SwapManager     swap.Manager
	RecoveryDB      RecoveryDB
	Net             NetSender
}

// NewBackend returns a new Backend
func NewBackend(cfg *Config) (Backend, error) {
	if (cfg.SwapCreatorAddr == ethcommon.Address{}) {
		return nil, errNilSwapContractOrAddress
	}

	swapCreator, err := contracts.NewSwapCreator(cfg.SwapCreatorAddr, cfg.EthereumClient.Raw())
	if err != nil {
		return nil, err
	}

	return &backend{
		ctx:                   cfg.Ctx,
		env:                   cfg.Environment,
		moneroWallet:          cfg.MoneroClient,
		ethClient:             cfg.EthereumClient,
		swapCreator:           swapCreator,
		swapCreatorAddr:       cfg.SwapCreatorAddr,
		swapManager:           cfg.SwapManager,
		swapTimeout:           common.SwapTimeoutFromEnv(cfg.Environment),
		NetSender:             cfg.Net,
		perSwapXMRDepositAddr: make(map[types.Hash]*mcrypto.Address),
		recoveryDB:            cfg.RecoveryDB,
		relayerHash:           make(map[types.Hash][4]byte),
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
		return txsender.NewExternalSender(b.ctx, b.env, b.ethClient.Raw(), b.swapCreatorAddr, asset)
	}

	return txsender.NewSenderWithPrivateKey(b.ctx, b.ETHClient(), b.swapCreatorAddr, b.swapCreator, erc20Contract), nil
}

func (b *backend) RecoveryDB() RecoveryDB {
	return b.recoveryDB
}

func (b *backend) SwapCreator() *contracts.SwapCreator {
	return b.swapCreator
}

func (b *backend) SwapCreatorAddr() ethcommon.Address {
	return b.swapCreatorAddr
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

// SetSwapTimeout sets the duration between the swap being initiated on-chain and the timeout t1,
// and the duration between t1 and t2.
func (b *backend) SetSwapTimeout(timeout time.Duration) {
	b.swapTimeout = timeout
}

func (b *backend) NewSwapCreator(addr ethcommon.Address) (*contracts.SwapCreator, error) {
	return contracts.NewSwapCreator(addr, b.ethClient.Raw())
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

// HasOngoingSwapAsTaker returns nil if we have an ongoing swap with the given peer where
// we're the xmrtaker, otherwise returns an error.
func (b *backend) HasOngoingSwapAsTaker(remotePeer peer.ID) error {
	swaps, err := b.swapManager.GetOngoingSwapsSnapshot()
	if err != nil {
		return err
	}

	for _, swap := range swaps {
		if swap.PeerID != remotePeer {
			continue
		}

		if swap.IsTaker() {
			return nil
		}
	}

	return fmt.Errorf("do not have an ongoing swap with peer %s as taker", remotePeer)
}

// HandleRelayClaimRequest validates and sends the transaction for a relay claim request
func (b *backend) HandleRelayClaimRequest(
	remotePeer peer.ID,
	request *message.RelayClaimRequest,
) (*message.RelayClaimResponse, error) {
	defer b.clearRelayerAddressHash(request.RelaySwap.RelayerHash)

	if request.OfferID != nil {
		has := b.swapManager.HasOngoingSwap(*request.OfferID)
		if !has {
			return nil, fmt.Errorf("cannot relay taker-specific claim request; no ongoing swap for swap %s", *request.OfferID)
		}

		info, err := b.swapManager.GetOngoingSwapSnapshot(*request.OfferID)
		if err != nil {
			return nil, err
		}

		if !info.IsTaker() {
			return nil, fmt.Errorf("cannot relay taker-specific claim request; not the xmr-taker for swap %s", *request.OfferID)
		}

		if remotePeer != info.PeerID {
			return nil, fmt.Errorf("cannot relay taker-specific claim request from peer %s; unexpected peer for swap %s",
				remotePeer, *request.OfferID)
		}
	}

	// In the taker relay scenario, the net layer has already validated that we
	// have an ongoing swap with the requesting peer that uses the passed
	// offerID, but we have not verified that the claim in the swap matches the
	// offerID. The backend, with its access to the recovery DB, is in the best
	// position to perform this check. The remaining validations will be in the
	// relayer library.
	if request.OfferID != nil {
		swapInfo, err := b.recoveryDB.GetContractSwapInfo(*request.OfferID)
		if err != nil {
			return nil, fmt.Errorf("swap info for taker claim request not found: %w", err)
		}
		if swapInfo.SwapID != request.RelaySwap.Swap.SwapID() {
			return nil, errors.New("counterparty claim request has invalid swap ID")
		}
	}

	b.relayerHashMu.RLock()
	salt := b.relayerHash[request.RelaySwap.RelayerHash]
	b.relayerHashMu.RUnlock()

	return relayer.ValidateAndSendTransaction(
		b.Ctx(),
		request,
		b.ETHClient(),
		b.SwapCreatorAddr(),
		salt,
	)
}

func (b *backend) GetRelayerAddressHash() (types.Hash, error) {
	address := b.ETHClient().Address()
	var salt [4]byte
	_, err := rand.Read(salt[:])
	if err != nil {
		return types.Hash{}, err
	}

	hash := crypto.Keccak256Hash(append(address.Bytes(), salt[:]...))
	b.relayerHashMu.Lock()
	defer b.relayerHashMu.Unlock()
	b.relayerHash[hash] = salt
	return hash, nil
}

func (b *backend) clearRelayerAddressHash(hash types.Hash) {
	b.relayerHashMu.Lock()
	defer b.relayerHashMu.Unlock()
	delete(b.relayerHash, hash)
}

func (b *backend) SubmitClaimToRelayer(
	relayerID peer.ID,
	offerID *types.Hash,
	relaySwap *contracts.SwapCreatorRelaySwap,
	secret [32]byte,
) (*message.RelayClaimResponse, error) {
	// get the relayer's address hash
	relayerAddrHash, err := b.QueryRelayerAddress(relayerID)
	if err != nil {
		return nil, err
	}

	// set relayer address hash and sign as front-run prevention
	relaySwap.RelayerHash = relayerAddrHash

	req, err := relayer.CreateRelayClaimRequest(b.ETHClient().PrivateKey(), relaySwap, secret)
	if err != nil {
		return nil, err
	}

	if offerID != nil {
		req.OfferID = offerID
	}

	return b.SubmitRelayRequest(relayerID, req)
}

func (b *backend) TransferXMR(to *mcrypto.Address, amount *coins.PiconeroAmount) (string, error) {
	res, err := b.moneroWallet.Transfer(b.ctx, to, 0, amount, 1)
	if err != nil {
		return "", err
	}

	return res.TxID, nil

}

func (b *backend) SweepXMR(to *mcrypto.Address) ([]string, error) {
	res, err := b.moneroWallet.SweepAll(b.ctx, to, 0, 1)
	if err != nil {
		return nil, err
	}

	txIDs := make([]string, len(res))
	for i, transfer := range res {
		txIDs[i] = transfer.TxID
	}

	return txIDs, nil
}

func (b *backend) TransferETH(
	to ethcommon.Address,
	amount *coins.WeiAmount,
	gasLimit *uint64,
) (*ethtypes.Receipt, error) {
	return b.ethClient.Transfer(b.ctx, to, amount, gasLimit)
}

func (b *backend) SweepETH(to ethcommon.Address) (*ethtypes.Receipt, error) {
	return b.ethClient.Sweep(b.ctx, to)
}
