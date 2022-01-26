package main

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/swapfactory"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func getOrDeploySwapFactory(address ethcommon.Address, env common.Environment, chainID *big.Int,
	privkey *ecdsa.PrivateKey, ec *ethclient.Client) (*swapfactory.SwapFactory, ethcommon.Address, error) {
	var (
		sf *swapfactory.SwapFactory
	)

	if env == common.Development && (address == ethcommon.Address{}) {
		txOpts, err := bind.NewKeyedTransactorWithChainID(privkey, chainID)
		if err != nil {
			return nil, ethcommon.Address{}, err
		}

		// deploy SwapFactory.sol
		var tx *ethtypes.Transaction
		address, tx, sf, err = deploySwapFactory(ec, txOpts)
		if err != nil {
			return nil, ethcommon.Address{}, err
		}

		log.Infof("deployed SwapFactory.sol: address=%s tx hash=%s", address, tx.Hash())
	} else {
		var err error
		sf, err = getSwapFactory(ec, address)
		if err != nil {
			return nil, ethcommon.Address{}, err
		}
		log.Infof("loaded SwapFactory.sol from address %s", address)
	}

	return sf, address, nil
}

func getSwapFactory(client *ethclient.Client, addr ethcommon.Address) (*swapfactory.SwapFactory, error) {
	return swapfactory.NewSwapFactory(addr, client)
}

func deploySwapFactory(client *ethclient.Client, txOpts *bind.TransactOpts) (ethcommon.Address, *ethtypes.Transaction, *swapfactory.SwapFactory, error) { //nolint:lll
	return swapfactory.DeploySwapFactory(txOpts, client)
}
