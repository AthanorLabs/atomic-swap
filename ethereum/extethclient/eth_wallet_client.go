package extethclient

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log"

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

	Balance(ctx context.Context) (*big.Int, error)
	ERC20Balance(ctx context.Context, token ethcommon.Address) (*big.Int, error)

	ERC20Info(ctx context.Context, token ethcommon.Address) (name string, symbol string, decimals uint8, err error)

	SetGasPrice(uint64)
	SetGasLimit(uint64)
	CallOpts(ctx context.Context) *bind.CallOpts
	TxOpts(ctx context.Context) (*bind.TransactOpts, error)
	ChainID() *big.Int

	WaitForReceipt(ctx context.Context, txHash ethcommon.Hash) (*ethtypes.Receipt, error)
	WaitForTimestamp(ctx context.Context, ts time.Time) error
	LatestBlockTimestamp(ctx context.Context) (time.Time, error)

	Raw() *ethclient.Client
}

type ethClient struct {
	ec         *ethclient.Client
	ethPrivKey *ecdsa.PrivateKey
	ethAddress ethcommon.Address
	gasPrice   *big.Int
	gasLimit   uint64
	chainID    *big.Int
}

// NewEthClient creates and returns our extended ethereum client/wallet. The passed context
// is only used for creation.
func NewEthClient(ctx context.Context, ec *ethclient.Client, privKey *ecdsa.PrivateKey) (EthClient, error) {
	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	var addr ethcommon.Address
	if privKey != nil {
		addr = common.EthereumPrivateKeyToAddress(privKey)
	}

	return &ethClient{
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

func (c *ethClient) Balance(ctx context.Context) (*big.Int, error) {
	addr := c.Address()
	bal, err := c.ec.BalanceAt(ctx, addr, nil)
	if err != nil {
		return nil, err
	}
	return bal, nil
}

func (c *ethClient) ERC20Balance(ctx context.Context, token ethcommon.Address) (*big.Int, error) {
	tokenContract, err := contracts.NewIERC20(token, c.ec)
	if err != nil {
		return big.NewInt(0), err
	}
	return tokenContract.BalanceOf(c.CallOpts(ctx), c.Address())
}

func (c *ethClient) ERC20Info(ctx context.Context, token ethcommon.Address) (
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

// SetGasPrice sets the ethereum gas price (in wei) for use in transactions. In most
// cases, you should not use this function and let the ethereum client determine the
// suggested gas price at the current time. Setting a value of zero reverts to using
// the raw ethereum client's suggested price.
func (c *ethClient) SetGasPrice(gasPrice uint64) {
	if gasPrice == 0 {
		c.gasPrice = nil
		return
	}
	c.gasPrice = big.NewInt(0).SetUint64(gasPrice)
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

func (c *ethClient) LatestBlockTimestamp(ctx context.Context) (time.Time, error) {
	hdr, err := c.ec.HeaderByNumber(ctx, nil)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(int64(hdr.Time), 0), nil
}

func (c *ethClient) Raw() *ethclient.Client {
	return c.ec
}
