package xmrmaker

import (
	"context"
	"fmt"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net/message"
)

func convertContractSwap(msg *message.ContractSwap) contracts.SwapFactorySwap {
	return contracts.SwapFactorySwap{
		Owner:        msg.Owner,
		Claimer:      msg.Claimer,
		PubKeyClaim:  msg.PubKeyClaim,
		PubKeyRefund: msg.PubKeyRefund,
		Timeout0:     msg.Timeout0,
		Timeout1:     msg.Timeout1,
		Asset:        msg.Asset,
		Value:        msg.Value,
		Nonce:        msg.Nonce,
	}
}

func sweepRefund(
	ctx context.Context,
	env common.Environment,
	swapID types.Hash,
	abWalletConf *monero.WalletClientConf,
	walletRestoreHeight uint64,
	abWalletKey *mcrypto.PrivateKeyPair,
	refundDest mcrypto.Address,
) error {
	log.Infof("Refunding XMR from swap ID %s to address %s", swapID, abWalletKey.Address(env))
	abCli, err := monero.CreateSpendWalletFromKeys(abWalletConf, abWalletKey, walletRestoreHeight)
	if err != nil {
		return err
	}
	defer abCli.CloseAndRemoveWallet()

	transfers, err := abCli.SweepAll(ctx, refundDest, 0, monero.SweepToSelfConfirmations)
	if err != nil {
		return fmt.Errorf("swap ID %s failed to recover refund: %w", swapID, err)
	}
	for _, transfer := range transfers {
		// It is unlikely that anyone will ever see more than one of these messages
		log.Infof("Refunded %s XMR from swap to primary wallet (%s XMR lost to fees)",
			coins.FmtPiconeroAmtAsXMR(transfer.Amount), coins.FmtPiconeroAmtAsXMR(transfer.Fee))
	}
	return nil
}
