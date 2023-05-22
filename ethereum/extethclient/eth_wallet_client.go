// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package extethclient provides libraries for interacting with an ethereum node
// using a specific private key.
package extethclient

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
)

var log = logging.Logger("extethclient")

// EthClient provides management of a private key and other convenience functions layered
// on top of the go-ethereum client. You can still access the raw go-ethereum client via
// the Raw() method.
type EthClient interface {
	Address() ethcommon.Address
	SetAddress(addr ethcommon.Address)
	PrivateKey() *ecdsa.PrivateKey
	HasPrivateKey() bool
	Endpoint() string

	Balance(ctx context.Context) (*coins.WeiAmount, error)
	ERC20Balance(ctx context.Context, token ethcommon.Address) (*coins.ERC20TokenAmount, error)

	ERC20Info(ctx context.Context, tokenAddr ethcommon.Address) (*coins.ERC20TokenInfo, error)

	SetGasPrice(uint64)
	SetGasLimit(uint64)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	CallOpts(ctx context.Context) *bind.CallOpts
	TxOpts(ctx context.Context) (*bind.TransactOpts, error)
	ChainID() *big.Int
	Lock()   // Lock the wallet so only one transaction runs at at time
	Unlock() // Unlock the wallet after a transaction is complete

	// transfers ETH to the given address
	// does not need locking, as it locks internally
	Transfer(ctx context.Context, to ethcommon.Address, amount *coins.WeiAmount) (ethcommon.Hash, error)

	// attempts to cancel a transaction with the given nonce by sending a zero-value tx to ourselves
	CancelTxWithNonce(ctx context.Context, nonce uint64, gasPrice *big.Int) (ethcommon.Hash, error)

	WaitForReceipt(ctx context.Context, txHash ethcommon.Hash) (*ethtypes.Receipt, error)
	WaitForTimestamp(ctx context.Context, ts time.Time) error
	LatestBlockTimestamp(ctx context.Context) (time.Time, error)

	Close()
	Raw() *ethclient.Client
}

type ethClient struct {
	endpoint   string
	ec         *ethclient.Client
	ethPrivKey *ecdsa.PrivateKey
	ethAddress ethcommon.Address
	gasPrice   *big.Int
	gasLimit   uint64
	chainID    *big.Int
	mu         sync.Mutex
}

// NewEthClient creates and returns our extended ethereum client/wallet. The passed context
// is only used for creation. The privKey can be nil if you are using an external signer.
func NewEthClient(
	ctx context.Context,
	env common.Environment,
	endpoint string,
	privKey *ecdsa.PrivateKey,
) (EthClient, error) {
	ec, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, err
	}

	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	if err = validateChainID(env, chainID); err != nil {
		return nil, err
	}

	var addr ethcommon.Address
	if privKey != nil {
		addr = common.EthereumPrivateKeyToAddress(privKey)
	}

	return &ethClient{
		endpoint:   endpoint,
		ec:         ec,
		ethPrivKey: privKey,
		ethAddress: addr,
		chainID:    chainID,
	}, nil
}

func (c *ethClient) Address() ethcommon.Address {
	return c.ethAddress
}

func (c *ethClient) SetAddress(addr ethcommon.Address) {
	if c.HasPrivateKey() {
		panic("SetAddress should not have been invoked when using an external signer")
	}
	c.ethAddress = addr
}

func (c *ethClient) PrivateKey() *ecdsa.PrivateKey {
	return c.ethPrivKey
}

func (c *ethClient) HasPrivateKey() bool {
	return c.ethPrivKey != nil
}

// Endpoint returns the endpoint URL that we are connected to
func (c *ethClient) Endpoint() string {
	return c.endpoint
}

func (c *ethClient) Balance(ctx context.Context) (*coins.WeiAmount, error) {
	addr := c.Address()
	bal, err := c.ec.BalanceAt(ctx, addr, nil)
	if err != nil {
		return nil, err
	}
	return coins.NewWeiAmount(bal), nil
}

// SuggestGasPrice returns the underlying eth client's suggested gas price
// unless the user specified a fixed gas price to use, in which case the user
// supplied value is returned.
func (c *ethClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	if c.gasPrice != nil {
		return c.gasPrice, nil
	}
	return c.Raw().SuggestGasPrice(ctx)
}

func (c *ethClient) ERC20Balance(ctx context.Context, tokenAddr ethcommon.Address) (*coins.ERC20TokenAmount, error) {
	tokenContract, err := contracts.NewIERC20(tokenAddr, c.ec)
	if err != nil {
		return nil, err
	}

	bal, err := tokenContract.BalanceOf(c.CallOpts(ctx), c.Address())
	if err != nil {
		return nil, err
	}

	tokenInfo, err := c.erc20Info(ctx, tokenAddr, tokenContract)
	if err != nil {
		return nil, err
	}

	return coins.NewERC20TokenAmountFromBigInt(bal, tokenInfo), nil
}

func (c *ethClient) erc20Info(
	ctx context.Context,
	tokenAddr ethcommon.Address,
	tokenContract *contracts.IERC20,
) (*coins.ERC20TokenInfo, error) {
	name, err := tokenContract.Name(c.CallOpts(ctx))
	if err != nil {
		return nil, err
	}

	symbol, err := tokenContract.Symbol(c.CallOpts(ctx))
	if err != nil {
		return nil, err
	}

	// TODO: Do we support ERC20 tokens that do not have this method?
	decimals, err := tokenContract.Decimals(c.CallOpts(ctx))
	if err != nil {
		return nil, err
	}

	return coins.NewERC20TokenInfo(tokenAddr, decimals, name, symbol), nil
}

