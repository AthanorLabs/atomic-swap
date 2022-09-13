package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"path"

	"github.com/athanorlabs/atomic-swap/common"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
	"github.com/athanorlabs/atomic-swap/swapfactory"

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
	chainID *big.Int,
	privkey *ecdsa.PrivateKey,
	ec *ethclient.Client,
) (*swapfactory.SwapFactory, ethcommon.Address, error) {
	var (
		sf *swapfactory.SwapFactory
	)

	if env != common.Mainnet && (address == ethcommon.Address{}) {
		if privkey == nil {
			return nil, ethcommon.Address{}, errNoEthereumPrivateKey
		}

		txOpts, err := bind.NewKeyedTransactorWithChainID(privkey, chainID)
		if err != nil {
			return nil, ethcommon.Address{}, fmt.Errorf("failed to make transactor: %w", err)
		}

		// deploy SwapFactory.sol
		var tx *ethtypes.Transaction
		address, tx, sf, err = deploySwapFactory(ec, txOpts)
		if err != nil {
			return nil, ethcommon.Address{}, fmt.Errorf("failed to deploy swap factory: %w; please check your chain ID", err)
		}

		log.Infof("deployed SwapFactory.sol: address=%s tx hash=%s", address, tx.Hash())

		// store the contract address on disk
		fp := path.Join(dataDir, "contractaddress")
		if err = pcommon.WriteContractAddressToFile(fp, address.String()); err != nil {
			return nil, ethcommon.Address{}, fmt.Errorf("failed to write contract address to file: %w", err)
		}
	} else {
		var err error
		sf, err = getSwapFactory(ec, address)
		if err != nil {
			return nil, ethcommon.Address{}, err
		}
		log.Infof("loaded SwapFactory.sol from address %s", address)

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

	expectedCode := ethcommon.FromHex(swapfactory.SwapFactoryMetaData.Bin)
	if !bytes.Contains(expectedCode, code) {
		return errInvalidSwapContract
	}

	return nil
}

func getSwapFactory(client *ethclient.Client, addr ethcommon.Address) (*swapfactory.SwapFactory, error) {
	return swapfactory.NewSwapFactory(addr, client)
}

func deploySwapFactory(client *ethclient.Client, txOpts *bind.TransactOpts) (ethcommon.Address, *ethtypes.Transaction, *swapfactory.SwapFactory, error) { //nolint:lll
	return swapfactory.DeploySwapFactory(txOpts, client)
}
