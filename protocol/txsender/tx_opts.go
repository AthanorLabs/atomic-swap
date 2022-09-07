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

// Inner returns a copy of bind.TransactOpts. The original is never used in a transaction,
// so the copies given out do not have the Nonce, and Gas* fields initialised allowing them
// to be initialised dynamically. Reference:
// https://github.com/ethereum/go-ethereum/blob/v1.10.23/accounts/abi/bind/base.go#L49-L63
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