func (c *ethClient) ERC20Info(ctx context.Context, tokenAddr ethcommon.Address) (*coins.ERC20TokenInfo, error) {
	tokenContract, err := contracts.NewIERC20(tokenAddr, c.ec)
	if err != nil {
		return nil, err
	}

	return c.erc20Info(ctx, tokenAddr, tokenContract)
}

// SetGasPrice sets the ethereum gas price (in wei) for use in transactions. In most
// cases, you should not use this function and let the ethereum client determine the
// suggested gas price at the current time. Setting a value of zero reverts to using
// the raw ethereum client's suggested price.
func (c *ethClient) SetGasPrice(gasPrice uint64) {
	if gasPrice == 0 {
		c.gasPrice = nil
		return
	}
	c.gasPrice = new(big.Int).SetUint64(gasPrice)
}

// SetGasLimit sets the ethereum gas limit to use (in wei). In most cases you should not
// use this function and let the ethereum client dynamically determine the gas limit based
// on a simulation of the contract transaction.
func (c *ethClient) SetGasLimit(gasLimit uint64) {
	c.gasLimit = gasLimit
}

func (c *ethClient) CallOpts(ctx context.Context) *bind.CallOpts {
	return &bind.CallOpts{
		Pending:     false,
		From:        c.ethAddress, // might be all zeros if using external signer
		BlockNumber: nil,
		Context:     ctx,
	}
}

func (c *ethClient) TxOpts(ctx context.Context) (*bind.TransactOpts, error) {
	if !c.HasPrivateKey() {
		panic("TxOpts() should not have been invoked when using an external signer")
	}

	txOpts, err := bind.NewKeyedTransactorWithChainID(c.ethPrivKey, c.chainID)
	if err != nil {
		return nil, err
	}
	txOpts.Context = ctx

	// TODO: set gas limit + price based on network (#153)
	txOpts.GasPrice = c.gasPrice
	txOpts.GasLimit = c.gasLimit

	return txOpts, nil
}

func (c *ethClient) ChainID() *big.Int {
	return c.chainID
}

// WaitForReceipt waits for the receipt for the given transaction to be available and returns it.
func (c *ethClient) WaitForReceipt(ctx context.Context, txHash ethcommon.Hash) (*ethtypes.Receipt, error) {
	return block.WaitForReceipt(ctx, c.ec, txHash)
}

func (c *ethClient) WaitForTimestamp(ctx context.Context, ts time.Time) error {
	hdr, err := block.WaitForEthBlockAfterTimestamp(ctx, c.ec, ts)
	if err != nil {
		return err
	}

	log.Debugf("Wait complete for block %d with ts=%s >= %s",
		hdr.Number.Uint64(),
		time.Unix(int64(hdr.Time), 0).Format(common.TimeFmtSecs),
		ts.Format(common.TimeFmtSecs),
	)
	return nil
}

func (c *ethClient) LatestBlockTimestamp(ctx context.Context) (time.Time, error) {
	hdr, err := c.ec.HeaderByNumber(ctx, nil)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(int64(hdr.Time), 0), nil
}

func (c *ethClient) Lock() {
	c.mu.Lock()
}

func (c *ethClient) Unlock() {
	c.mu.Unlock()
}

func (c *ethClient) Close() {
	c.ec.Close()
}

func (c *ethClient) Raw() *ethclient.Client {
	return c.ec
}

func (c *ethClient) CancelTxWithNonce(
	ctx context.Context,
	nonce uint64,
	gasPrice *big.Int,
) (ethcommon.Hash, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tx := ethtypes.NewTransaction(nonce, c.ethAddress, big.NewInt(0), 21000, gasPrice, nil)

	signer := ethtypes.LatestSignerForChainID(c.chainID)
	signedTx, err := ethtypes.SignTx(tx, signer, c.ethPrivKey)
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to sign tx: %w", err)
	}

	err = c.ec.SendTransaction(ctx, signedTx)
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to send tx: %w", err)
	}

	return signedTx.Hash(), nil
}

func (c *ethClient) Transfer(
	ctx context.Context,
	to ethcommon.Address,
	amount *coins.WeiAmount,
) (ethcommon.Hash, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	nonce, err := c.ec.NonceAt(ctx, c.ethAddress, nil)
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := c.ec.SuggestGasPrice(ctx)
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to get gas price: %w", err)
	}

	tx := ethtypes.NewTransaction(nonce, to, amount.BigInt(), 21000, gasPrice, nil)

	signer := ethtypes.LatestSignerForChainID(c.chainID)
	signedTx, err := ethtypes.SignTx(tx, signer, c.ethPrivKey)
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to sign tx: %w", err)
	}

	err = c.ec.SendTransaction(ctx, signedTx)
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx.Hash(), nil
}

func validateChainID(env common.Environment, chainID *big.Int) error {
	switch env {
	case common.Mainnet:
		if chainID.Cmp(big.NewInt(common.MainnetChainID)) != 0 {
			return fmt.Errorf("expected Mainnet chain ID (%d), but found %s", common.MainnetChainID, chainID)
		}
	case common.Stagenet:
		if chainID.Cmp(big.NewInt(common.SepoliaChainID)) != 0 {
			return fmt.Errorf("expected Sepolia chain ID (%d), but found %s", common.SepoliaChainID, chainID)
		}
	case common.Development:
		if chainID.Cmp(big.NewInt(common.GanacheChainID)) != 0 {
			return fmt.Errorf("expected Ganache chain ID (%d), but found %s", common.GanacheChainID, chainID)
		}
	default:
		panic("unhandled environment type")
	}

	return nil
}
