package main

import (
	"context"
	"fmt"
	"math/big"

	rcommon "github.com/athanorlabs/go-relayer/common"
	rcontracts "github.com/athanorlabs/go-relayer/impls/gsnforwarder"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	contracts "github.com/athanorlabs/atomic-swap/ethereum"
)

func deployOrGetForwarder(
	ctx context.Context,
	addressString string,
	ec *ethclient.Client,
	key *rcommon.Key,
	chainID *big.Int,
) (*rcontracts.IForwarder, ethcommon.Address, error) {
	txOpts, err := bind.NewKeyedTransactorWithChainID(key.PrivateKey(), chainID)
	if err != nil {
		return nil, ethcommon.Address{}, fmt.Errorf("failed to make transactor: %w", err)
	}

	if addressString == "" {
		address, tx, _, err := rcontracts.DeployForwarder(txOpts, ec) //nolint:govet
		if err != nil {
			return nil, ethcommon.Address{}, err
		}

		_, err = bind.WaitMined(ctx, ec, tx)
		if err != nil {
			return nil, ethcommon.Address{}, err
		}

		log.Infof("deployed Forwarder.sol to %s", address)
		f, err := rcontracts.NewIForwarder(address, ec)
		if err != nil {
			return nil, ethcommon.Address{}, err
		}

		return f, address, nil
	}

	ok := ethcommon.IsHexAddress(addressString)
	if !ok {
		return nil, ethcommon.Address{}, errInvalidAddress
	}

	address := ethcommon.HexToAddress(addressString)
	err = contracts.CheckForwarderContractCode(context.Background(), ec, address)
	if err != nil {
		return nil, ethcommon.Address{}, err
	}

	log.Infof("loaded Forwarder.sol at %s", address)
	f, err := rcontracts.NewIForwarder(address, ec)
	if err != nil {
		return nil, ethcommon.Address{}, err
	}

	return f, address, nil
}
