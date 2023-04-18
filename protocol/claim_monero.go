// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package protocol

import (
	"context"
	"fmt"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/monero"

	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("protocol")
)

// GetClaimKeypair returns the private key pair required for a monero claim.
// The key pair is the summation of each party's private spend and view keys:
// (s_a + s_b) and (v_a + v_b).
func GetClaimKeypair(
	skA, skB *mcrypto.PrivateSpendKey,
	vkA, vkB *mcrypto.PrivateViewKey,
) *mcrypto.PrivateKeyPair {
	skAB := mcrypto.SumPrivateSpendKeys(skA, skB)
	vkAB := mcrypto.SumPrivateViewKeys(vkA, vkB)
	kpAB := mcrypto.NewPrivateKeyPair(skAB, vkAB)
	return kpAB
}

// ClaimMonero claims the XMR located in the wallet controlled by the private keypair `kpAB`.
// If noTransferBack is unset, it sweeps the XMR to `depositAddr`.
func ClaimMonero(
	ctx context.Context,
	env common.Environment,
	id types.Hash,
	xmrClient monero.WalletClient,
	walletScanHeight uint64,
	kpAB *mcrypto.PrivateKeyPair,
	depositAddr *mcrypto.Address,
	noTransferBack bool,
) error {
	conf := xmrClient.CreateWalletConf(fmt.Sprintf("swap-wallet-claim-%s", id))
	abWalletCli, err := monero.CreateSpendWalletFromKeys(conf, kpAB, walletScanHeight)
	if err != nil {
		return err
	}

	address := kpAB.PublicKeyPair().Address(env)
	if noTransferBack {
		abWalletCli.Close()
		log.Infof("monero claimed in account %s with wallet file %s", address, conf.WalletFilePath)
		return nil
	}
	defer abWalletCli.CloseAndRemoveWallet()

	log.Infof("monero claimed in account %s; transferring to deposit account %s",
		address, depositAddr)

	err = depositAddr.ValidateEnv(env)
	if err != nil {
		log.Errorf(
			"failed to transfer XMR out of swap wallet, dest address %s is invalid: %s",
			address,
			err,
		)
		return err
	}

	transfers, err := abWalletCli.SweepAll(ctx, depositAddr, 0, monero.SweepToSelfConfirmations)
	if err != nil {
		return fmt.Errorf("failed to send funds to deposit account: %w", err)
	}

	log.Debugf("got %d sweep receipts", len(transfers))
	for _, transfer := range transfers {
		log.Infof("transferred %s XMR to primary wallet (%s XMR lost to fees)",
			coins.FmtPiconeroAsXMR(transfer.Amount),
			coins.FmtPiconeroAsXMR(transfer.Fee),
		)
	}

	return nil
}
