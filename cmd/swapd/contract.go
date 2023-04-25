// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/vjson"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	contractAddressesFile = "contract-addresses.json"
)

var (
	errNoEthPrivateKey = fmt.Errorf("must provide --%s file for non-development environment", flagEthPrivKey)
)

type contractAddresses struct {
	SwapCreatorAddr ethcommon.Address `json:"swapCreatorAddr" validate:"required"`
	ForwarderAddr   ethcommon.Address `json:"forwarderAddr" validate:"required"`
}

func getOrDeploySwapCreator(
	ctx context.Context,
	swapCreatorAddr ethcommon.Address,
	env common.Environment,
	dataDir string,
	ec extethclient.EthClient,
	forwarderAddr ethcommon.Address,
) (ethcommon.Address, error) {
	var err error
	if (swapCreatorAddr == ethcommon.Address{}) {
		if env == common.Mainnet {
			log.Warnf("you are deploying SwapCreator.sol on mainnet! giving you a few seconds to cancel if this is unintended")
			time.Sleep(10 * time.Second)
		}

		swapCreatorAddr, _, err = deploySwapCreator(ctx, ec.Raw(), ec.PrivateKey(), forwarderAddr, dataDir)
		if err != nil {
			return ethcommon.Address{}, fmt.Errorf("failed to deploy swap creator: %w", err)
		}
	} else {
		// otherwise, load the contract from the given address
		// and check that its bytecode is valid (ie. matches the
		// bytecode of this repo's swap contract)
		_, err = contracts.CheckSwapCreatorContractCode(ctx, ec.Raw(), swapCreatorAddr)
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
	forwarderAddr ethcommon.Address,
	dataDir string,
) (ethcommon.Address, *contracts.SwapCreator, error) {

	if privkey == nil {
		return ethcommon.Address{}, nil, errNoEthPrivateKey
	}

	if (forwarderAddr == ethcommon.Address{}) {
		// deploy forwarder contract as well
		var err error
		forwarderAddr, err = contracts.DeployGSNForwarderWithKey(ctx, ec, privkey)
		if err != nil {
			return ethcommon.Address{}, nil, err
		}
	} else {
		// TODO: ignore this is the forwarderAddr is the one that's hardcoded for this network
		if err := contracts.CheckForwarderContractCode(ctx, ec, forwarderAddr); err != nil {
			return ethcommon.Address{}, nil, err
		}
	}

	swapCreatorAddr, sf, err := contracts.DeploySwapCreatorWithKey(ctx, ec, privkey, forwarderAddr)
	if err != nil {
		return ethcommon.Address{}, nil, err
	}

	// store the contract addresses on disk
	err = writeContractAddressesToFile(
		path.Join(dataDir, contractAddressesFile),
		&contractAddresses{
			SwapCreatorAddr: swapCreatorAddr,
			ForwarderAddr:   forwarderAddr,
		},
	)
	if err != nil {
		return ethcommon.Address{}, nil, fmt.Errorf("failed to write contract address to file: %w", err)
	}

	return swapCreatorAddr, sf, nil
}

// writeContractAddressesToFile writes the contract addresses to the given file
func writeContractAddressesToFile(filePath string, addresses *contractAddresses) error {
	jsonData, err := vjson.MarshalIndentStruct(addresses, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Clean(filePath), jsonData, 0600)
}
