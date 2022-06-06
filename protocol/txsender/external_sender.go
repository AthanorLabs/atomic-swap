package txsender

import (
	"context"
	"math/big"

	"github.com/noot/atomic-swap/swapfactory"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ExternalSender struct {
	ctx context.Context
	ec  *ethclient.Client
	abi *abi.ABI

	// outgoing encoded txs to be signed
	out chan []byte

	// incoming tx hashes
	in chan ethcommon.Hash
}

func NewExternalSender(ctx context.Context, ec *ethclient.Client) (*ExternalSender, error) {
	abi, err := swapfactory.SwapFactoryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	return &ExternalSender{
		ctx: ctx,
		ec:  ec,
		abi: abi,
		out: make(chan []byte),
		in:  make(chan ethcommon.Hash),
	}, nil
}

func (s *ExternalSender) OngoingCh() <-chan []byte {
	return s.out
}

func (s *ExternalSender) IncomingCh() chan<- ethcommon.Hash {
	return s.in
}

func (s *ExternalSender) NewSwap(_pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _claimer ethcommon.Address, _timeoutDuration *big.Int, _nonce *big.Int) (ethcommon.Hash, *ethtypes.Receipt, error) {
	input, err := s.abi.Pack("new_swap", _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration, _nonce)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	// TODO: how to make this synchronous?
	s.out <- input
	txHash := <-s.in

	receipt, err := waitForReceipt(s.ctx, s.ec, txHash)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return txHash, receipt, nil
}

func (s *ExternalSender) SetReady(_swap swapfactory.SwapFactorySwap) (ethcommon.Hash, *ethtypes.Receipt, error) {
	input, err := s.abi.Pack("set_ready", _swap)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	s.out <- input
	txHash := <-s.in

	receipt, err := waitForReceipt(s.ctx, s.ec, txHash)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return txHash, receipt, nil
}

func (s *ExternalSender) Claim(_swap swapfactory.SwapFactorySwap, _s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	input, err := s.abi.Pack("claim", _swap, _s)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	s.out <- input
	txHash := <-s.in

	receipt, err := waitForReceipt(s.ctx, s.ec, txHash)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return txHash, receipt, nil
}

func (s *ExternalSender) Refund(_swap swapfactory.SwapFactorySwap, _s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	input, err := s.abi.Pack("refund", _swap, _s)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	s.out <- input
	txHash := <-s.in

	receipt, err := waitForReceipt(s.ctx, s.ec, txHash)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	return txHash, receipt, nil
}
