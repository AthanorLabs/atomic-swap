package txsender

import (
	"context"
	"fmt"
	"math/big"

	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Sender signs and submits transactions to the chain
type Sender interface {
	SetContract(*contracts.SwapFactory)
	SetContractAddress(ethcommon.Address)
	Approve(spender ethcommon.Address,
		amount *big.Int) (ethcommon.Hash, *ethtypes.Receipt, error) // for ERC20 swaps
	NewSwap(_pubKeyClaim [32]byte, _pubKeyRefund [32]byte, _claimer ethcommon.Address,
		_timeoutDuration *big.Int, _nonce *big.Int, _ethAsset types.EthAsset,
		amount *big.Int) (ethcommon.Hash, *ethtypes.Receipt, error)
	SetReady(_swap contracts.SwapFactorySwap) (ethcommon.Hash, *ethtypes.Receipt, error)
	Claim(_swap contracts.SwapFactorySwap,
		_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error)
	Refund(_swap contracts.SwapFactorySwap,
		_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error)
}

type privateKeySender struct {
	ctx           context.Context
	ec            *ethclient.Client
	swapContract  *contracts.SwapFactory
	erc20Contract *contracts.IERC20
	txOpts        *TxOpts
}

// NewSenderWithPrivateKey returns a new *privateKeySender
func NewSenderWithPrivateKey(
	ctx context.Context,
	ec *ethclient.Client,
	swapContract *contracts.SwapFactory,
	erc20Contract *contracts.IERC20,
	txOpts *TxOpts,
) Sender {
	return &privateKeySender{
		ctx:           ctx,
		ec:            ec,
		swapContract:  swapContract,
		erc20Contract: erc20Contract,
		txOpts:        txOpts,
	}
}

func (s *privateKeySender) SetContract(contract *contracts.SwapFactory) {
	s.swapContract = contract
}

func (s *privateKeySender) SetContractAddress(_ ethcommon.Address) {}

func (s *privateKeySender) Approve(spender ethcommon.Address,
	amount *big.Int) (ethcommon.Hash, *ethtypes.Receipt, error) {
	s.txOpts.Lock()
	defer s.txOpts.Unlock()
	txOpts := s.txOpts.Inner()

	tx, err := s.erc20Contract.Approve(&txOpts, spender, amount)
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

func (s *privateKeySender) NewSwap(_pubKeyClaim [32]byte, _pubKeyRefund [32]byte,
	_claimer ethcommon.Address, _timeoutDuration *big.Int, _nonce *big.Int, _ethAsset types.EthAsset,
	value *big.Int) (ethcommon.Hash, *ethtypes.Receipt, error) {
	s.txOpts.Lock()
	defer s.txOpts.Unlock()
	txOpts := s.txOpts.Inner()

	// transfer ETH if we're not doing an ERC20 swap
	if _ethAsset == types.EthAssetETH {
		txOpts.Value = value
	}

	tx, err := s.swapContract.NewSwap(&txOpts, _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration,
		ethcommon.Address(_ethAsset), value, _nonce)
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

func (s *privateKeySender) SetReady(_swap contracts.SwapFactorySwap) (ethcommon.Hash, *ethtypes.Receipt, error) {
	s.txOpts.Lock()
	defer s.txOpts.Unlock()
	txOpts := s.txOpts.Inner()

	tx, err := s.swapContract.SetReady(&txOpts, _swap)
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

func (s *privateKeySender) Claim(_swap contracts.SwapFactorySwap,
	_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	s.txOpts.Lock()
	defer s.txOpts.Unlock()
	txOpts := s.txOpts.Inner()

	tx, err := s.swapContract.Claim(&txOpts, _swap, _s)
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

func (s *privateKeySender) Refund(_swap contracts.SwapFactorySwap,
	_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	s.txOpts.Lock()
	defer s.txOpts.Unlock()
	txOpts := s.txOpts.Inner()

	tx, err := s.swapContract.Refund(&txOpts, _swap, _s)
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
