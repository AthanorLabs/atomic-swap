package txsender

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/swapfactory"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	maxRetries           = 360
	receiptSleepDuration = time.Second * 10
)

var (
	errReceiptTimeOut = errors.New("failed to get receipt, timed out")
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
	s.txOpts.Value = value
	defer func() {
		s.txOpts.Value = nil
	}()

	tx, err := s.contract.NewSwap(s.txOpts, _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration, _nonce)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	receipt, err := waitForReceipt(s.ctx, s.ec, tx.Hash())
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return tx.Hash(), receipt, nil
}

func (s *privateKeySender) SetReady(_ types.Hash,
	_swap swapfactory.SwapFactorySwap) (ethcommon.Hash, *ethtypes.Receipt, error) {
	tx, err := s.contract.SetReady(s.txOpts, _swap)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	receipt, err := waitForReceipt(s.ctx, s.ec, tx.Hash())
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return tx.Hash(), receipt, nil
}

func (s *privateKeySender) Claim(_ types.Hash, _swap swapfactory.SwapFactorySwap,
	_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	tx, err := s.contract.Claim(s.txOpts, _swap, _s)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	receipt, err := waitForReceipt(s.ctx, s.ec, tx.Hash())
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return tx.Hash(), receipt, nil
}

func (s *privateKeySender) Refund(_ types.Hash, _swap swapfactory.SwapFactorySwap,
	_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	tx, err := s.contract.Refund(s.txOpts, _swap, _s)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	receipt, err := waitForReceipt(s.ctx, s.ec, tx.Hash())
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return tx.Hash(), receipt, nil
}

func waitForReceipt(ctx context.Context, ec *ethclient.Client, txHash ethcommon.Hash) (*ethtypes.Receipt, error) {
	for i := 0; i < maxRetries; i++ {
		receipt, err := ec.TransactionReceipt(ctx, txHash)
		if err != nil {
			time.Sleep(receiptSleepDuration)
			continue
		}

		return receipt, nil
	}

	return nil, errReceiptTimeOut
}
