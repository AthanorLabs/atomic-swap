package txsender

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/swapfactory"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	errTransactionTimeout = errors.New("timed out waiting for transaction to be signed")

	transactionTimeout = time.Minute * 2 // arbitrary, TODO vary this based on env
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

	// outgoing encoded txs to be signed
	out chan *Transaction

	// incoming tx hashes
	in chan ethcommon.Hash
}

// NewExternalSender returns a new ExternalSender
func NewExternalSender(ctx context.Context, ec *ethclient.Client,
	contractAddr ethcommon.Address) (*ExternalSender, error) {
	abi, err := swapfactory.SwapFactoryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	return &ExternalSender{
		ctx:          ctx,
		ec:           ec,
		abi:          abi,
		contractAddr: contractAddr,
		out:          make(chan *Transaction),
		in:           make(chan ethcommon.Hash),
	}, nil
}

// OngoingCh returns the channel of outgoing transactions to be signed and submitted
func (s *ExternalSender) OngoingCh() <-chan *Transaction {
	return s.out
}

// IncomingCh returns the channel of incoming transaction hashes that have been signed and submitted
func (s *ExternalSender) IncomingCh() chan<- ethcommon.Hash {
	return s.in
}

// NewSwap prompts the external sender to sign a new_swap transaction
func (s *ExternalSender) NewSwap(_pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _claimer ethcommon.Address,
	_timeoutDuration *big.Int, _nonce *big.Int, value *big.Int) (ethcommon.Hash, *ethtypes.Receipt, error) {
	input, err := s.abi.Pack("new_swap", _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration, _nonce)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	tx := &Transaction{
		To:    s.contractAddr,
		Data:  fmt.Sprintf("0x%x", input),
		Value: fmt.Sprintf("%v", common.EtherAmount(*value).AsEther()),
	}

	s.out <- tx
	var txHash ethcommon.Hash
	select {
	case <-time.After(transactionTimeout):
		return ethcommon.Hash{}, nil, errTransactionTimeout
	case txHash = <-s.in:
	}

	receipt, err := waitForReceipt(s.ctx, s.ec, txHash)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return txHash, receipt, nil
}

// SetReady prompts the external sender to sign a set_ready transaction
func (s *ExternalSender) SetReady(_swap swapfactory.SwapFactorySwap) (ethcommon.Hash, *ethtypes.Receipt, error) {
	input, err := s.abi.Pack("set_ready", _swap)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return s.sendAndReceive(input)
}

// Claim prompts the external sender to sign a claim transaction
func (s *ExternalSender) Claim(_swap swapfactory.SwapFactorySwap,
	_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	input, err := s.abi.Pack("claim", _swap, _s)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return s.sendAndReceive(input)
}

// Refund prompts the external sender to sign a refund transaction
func (s *ExternalSender) Refund(_swap swapfactory.SwapFactorySwap,
	_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	input, err := s.abi.Pack("refund", _swap, _s)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return s.sendAndReceive(input)
}

func (s *ExternalSender) sendAndReceive(input []byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	tx := &Transaction{
		To:   s.contractAddr,
		Data: fmt.Sprintf("0x%x", input),
	}

	s.out <- tx
	var txHash ethcommon.Hash
	select {
	case <-time.After(transactionTimeout):
		return ethcommon.Hash{}, nil, errTransactionTimeout
	case txHash = <-s.in:
	}

	receipt, err := waitForReceipt(s.ctx, s.ec, txHash)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return txHash, receipt, nil
}
