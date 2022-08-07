package txsender

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/ethereum/block"
	"github.com/noot/atomic-swap/swapfactory"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Sender signs and submits transactions to the chain
type Sender interface {
	SetContract(*swapfactory.SwapFactory)
	SetContractAddress(ethcommon.Address)
	NewSwap(id types.Hash, _pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _claimer ethcommon.Address,
		_timeoutDuration *big.Int, _nonce *big.Int, amount *big.Int) (ethcommon.Hash, *ethtypes.Receipt, error)
	SetReady(id types.Hash, _swap swapfactory.SwapFactorySwap) (ethcommon.Hash, *ethtypes.Receipt, error)
	Claim(id types.Hash, _swap swapfactory.SwapFactorySwap,
		_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error)
	Refund(id types.Hash, _swap swapfactory.SwapFactorySwap,
		_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error)
}

type privateKeySender struct {
	ctx      context.Context
	ec       *ethclient.Client
	contract *swapfactory.SwapFactory
	txOpts   *bind.TransactOpts
	txLock   sync.Mutex // locks from TX start until receipt so we don't reuse ETH nonce values
}

// NewSenderWithPrivateKey returns a new *privateKeySender
func NewSenderWithPrivateKey(ctx context.Context, ec *ethclient.Client, contract *swapfactory.SwapFactory,
	txOpts *bind.TransactOpts) Sender {
	return &privateKeySender{
		ctx:      ctx,
		ec:       ec,
		contract: contract,
		txOpts:   txOpts,
	}
}

func (s *privateKeySender) SetContract(contract *swapfactory.SwapFactory) {
	s.contract = contract
}

func (s *privateKeySender) SetContractAddress(_ ethcommon.Address) {}

func (s *privateKeySender) NewSwap(_ types.Hash, _pubKeyClaim [32]byte, _pubKeyRefund [32]byte,
	_claimer ethcommon.Address, _timeoutDuration *big.Int, _nonce *big.Int,
	value *big.Int) (ethcommon.Hash, *ethtypes.Receipt, error) {
	s.txLock.Lock()
	defer s.txLock.Unlock()
	txOpts := *s.txOpts // make a copy, so we don't modify the original
	txOpts.Value = value
	tx, err := s.contract.NewSwap(&txOpts, _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration, _nonce)
	if err != nil {
		err = fmt.Errorf("new_swap tx creation failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ec, tx.Hash())
	if err != nil {
		err = fmt.Errorf("new_swap failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	return tx.Hash(), receipt, nil
}

func (s *privateKeySender) SetReady(_ types.Hash,
	_swap swapfactory.SwapFactorySwap) (ethcommon.Hash, *ethtypes.Receipt, error) {
	s.txLock.Lock()
	defer s.txLock.Unlock()
	txOpts := *s.txOpts // make a copy, so we don't modify the original
	tx, err := s.contract.SetReady(&txOpts, _swap)
	if err != nil {
		err = fmt.Errorf("set_ready tx creation failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ec, tx.Hash())
	if err != nil {
		err = fmt.Errorf("set_ready failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	return tx.Hash(), receipt, nil
}

func (s *privateKeySender) Claim(_ types.Hash, _swap swapfactory.SwapFactorySwap,
	_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	s.txLock.Lock()
	defer s.txLock.Unlock()
	txOpts := *s.txOpts // make a copy, so we don't modify the original
	tx, err := s.contract.Claim(&txOpts, _swap, _s)
	if err != nil {
		err = fmt.Errorf("claim tx creation failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ec, tx.Hash())
	if err != nil {
		err = fmt.Errorf("claim failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	return tx.Hash(), receipt, nil
}

func (s *privateKeySender) Refund(_ types.Hash, _swap swapfactory.SwapFactorySwap,
	_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	s.txLock.Lock()
	defer s.txLock.Unlock()
	txOpts := *s.txOpts // make a copy, so we don't modify the original
	tx, err := s.contract.Refund(&txOpts, _swap, _s)
	if err != nil {
		err = fmt.Errorf("refund tx creation failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ec, tx.Hash())
	if err != nil {
		err = fmt.Errorf("refund failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	return tx.Hash(), receipt, nil
}
