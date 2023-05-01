// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package contracts

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log"

	"github.com/athanorlabs/atomic-swap/ethereum/block"
)

var log = logging.Logger("contracts")

// DeploySwapCreatorWithKey deploys the SwapCreator contract using the passed privKey to
// pay for the deployment.
func DeploySwapCreatorWithKey(
	ctx context.Context,
	ec *ethclient.Client,
	privKey *ecdsa.PrivateKey,
) (ethcommon.Address, *SwapCreator, error) {
	txOpts, err := newTXOpts(ctx, ec, privKey)
	if err != nil {
		return ethcommon.Address{}, nil, err
	}

	log.Infof("deploying SwapCreator.sol")
	address, tx, sf, err := DeploySwapCreator(txOpts, ec)
	if err != nil {
		return ethcommon.Address{}, nil, fmt.Errorf("failed to deploy swap creator: %w", err)
	}

	_, err = block.WaitForReceipt(ctx, ec, tx.Hash())
	if err != nil {
		return ethcommon.Address{}, nil, err
	}

	log.Infof("deployed SwapCreator.sol: address=%s tx hash=%s", address, tx.Hash())
	return address, sf, nil
}

func newTXOpts(ctx context.Context, ec *ethclient.Client, privkey *ecdsa.PrivateKey) (*bind.TransactOpts, error) {
	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	txOpts, err := bind.NewKeyedTransactorWithChainID(privkey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to make transactor: %w", err)
	}
	return txOpts, nil
}
