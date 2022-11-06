package contracts

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/athanorlabs/go-relayer/common"
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
	privKey *ecdsa.PrivateKey,
	forwarderAddr ethcommon.Address,
) (ethcommon.Address, *SwapFactory, error) {

	txOpts, err := newTXOpts(ctx, ec, privKey)
	if err != nil {
		return ethcommon.Address{}, nil, err
	}

	if (forwarderAddr != ethcommon.Address{}) {
		if err = registerDomainSeparatorIfNeeded(ctx, ec, privKey, forwarderAddr); err != nil {
			return ethcommon.Address{}, nil, fmt.Errorf("failed to deploy swap factory: %w", err)
		}
	}

	address, tx, sf, err := DeploySwapFactory(txOpts, ec, forwarderAddr)
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
	privKey *ecdsa.PrivateKey,
) (ethcommon.Address, error) {

	txOpts, err := newTXOpts(ctx, ec, privKey)
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

	err = registerDomainSeparator(ctx, ec, privKey, address, contract)
	if err != nil {
		return ethcommon.Address{}, err
	}

	return address, nil
}

func isDomainSeparatorRegistered(
	ctx context.Context,
	ec *ethclient.Client,
	forwarderAddr ethcommon.Address,
	forwarder *gsnforwarder.Forwarder,
) (bool, error) {
	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return false, err
	}
	name := gsnforwarder.DefaultName
	version := gsnforwarder.DefaultVersion
	ds, err := common.GetEIP712DomainSeparator(name, version, chainID, forwarderAddr)
	if err != nil {
		return false, err
	}
	return forwarder.Domains(nil, ds) // isRegistered, error
}

func registerDomainSeparatorIfNeeded(
	ctx context.Context,
	ec *ethclient.Client,
	privKey *ecdsa.PrivateKey,
	forwarderAddr ethcommon.Address,
) error {
	forwarder, err := gsnforwarder.NewForwarder(forwarderAddr, ec)
	if err != nil {
		return err
	}

	isRegistered, err := isDomainSeparatorRegistered(ctx, ec, forwarderAddr, forwarder)
	if err != nil {
		return err
	}

	if !isRegistered {
		err = registerDomainSeparator(ctx, ec, privKey, forwarderAddr, forwarder)
		if err != nil {
			return err
		}
	}

	return nil
}

func registerDomainSeparator(
	ctx context.Context,
	ec *ethclient.Client,
	privKey *ecdsa.PrivateKey,
	forwarderAddr ethcommon.Address,
	forwarder *gsnforwarder.Forwarder,
) error {

	txOpts, err := newTXOpts(ctx, ec, privKey)
	if err != nil {
		return err
	}

	tx, err := forwarder.RegisterDomainSeparator(txOpts, gsnforwarder.DefaultName, gsnforwarder.DefaultVersion)
	if err != nil {
		return fmt.Errorf("failed to register domain separator: %w", err)
	}

	_, err = block.WaitForReceipt(ctx, ec, tx.Hash())
	if err != nil {
		return err
	}

	log.Debugf("registered domain separator in forwarder at %s: name=%s version=%s",
		forwarderAddr,
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
