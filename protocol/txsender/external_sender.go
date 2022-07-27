package txsender

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/ethereum/block"
	"github.com/noot/atomic-swap/swapfactory"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	errTransactionTimeout = errors.New("timed out waiting for transaction to be signed")
	errNoSwapWithID       = errors.New("no swap with given id")

	transactionTimeout = time.Minute * 2 // amount of time user has to sign message
)

// Transaction represents a transaction to be signed by the front-end
type Transaction struct {
	To    ethcommon.Address
	Data  string
	Value string
}

type swapChs struct {
	// outgoing encoded txs to be signed
	out chan *Transaction
	// incoming tx hashes
	in chan ethcommon.Hash
}

// ExternalSender represents a transaction signer and sender that is external to the daemon (ie. a front-end)
type ExternalSender struct {
	ctx          context.Context
	ec           *ethclient.Client
	abi          *abi.ABI
	contractAddr ethcommon.Address

	sync.RWMutex

	swaps map[types.Hash]*swapChs
}

// NewExternalSender returns a new ExternalSender
func NewExternalSender(ctx context.Context, env common.Environment, ec *ethclient.Client,
	contractAddr ethcommon.Address) (*ExternalSender, error) {
	abi, err := swapfactory.SwapFactoryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	switch env {
	case common.Mainnet, common.Stagenet:
		transactionTimeout = time.Hour
	}

	return &ExternalSender{
		ctx:          ctx,
		ec:           ec,
		abi:          abi,
		contractAddr: contractAddr,
		swaps:        make(map[types.Hash]*swapChs),
	}, nil
}

// SetContract ...
func (s *ExternalSender) SetContract(_ *swapfactory.SwapFactory) {}

// SetContractAddress ...
func (s *ExternalSender) SetContractAddress(addr ethcommon.Address) {
	s.contractAddr = addr
}

// OngoingCh returns the channel of outgoing transactions to be signed and submitted
func (s *ExternalSender) OngoingCh(id types.Hash) (<-chan *Transaction, error) {
	s.RLock()
	defer s.RUnlock()
	chs, has := s.swaps[id]
	if !has {
		return nil, errNoSwapWithID
	}

	return chs.out, nil
}

// IncomingCh returns the channel of incoming transaction hashes that have been signed and submitted
func (s *ExternalSender) IncomingCh(id types.Hash) (chan<- ethcommon.Hash, error) {
	s.RLock()
	defer s.RUnlock()
	chs, has := s.swaps[id]
	if !has {
		return nil, errNoSwapWithID
	}
	return chs.in, nil
}

// AddID initialises the sender with a swap w/ the given ID
func (s *ExternalSender) AddID(id types.Hash) {
	s.Lock()
	defer s.Unlock()
	_, has := s.swaps[id]
	if has {
		return
	}

	s.swaps[id] = &swapChs{
		out: make(chan *Transaction),
		in:  make(chan ethcommon.Hash),
	}
}

// DeleteID deletes the swap w/ the given ID from the sender
func (s *ExternalSender) DeleteID(id types.Hash) {
	s.Lock()
	defer s.Unlock()
	delete(s.swaps, id)
}

// NewSwap prompts the external sender to sign a new_swap transaction
func (s *ExternalSender) NewSwap(id types.Hash, _pubKeyClaim [32]byte, _pubKeyRefund [32]byte,
	_claimer ethcommon.Address, _timeoutDuration *big.Int, _nonce *big.Int,
	value *big.Int) (ethcommon.Hash, *ethtypes.Receipt, error) {
	input, err := s.abi.Pack("new_swap", _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration, _nonce)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	tx := &Transaction{
		To:    s.contractAddr,
		Data:  fmt.Sprintf("0x%x", input),
		Value: fmt.Sprintf("%v", common.EtherAmount(*value).AsEther()),
	}

	s.RLock()
	defer s.RUnlock()
	chs, has := s.swaps[id]
	if !has {
		return ethcommon.Hash{}, nil, errNoSwapWithID
	}

	chs.out <- tx
	var txHash ethcommon.Hash
	select {
	case <-time.After(transactionTimeout):
		return ethcommon.Hash{}, nil, errTransactionTimeout
	case txHash = <-chs.in:
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ec, txHash)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return txHash, receipt, nil
}

// SetReady prompts the external sender to sign a set_ready transaction
func (s *ExternalSender) SetReady(id types.Hash,
	_swap swapfactory.SwapFactorySwap) (ethcommon.Hash, *ethtypes.Receipt, error) {
	input, err := s.abi.Pack("set_ready", _swap)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return s.sendAndReceive(id, input)
}

// Claim prompts the external sender to sign a claim transaction
func (s *ExternalSender) Claim(id types.Hash, _swap swapfactory.SwapFactorySwap,
	_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	input, err := s.abi.Pack("claim", _swap, _s)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return s.sendAndReceive(id, input)
}

// Refund prompts the external sender to sign a refund transaction
func (s *ExternalSender) Refund(id types.Hash, _swap swapfactory.SwapFactorySwap,
	_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	input, err := s.abi.Pack("refund", _swap, _s)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return s.sendAndReceive(id, input)
}

func (s *ExternalSender) sendAndReceive(id types.Hash,
	input []byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	tx := &Transaction{
		To:   s.contractAddr,
		Data: fmt.Sprintf("0x%x", input),
	}

	s.RLock()
	defer s.RUnlock()
	chs, has := s.swaps[id]
	if !has {
		return ethcommon.Hash{}, nil, errNoSwapWithID
	}

	chs.out <- tx
	var txHash ethcommon.Hash
	select {
	case <-time.After(transactionTimeout):
		return ethcommon.Hash{}, nil, errTransactionTimeout
	case txHash = <-chs.in:
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ec, txHash)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return txHash, receipt, nil
}
