package contracts

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/athanorlabs/go-relayer/impls/gsnforwarder"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log"

	"github.com/athanorlabs/atomic-swap/ethereum/block"
)

var log = logging.Logger("contracts")

// DeploySwapFactoryWithKey deploys the SwapFactory contract using the passed privKey to
// pay for the gas.
func DeploySwapFactoryWithKey(
	ctx context.Context,
	ec *ethclient.Client,
	privkey *ecdsa.PrivateKey,
	forwarderAddress ethcommon.Address,
) (ethcommon.Address, *SwapFactory, error) {

	txOpts, err := newTXOpts(ctx, ec, privkey)
	if err != nil {
		return ethcommon.Address{}, nil, err
	}

	if (forwarderAddress != ethcommon.Address{}) {
		// ensure domain separator is registered
		forwarder, err := gsnforwarder.NewForwarder(forwarderAddress, ec) //nolint:govet // shadow declaration of err
		if err != nil {
			return ethcommon.Address{}, nil, err
		}

		err = registerDomainSeparator(ctx, ec, privkey, forwarderAddress, forwarder)
		if err != nil {
			return ethcommon.Address{}, nil, err
		}
	}

	address, tx, sf, err := DeploySwapFactory(txOpts, ec, forwarderAddress)
	if err != nil {
		return ethcommon.Address{}, nil, fmt.Errorf("failed to deploy swap factory: %w", err)
	}

	_, err = block.WaitForReceipt(ctx, ec, tx.Hash())
	if err != nil {
		return ethcommon.Address{}, nil, err
	}

	log.Infof("deployed SwapFactory.sol: address=%s tx hash=%s", address, tx.Hash())

	return address, sf, nil
}

// DeployGSNForwarderWithKey deploys and registers the GSN forwarder using the passed
// private key to pay the gas fees.
func DeployGSNForwarderWithKey(
	ctx context.Context,
	ec *ethclient.Client,
	privkey *ecdsa.PrivateKey,
) (ethcommon.Address, error) {

	txOpts, err := newTXOpts(ctx, ec, privkey)
	if err != nil {
		return ethcommon.Address{}, err
	}

	address, tx, contract, err := gsnforwarder.DeployForwarder(txOpts, ec)
	if err != nil {
		return ethcommon.Address{}, fmt.Errorf("failed to deploy Forwarder.sol: %w", err)
	}

	_, err = block.WaitForReceipt(ctx, ec, tx.Hash())
	if err != nil {
		return ethcommon.Address{}, err
	}

	err = registerDomainSeparator(ctx, ec, privkey, address, contract)
	if err != nil {
		return ethcommon.Address{}, err
	}

	return address, nil
}

func registerDomainSeparator(
	ctx context.Context,
	ec *ethclient.Client,
	privkey *ecdsa.PrivateKey,
	address ethcommon.Address,
	contract *gsnforwarder.Forwarder,
) error {

	txOpts, err := newTXOpts(ctx, ec, privkey)
	if err != nil {
		return err
	}

	tx, err := contract.RegisterDomainSeparator(txOpts, gsnforwarder.DefaultName, gsnforwarder.DefaultVersion)
	if err != nil {
		return fmt.Errorf("failed to register domain separator: %w", err)
	}

	_, err = block.WaitForReceipt(ctx, ec, tx.Hash())
	if err != nil {
		return err
	}

	log.Debugf("registered domain separator in forwarder at %s: name=%s version=%s",
		address,
		gsnforwarder.DefaultName,
		gsnforwarder.DefaultVersion,
	)
	return nil
}

func newTXOpts(ctx context.Context, ec *ethclient.Client, privkey *ecdsa.PrivateKey) (*bind.TransactOpts, error) {
	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	txOpts, err := bind.NewKeyedTransactorWithChainID(privkey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to make transactor: %w", err)
	}
	return txOpts, nil
}
