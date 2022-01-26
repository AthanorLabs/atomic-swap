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
	privkey *ecdsa.PrivateKey, ec *ethclient.Client) (*swapfactory.SwapFactory, error) {
	var (
		sf  *swapfactory.SwapFactory
		err error
	)

	if env == common.Development && address.String() != "" {
		txOpts, err := bind.NewKeyedTransactorWithChainID(privkey, chainID)
		if err != nil {
			return nil, err
		}

		// deploy SwapFactory.sol
		var (
			addr ethcommon.Address
			tx   *ethtypes.Transaction
		)

		addr, tx, sf, err = deploySwapFactory(ec, txOpts)
		if err != nil {
			return nil, err
		}

		log.Infof("deployed SwapFactory.sol: address=%s tx hash=%s", addr, tx.Hash())
	} else {
		sf, err = getSwapFactory(ec, address)
		if err != nil {
			return nil, err
		}
	}

	return sf, nil
}

func getSwapFactory(client *ethclient.Client, addr ethcommon.Address) (*swapfactory.SwapFactory, error) {
	return swapfactory.NewSwapFactory(addr, client)
}

func deploySwapFactory(client *ethclient.Client, txOpts *bind.TransactOpts) (ethcommon.Address, *ethtypes.Transaction, *swapfactory.SwapFactory, error) {
	return swapfactory.DeploySwapFactory(txOpts, client)
}
