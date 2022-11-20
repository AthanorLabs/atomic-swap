// Package txsender provides a common Sender interface for swapd instances using an
// ethereum key that is directly managed by swapd (`privateKeySender`) as well as an
// external sender (`ExternalSender`), where private key management and transaction
// signing is done via an external entity like Metamask.
package txsender

import (
	"context"
	"fmt"
	"math/big"

	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
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
	ethClient     extethclient.EthClient
	swapContract  *contracts.SwapFactory
	erc20Contract *contracts.IERC20
}

// NewSenderWithPrivateKey returns a new *privateKeySender
func NewSenderWithPrivateKey(
	ctx context.Context,
	ethClient extethclient.EthClient,
	swapContract *contracts.SwapFactory,
	erc20Contract *contracts.IERC20,
) Sender {
	return &privateKeySender{
		ctx:           ctx,
		ethClient:     ethClient,
		swapContract:  swapContract,
		erc20Contract: erc20Contract,
	}
}

func (s *privateKeySender) SetContract(contract *contracts.SwapFactory) {
	s.swapContract = contract
}

func (s *privateKeySender) SetContractAddress(_ ethcommon.Address) {}

func (s *privateKeySender) Approve(spender ethcommon.Address,
	amount *big.Int) (ethcommon.Hash, *ethtypes.Receipt, error) {
	s.ethClient.Lock()
	defer s.ethClient.Unlock()
	txOpts, err := s.ethClient.TxOpts(s.ctx)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	tx, err := s.erc20Contract.Approve(txOpts, spender, amount)
	if err != nil {
		err = fmt.Errorf("set_ready tx creation failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ethClient.Raw(), tx.Hash())
	if err != nil {
		err = fmt.Errorf("set_ready failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	return tx.Hash(), receipt, nil
}

func (s *privateKeySender) NewSwap(_pubKeyClaim [32]byte, _pubKeyRefund [32]byte,
	_claimer ethcommon.Address, _timeoutDuration *big.Int, _nonce *big.Int, _ethAsset types.EthAsset,
	value *big.Int) (ethcommon.Hash, *ethtypes.Receipt, error) {
	s.ethClient.Lock()
	defer s.ethClient.Unlock()
	txOpts, err := s.ethClient.TxOpts(s.ctx)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	// transfer ETH if we're not doing an ERC20 swap
	if _ethAsset == types.EthAssetETH {
		txOpts.Value = value
	}

	tx, err := s.swapContract.NewSwap(txOpts, _pubKeyClaim, _pubKeyRefund, _claimer, _timeoutDuration,
		ethcommon.Address(_ethAsset), value, _nonce)
	if err != nil {
		err = fmt.Errorf("new_swap tx creation failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ethClient.Raw(), tx.Hash())
	if err != nil {
		err = fmt.Errorf("new_swap failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	return tx.Hash(), receipt, nil
}

func (s *privateKeySender) SetReady(_swap contracts.SwapFactorySwap) (ethcommon.Hash, *ethtypes.Receipt, error) {
	s.ethClient.Lock()
	defer s.ethClient.Unlock()
	txOpts, err := s.ethClient.TxOpts(s.ctx)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	tx, err := s.swapContract.SetReady(txOpts, _swap)
	if err != nil {
		err = fmt.Errorf("set_ready tx creation failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ethClient.Raw(), tx.Hash())
	if err != nil {
		err = fmt.Errorf("set_ready failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	return tx.Hash(), receipt, nil
}

func (s *privateKeySender) Claim(_swap contracts.SwapFactorySwap,
	_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	s.ethClient.Lock()
	defer s.ethClient.Unlock()
	txOpts, err := s.ethClient.TxOpts(s.ctx)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	tx, err := s.swapContract.Claim(txOpts, _swap, _s)
	if err != nil {
		err = fmt.Errorf("claim tx creation failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ethClient.Raw(), tx.Hash())
	if err != nil {
		err = fmt.Errorf("claim failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	return tx.Hash(), receipt, nil
}

func (s *privateKeySender) Refund(_swap contracts.SwapFactorySwap,
	_s [32]byte) (ethcommon.Hash, *ethtypes.Receipt, error) {
	s.ethClient.Lock()
	defer s.ethClient.Unlock()
	txOpts, err := s.ethClient.TxOpts(s.ctx)
	if err != nil {
		return ethcommon.Hash{}, nil, err
	}

	tx, err := s.swapContract.Refund(txOpts, _swap, _s)
	if err != nil {
		err = fmt.Errorf("refund tx creation failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	receipt, err := block.WaitForReceipt(s.ctx, s.ethClient.Raw(), tx.Hash())
	if err != nil {
		err = fmt.Errorf("refund failed, %w", err)
		return ethcommon.Hash{}, nil, err
	}

	return tx.Hash(), receipt, nil
}
