package extethclient

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
)

type transferConfig struct {
	ctx      context.Context
	ec       *ethclient.Client
	pk       *ecdsa.PrivateKey
	destAddr ethcommon.Address
	amount   *coins.WeiAmount
	gasLimit uint64
	gasPrice *coins.WeiAmount
	nonce    uint64
}

// transfer handles almost any use case for transferring ETH by having all the
// configurable values (nonce, gas-price, etc.) set by the caller.
func transfer(cfg *transferConfig) (*ethtypes.Receipt, error) {
	ctx := cfg.ctx
	ec := cfg.ec

	tx := ethtypes.NewTx(&ethtypes.LegacyTx{
		Nonce:    cfg.nonce,
		To:       &cfg.destAddr,
		Value:    cfg.amount.BigInt(),
		Gas:      cfg.gasLimit,
		GasPrice: cfg.gasPrice.BigInt(),
	})

	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	signer := ethtypes.LatestSignerForChainID(chainID)
	signedTx, err := ethtypes.SignTx(tx, signer, cfg.pk)

	if err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	txHash := signedTx.Hash()

	err = ec.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transfer transaction: %w", err)
	}

	log.Infof("transfer of %s ETH to %s sent to mempool with txID %s, nonce %d, gas-price %s ETH",
		cfg.amount.AsStdString(), cfg.destAddr, txHash, cfg.nonce, cfg.gasPrice.AsStdString())

	receipt, err := block.WaitForReceipt(ctx, ec, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed waiting for txID %s receipt: %w", txHash, err)
	}

	log.Infof("transfer included in chain %s", common.ReceiptInfo(receipt))

	return receipt, nil
}
