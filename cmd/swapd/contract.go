package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/athanorlabs/atomic-swap/common"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	errNoEthereumPrivateKey = errors.New("must provide --ethereum-privkey file for non-development environment")
)

func getOrDeploySwapFactory(
	ctx context.Context,
	address ethcommon.Address,
	env common.Environment,
	dataDir string,
	privkey *ecdsa.PrivateKey,
	ec *ethclient.Client,
	forwarderAddress ethcommon.Address,
) (*contracts.SwapFactory, ethcommon.Address, error) {
	var (
		sf  *contracts.SwapFactory
		err error
	)

	if env != common.Mainnet && (address == ethcommon.Address{}) {
		// we're on a development or testnet environment and we have no deployed contract,
		// so let's deploy one
		address, sf, err = deploySwapFactory(ctx, ec, privkey, forwarderAddress, dataDir)
		if err != nil {
			return nil, ethcommon.Address{}, fmt.Errorf("failed to deploy swap factory: %w", err)
		}
	} else {
		// otherwise, load the contract from the given address
		// and check that its bytecode is valid (ie. matches the
		// bytecode of this repo's swap contract)
		sf, err = getSwapFactory(ec, address)
		if err != nil {
			return nil, ethcommon.Address{}, err
		}
		log.Infof("loaded SwapFactory.sol from address %s", address)

		_, err = contracts.CheckSwapFactoryContractCode(ctx, ec, address)
		if err != nil {
			return nil, ethcommon.Address{}, err
		}
	}

	return sf, address, nil
}

func getSwapFactory(client *ethclient.Client, addr ethcommon.Address) (*contracts.SwapFactory, error) {
	return contracts.NewSwapFactory(addr, client)
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
		address, err := contracts.DeployGSNForwarderWithKey(ctx, ec, privkey)
		if err != nil {
			return ethcommon.Address{}, nil, err
		}
		forwarderAddress = address
	}

	address, sf, err := contracts.DeploySwapFactoryWithKey(ctx, ec, privkey, forwarderAddress)
	if err != nil {
		return ethcommon.Address{}, nil, err
	}

	// store the contract address on disk
	fp := path.Join(dataDir, "contract-address.json")
	if err = writeContractAddressToFile(fp, address.String()); err != nil {
		return ethcommon.Address{}, nil, fmt.Errorf("failed to write contract address to file: %w", err)
	}

	return address, sf, nil
}

// writeContractAddressToFile writes the contract address to the given file
func writeContractAddressToFile(fp, addr string) error {
	bz, err := json.MarshalIndent(&struct {
		ContractAddress string
	}{
		ContractAddress: addr,
	}, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Clean(fp), bz, 0600)
}
