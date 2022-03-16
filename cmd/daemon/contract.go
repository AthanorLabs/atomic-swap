package main

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/noot/atomic-swap/common"
	pcommon "github.com/noot/atomic-swap/protocol"
	"github.com/noot/atomic-swap/swapfactory"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func getOrDeploySwapFactory(address ethcommon.Address, env common.Environment, basepath string, chainID *big.Int,
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

		// store the contract address on disk
		fp := fmt.Sprintf("%s/contractaddress", basepath)
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
	}

	return sf, address, nil
}

func getSwapFactory(client *ethclient.Client, addr ethcommon.Address) (*swapfactory.SwapFactory, error) {
	return swapfactory.NewSwapFactory(addr, client)
}

func deploySwapFactory(client *ethclient.Client, txOpts *bind.TransactOpts) (ethcommon.Address, *ethtypes.Transaction, *swapfactory.SwapFactory, error) { //nolint:lll
	return swapfactory.DeploySwapFactory(txOpts, client)
}
