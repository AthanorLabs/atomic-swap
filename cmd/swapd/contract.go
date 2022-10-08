package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"path"

	"github.com/athanorlabs/atomic-swap/common"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	errNoEthereumPrivateKey = errors.New("must provide --ethereum-privkey file for non-development environment")
	errInvalidSwapContract  = errors.New("given contract address does not contain correct code")
)

func getOrDeploySwapFactory(
	ctx context.Context,
	address ethcommon.Address,
	env common.Environment,
	dataDir string,
	privkey *ecdsa.PrivateKey,
	ec *ethclient.Client,
) (*contracts.SwapFactory, ethcommon.Address, error) {
	var (
		sf *contracts.SwapFactory
	)

	if env != common.Mainnet && (address == ethcommon.Address{}) {
		if privkey == nil {
			return nil, ethcommon.Address{}, errNoEthereumPrivateKey
		}

		chainID, err := ec.ChainID(ctx)
		if err != nil {
			return nil, ethcommon.Address{}, err
		}
		txOpts, err := bind.NewKeyedTransactorWithChainID(privkey, chainID)
		if err != nil {
			return nil, ethcommon.Address{}, fmt.Errorf("failed to make transactor: %w", err)
		}

		// deploy contracts.sol
		var tx *ethtypes.Transaction
		address, tx, sf, err = deploySwapFactory(ec, txOpts)
		if err != nil {
			return nil, ethcommon.Address{}, fmt.Errorf("failed to deploy swap factory: %w", err)
		}

		log.Infof("deployed SwapFactory.sol: address=%s tx hash=%s", address, tx.Hash())

		// store the contract address on disk
		fp := path.Join(dataDir, "contract-address.json")
		if err = pcommon.WriteContractAddressToFile(fp, address.String()); err != nil {
			return nil, ethcommon.Address{}, fmt.Errorf("failed to write contract address to file: %w", err)
		}
	} else {
		var err error
		sf, err = getSwapFactory(ec, address)
		if err != nil {
			return nil, ethcommon.Address{}, err
		}
		log.Infof("loaded contracts.sol from address %s", address)

		err = checkContractCode(ctx, ec, address)
		if err != nil {
			return nil, ethcommon.Address{}, err
		}
	}

	return sf, address, nil
}

func checkContractCode(ctx context.Context, ec *ethclient.Client, contractAddr ethcommon.Address) error {
	code, err := ec.CodeAt(ctx, contractAddr, nil)
	if err != nil {
		return err
	}

	expectedCode := ethcommon.FromHex(contracts.SwapFactoryMetaData.Bin)
	if !bytes.Contains(expectedCode, code) {
		return errInvalidSwapContract
	}

	return nil
}

func getSwapFactory(client *ethclient.Client, addr ethcommon.Address) (*contracts.SwapFactory, error) {
	return contracts.NewSwapFactory(addr, client)
}

func deploySwapFactory(client *ethclient.Client, txOpts *bind.TransactOpts) (ethcommon.Address, *ethtypes.Transaction, *contracts.SwapFactory, error) { //nolint:lll
	return contracts.DeploySwapFactory(txOpts, client)
}
