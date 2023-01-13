package txsender

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	errTransactionTimeout = errors.New("timed out waiting for transaction to be signed")
	transactionTimeout    = time.Minute * 2 // amount of time user has to sign message
)

// Transaction represents a transaction to be signed by the front-end
type Transaction struct {
	To    ethcommon.Address
	Data  string
	Value string
}

// ExternalSender represents a transaction signer and sender that is external to the daemon (ie. a front-end)
type ExternalSender struct {
	ctx          context.Context
	ec           *ethclient.Client
	abi          *abi.ABI
	contractAddr ethcommon.Address
	erc20Addr    ethcommon.Address

	sync.Mutex

	// outgoing encoded txs to be signed
	out chan *Transaction
	// incoming tx hashes
	in chan ethcommon.Hash
}

// NewExternalSender returns a new ExternalSender
func NewExternalSender(ctx context.Context, env common.Environment, ec *ethclient.Client,
	contractAddr ethcommon.Address, erc20Addr ethcommon.Address) (*ExternalSender, error) {
	abi, err := contracts.SwapFactoryMetaData.GetAbi()
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
		erc20Addr:    erc20Addr,
		out:          make(chan *Transaction),
		in:           make(chan ethcommon.Hash),
	}, nil
}

// SetContract ...
func (s *ExternalSender) SetContract(_ *contracts.SwapFactory) {}

// SetContractAddress ...
func (s *ExternalSender) SetContractAddress(addr ethcommon.Address) {
	s.contractAddr = addr
}

// OngoingCh returns the channel of outgoing transactions to be signed and submitted
func (s *ExternalSender) OngoingCh(id types.Hash) <-chan *Transaction {
	return s.out
}

// IncomingCh returns the channel of incoming transaction hashes that have been signed and submitted
func (s *ExternalSender) IncomingCh(id types.Hash) chan<- ethcommon.Hash {
	return s.in
}

// Approve prompts the external sender to sign an ERC20 Approve transaction
func (s *ExternalSender) Approve(spender ethcommon.Address,
	amount *big.Int) (ethcommon.Hash, *ethtypes.Receipt, error) {
	input, err := s.abi.Pack("approve", spender, amount)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return s.sendAndReceive(input, s.erc20Addr)
}

// NewSwap prompts the external sender to sign a new_swap transaction
func (s *ExternalSender) NewSwap(_pubKeyClaim [32]byte, _pubKeyRefund [32]byte,
	_claimer ethcommon.Address, _timeoutDuration *big.Int, _nonce *big.Int, _ethAsset types.EthAsset,
	value *big.Int) (ethcommon.Hash, *ethtypes.Receipt, error) {
	input, err := s.abi.Pack("new_swap", _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration,
		_ethAsset, value, _nonce)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	valueWei := coins.NewWeiAmount(value)
	tx := &Transaction{
		To:    s.contractAddr,
		Data:  fmt.Sprintf("0x%x", input),
		Value: fmt.Sprintf("%v", valueWei.AsEther()),
	}

	s.Lock()
	defer s.Unlock()

	s.out <- tx
	var txHash ethcommon.Hash
	select {
	case <-time.After(transactionTimeout):
		return ethcommon.Hash{}, nil, errTransactionTimeout
	case txHash = <-s.in:
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ec, txHash)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return txHash, receipt, nil
}

// SetReady prompts the external sender to sign a set_ready transaction
func (s *ExternalSender) SetReady(_swap contracts.SwapFactorySwap) (ethcommon.Hash, *ethtypes.Receipt, error) {
	input, err := s.abi.Pack("set_ready", _swap)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return s.sendAndReceive(input, s.contractAddr)
}

// Claim prompts the external sender to sign a claim transaction
func (s *ExternalSender) Claim(_swap contracts.SwapFactorySwap,
	_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	input, err := s.abi.Pack("claim", _swap, _s)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return s.sendAndReceive(input, s.contractAddr)
}

// Refund prompts the external sender to sign a refund transaction
func (s *ExternalSender) Refund(_swap contracts.SwapFactorySwap,
	_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	input, err := s.abi.Pack("refund", _swap, _s)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return s.sendAndReceive(input, s.contractAddr)
}

func (s *ExternalSender) sendAndReceive(input []byte, to ethcommon.Address) (ethcommon.Hash, *ethtypes.Receipt, error) {
	tx := &Transaction{
		To:   to,
		Data: fmt.Sprintf("0x%x", input),
	}

	s.Lock()
	defer s.Unlock()

	s.out <- tx
	var txHash ethcommon.Hash
	select {
	case <-time.After(transactionTimeout):
		return ethcommon.Hash{}, nil, errTransactionTimeout
	case txHash = <-s.in:
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ec, txHash)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return txHash, receipt, nil
}
