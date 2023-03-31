// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

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
	errNoEthereumPrivateKey = errors.New("must provide --ethereum-privkey file for non-development environment")
)

type contractAddresses struct {
	SwapFactory ethcommon.Address `json:"swapFactory" validate:"required"`
	Forwarder   ethcommon.Address `json:"forwarder" validate:"required"`
}

func getOrDeploySwapFactory(
	ctx context.Context,
	address ethcommon.Address,
	env common.Environment,
	dataDir string,
	ec extethclient.EthClient,
	forwarderAddress ethcommon.Address,
) (ethcommon.Address, error) {
	var err error

	if env != common.Mainnet && (address == ethcommon.Address{}) {
		// we're on a development or testnet environment and we have no deployed contract,
		// so let's deploy one
		address, _, err = deploySwapFactory(ctx, ec.Raw(), ec.PrivateKey(), forwarderAddress, dataDir)
		if err != nil {
			return ethcommon.Address{}, fmt.Errorf("failed to deploy swap factory: %w", err)
		}
	} else {
		// otherwise, load the contract from the given address
		// and check that its bytecode is valid (ie. matches the
		// bytecode of this repo's swap contract)
		_, err = contracts.CheckSwapFactoryContractCode(ctx, ec.Raw(), address)
		if err != nil {
			return ethcommon.Address{}, err
		}
	}

	return address, nil
}

func deploySwapFactory(
	ctx context.Context,
	ec *ethclient.Client,
	privkey *ecdsa.PrivateKey,
	forwarderAddress ethcommon.Address,
	dataDir string,
) (ethcommon.Address, *contracts.SwapFactory, error) {

	if privkey == nil {
		return ethcommon.Address{}, nil, errNoEthereumPrivateKey
	}

	if (forwarderAddress == ethcommon.Address{}) {
		// deploy forwarder contract as well
		var err error
		forwarderAddress, err = contracts.DeployGSNForwarderWithKey(ctx, ec, privkey)
		if err != nil {
			return ethcommon.Address{}, nil, err
		}
	} else {
		if err := contracts.CheckForwarderContractCode(ctx, ec, forwarderAddress); err != nil {
			return ethcommon.Address{}, nil, err
		}
	}

	swapFactoryAddress, sf, err := contracts.DeploySwapFactoryWithKey(ctx, ec, privkey, forwarderAddress)
	if err != nil {
		return ethcommon.Address{}, nil, err
	}

	// store the contract address on disk
	err = writeContractAddressToFile(
		path.Join(dataDir, contractAddressesFile),
		&contractAddresses{
			SwapFactory: swapFactoryAddress,
			Forwarder:   forwarderAddress,
		},
	)
	if err != nil {
		return ethcommon.Address{}, nil, fmt.Errorf("failed to write contract address to file: %w", err)
	}

	return swapFactoryAddress, sf, nil
}

// writeContractAddressToFile writes the contract address to the given file
func writeContractAddressToFile(filePath string, addresses *contractAddresses) error {
	jsonData, err := vjson.MarshalIndentStruct(addresses, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Clean(filePath), jsonData, 0600)
}
