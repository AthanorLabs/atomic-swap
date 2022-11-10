package backend

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/athanorlabs/atomic-swap/common"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
)

// EthClient is the lower level ethereum network and wallet operations used by the protocol backend
type EthClient interface {
	Address() ethcommon.Address
	SetAddress(addr ethcommon.Address)
	PrivateKey() *ecdsa.PrivateKey
	HasPrivateKey() bool

	Balance(ctx context.Context) (ethcommon.Address, *big.Int, error)
	ERC20BalanceAt(ctx context.Context, token ethcommon.Address, account ethcommon.Address,
		blockNumber *big.Int) (*big.Int, error)

	ERC20Info(ctx context.Context, token ethcommon.Address) (name string, symbol string, decimals uint8, err error)

	SetGasPrice(uint64)
	CallOpts(ctx context.Context) *bind.CallOpts
	TxOpts(ctx context.Context) (*bind.TransactOpts, error)
	ChainID() *big.Int

	WaitForReceipt(ctx context.Context, txHash ethcommon.Hash) (*ethtypes.Receipt, error)
	WaitForTimestamp(ctx context.Context, ts time.Time) error
	LatestBlockTimestamp(ctx context.Context) (time.Time, error)

	Raw() *ethclient.Client
}

type swapEthClient struct {
	ec         *ethclient.Client
	ethPrivKey *ecdsa.PrivateKey
	ethAddress ethcommon.Address
	gasPrice   *big.Int
	gasLimit   uint64
	chainID    *big.Int
}

func newSwapEthClient(ctx context.Context, ec *ethclient.Client, privKey *ecdsa.PrivateKey) (EthClient, error) {
	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	var addr ethcommon.Address
	if privKey != nil {
		addr = common.EthereumPrivateKeyToAddress(privKey)
	}

	return &swapEthClient{
		ec:         ec,
		ethPrivKey: privKey,
		ethAddress: addr,
		gasPrice:   nil,
		gasLimit:   0,
		chainID:    chainID,
	}, nil
}

func (c *swapEthClient) Address() ethcommon.Address {
	return c.ethAddress
}

func (c *swapEthClient) SetAddress(addr ethcommon.Address) {
	if c.HasPrivateKey() {
		panic("SetAddress should not have been invoked when using an external signer")
	}
	c.ethAddress = addr
}

func (c *swapEthClient) PrivateKey() *ecdsa.PrivateKey {
	return c.ethPrivKey
}

func (c *swapEthClient) HasPrivateKey() bool {
	return c.ethPrivKey != nil
}

func (c *swapEthClient) Balance(ctx context.Context) (ethcommon.Address, *big.Int, error) {
	addr := c.Address()
	bal, err := c.ec.BalanceAt(ctx, addr, nil)
	if err != nil {
		return ethcommon.Address{}, nil, err
	}
	return addr, bal, nil
}

func (c *swapEthClient) ERC20BalanceAt(
	ctx context.Context,
	token ethcommon.Address,
	account ethcommon.Address,
	blockNumber *big.Int, // TODO: Do we need this unused parameter?
) (*big.Int, error) {
	tokenContract, err := contracts.NewIERC20(token, c.ec)
	if err != nil {
		return big.NewInt(0), err
	}
	return tokenContract.BalanceOf(c.CallOpts(ctx), account)
}

func (c *swapEthClient) ERC20Info(ctx context.Context, token ethcommon.Address) (
	name string,
	symbol string,
	decimals uint8,
	err error,
) {
	tokenContract, err := contracts.NewIERC20(token, c.ec)
	if err != nil {
		return "", "", 18, err
	}
	name, err = tokenContract.Name(c.CallOpts(ctx))
	if err != nil {
		return "", "", 18, err
	}
	symbol, err = tokenContract.Symbol(c.CallOpts(ctx))
	if err != nil {
		return "", "", 18, err
	}
	decimals, err = tokenContract.Decimals(c.CallOpts(ctx))
	if err != nil {
		return "", "", 18, err
	}
	return name, symbol, decimals, nil
}

// SetGasPrice sets the ethereum gas price for the instance to use (in wei).
func (c *swapEthClient) SetGasPrice(gasPrice uint64) {
	c.gasPrice = big.NewInt(0).SetUint64(gasPrice)
}

func (c *swapEthClient) CallOpts(ctx context.Context) *bind.CallOpts {
	return &bind.CallOpts{
		Pending:     false,
		From:        c.ethAddress, // might be all zeros if using external signer
		BlockNumber: nil,
		Context:     ctx,
	}
}

func (c *swapEthClient) TxOpts(ctx context.Context) (*bind.TransactOpts, error) {
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
	if txOpts.GasPrice == nil {
		txOpts.GasPrice, err = c.ec.SuggestGasPrice(ctx)
		if err != nil {
			return nil, err
		}
	}
	txOpts.GasLimit = c.gasLimit

	return txOpts, nil
}

func (c *swapEthClient) ChainID() *big.Int {
	return c.chainID
}

func (c *swapEthClient) TransactionByHash(ctx context.Context, hash ethcommon.Hash) (tx *ethtypes.Transaction, isPending bool, err error) { //nolint:lll
	return c.ec.TransactionByHash(ctx, hash)
}

func (c *swapEthClient) TransactionReceipt(ctx context.Context, txHash ethcommon.Hash) (*ethtypes.Receipt, error) {
	return c.ec.TransactionReceipt(ctx, txHash)
}

// WaitForReceipt waits for the receipt for the given transaction to be available and returns it.
func (c *swapEthClient) WaitForReceipt(ctx context.Context, txHash ethcommon.Hash) (*ethtypes.Receipt, error) {
	return block.WaitForReceipt(ctx, c.ec, txHash)
}

func (c *swapEthClient) WaitForTimestamp(ctx context.Context, ts time.Time) error {
	hdr, err := block.WaitForEthBlockAfterTimestamp(ctx, c.ec, ts.Unix())
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

func (c *swapEthClient) LatestBlockTimestamp(ctx context.Context) (time.Time, error) {
	hdr, err := c.ec.HeaderByNumber(ctx, nil)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(int64(hdr.Time), 0), nil
}

func (c *swapEthClient) Raw() *ethclient.Client {
	return c.ec
}
