package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"path"

	"github.com/AthanorLabs/go-relayer/impls/gsnforwarder"
	"github.com/athanorlabs/atomic-swap/common"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	errNoEthereumPrivateKey = errors.New("must provide --ethereum-privkey file for non-development environment")
	//errInvalidSwapContract  = errors.New("given contract address does not contain correct code")
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

		err = pcommon.CheckContractCode(ctx, ec, address)
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

	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return ethcommon.Address{}, nil, err
	}
	txOpts, err := bind.NewKeyedTransactorWithChainID(privkey, chainID)
	if err != nil {
		return ethcommon.Address{}, nil, fmt.Errorf("failed to make transactor: %w", err)
	}

	if (forwarderAddress == ethcommon.Address{}) {
		// deploy forwarder contract as well
		address, err := deployForwarder(ctx, ec, txOpts) //nolint:govet
		if err != nil {
			return ethcommon.Address{}, nil, err
		}

		forwarderAddress = address
	}

	// deploy contracts.sol
	address, tx, sf, err := contracts.DeploySwapFactory(txOpts, ec, forwarderAddress)
	if err != nil {
		return ethcommon.Address{}, nil, fmt.Errorf("failed to deploy swap factory: %w", err)
	}

	_, err = block.WaitForReceipt(ctx, ec, tx.Hash())
	if err != nil {
		return ethcommon.Address{}, nil, err
	}

	log.Infof("deployed SwapFactory.sol: address=%s tx hash=%s", address, tx.Hash())

	// store the contract address on disk
	fp := path.Join(dataDir, "contract-address.json")
	if err = pcommon.WriteContractAddressToFile(fp, address.String()); err != nil {
		return ethcommon.Address{}, nil, fmt.Errorf("failed to write contract address to file: %w", err)
	}

	return address, sf, nil
}

func deployForwarder(
	ctx context.Context,
	ec *ethclient.Client,
	txOpts *bind.TransactOpts,
) (ethcommon.Address, error) {
	address, tx, contract, err := gsnforwarder.DeployForwarder(txOpts, ec)
	if err != nil {
		return ethcommon.Address{}, fmt.Errorf("failed to deploy Forwarder.sol: %w", err)
	}

	_, err = block.WaitForReceipt(ctx, ec, tx.Hash())
	if err != nil {
		return ethcommon.Address{}, err
	}

	tx, err = contract.RegisterDomainSeparator(txOpts, gsnforwarder.DefaultName, gsnforwarder.DefaultVersion)
	if err != nil {
		return ethcommon.Address{}, fmt.Errorf("failed to register domain separator: %w", err)
	}

	_, err = block.WaitForReceipt(ctx, ec, tx.Hash())
	if err != nil {
		return ethcommon.Address{}, err
	}

	return address, nil
}
