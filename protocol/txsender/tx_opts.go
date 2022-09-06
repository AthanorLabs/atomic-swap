package txsender

import (
	"crypto/ecdsa"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// TxOpts wraps go-ethereum's *bind.TransactOpts.
type TxOpts struct {
	inner *bind.TransactOpts
	mu    sync.Mutex // locks from TX start until receipt so we don't reuse ETH nonce values
}

// NewTxOpts returns a new *TxOpts from the given private key and chain ID
func NewTxOpts(privkey *ecdsa.PrivateKey, chainID *big.Int) (*TxOpts, error) {
	txOpts, err := bind.NewKeyedTransactorWithChainID(privkey, chainID)
	if err != nil {
		return nil, err
	}

	return &TxOpts{
		inner: txOpts,
	}, nil
}

// Inner returns the bind.TransactOpts contained
func (txOpts *TxOpts) Inner() bind.TransactOpts {
	return *txOpts.inner
}

// Lock ...
func (txOpts *TxOpts) Lock() {
	txOpts.mu.Lock()
}

// Unlock ...
func (txOpts *TxOpts) Unlock() {
	txOpts.mu.Unlock()
}
