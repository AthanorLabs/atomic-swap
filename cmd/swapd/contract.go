// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/athanorlabs/atomic-swap/common"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	errNoEthPrivateKey = fmt.Errorf("must provide --%s file for non-development environment", flagEthPrivKey)
)

func getOrDeploySwapCreator(
	ctx context.Context,
	swapCreatorAddr ethcommon.Address,
	env common.Environment,
	ec extethclient.EthClient,
) (ethcommon.Address, error) {
	var err error
	if (swapCreatorAddr == ethcommon.Address{}) {
		if env == common.Mainnet {
			log.Warnf("you are deploying SwapCreator.sol on mainnet! giving you a few seconds to cancel if this is unintended")
			time.Sleep(10 * time.Second)
		}

		swapCreatorAddr, err = deploySwapCreator(ctx, ec.Raw(), ec.PrivateKey())
		if err != nil {
			return ethcommon.Address{}, fmt.Errorf("failed to deploy swap creator: %w", err)
		}
	} else {
		// otherwise, load the contract from the given address
		// and check that its bytecode is valid (ie. matches the
		// bytecode of this repo's swap contract)
		err = contracts.CheckSwapCreatorContractCode(ctx, ec.Raw(), swapCreatorAddr)
		if err != nil {
			return ethcommon.Address{}, err
		}
	}

	return swapCreatorAddr, nil
}

func deploySwapCreator(
	ctx context.Context,
	ec *ethclient.Client,
	privkey *ecdsa.PrivateKey,
) (ethcommon.Address, error) {
	if privkey == nil {
		return ethcommon.Address{}, errNoEthPrivateKey
	}

	swapCreatorAddr, _, err := contracts.DeploySwapCreatorWithKey(ctx, ec, privkey)
	if err != nil {
		return ethcommon.Address{}, err
	}

	return swapCreatorAddr, nil
}
