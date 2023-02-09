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
// The key pair is the summation of each party's spend and view keys:
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

// ClaimMoneroInAddress claims the XMR located in the wallet controlled by the private keypair `kpAB`
// with the given address.
// If transferBack is true, it sweeps the XMR to `depositAddr`.
func ClaimMoneroInAddress(
	ctx context.Context,
	env common.Environment,
	id types.Hash,
	xmrClient monero.WalletClient,
	walletScanHeight uint64,
	kpAB *mcrypto.PrivateKeyPair,
	address mcrypto.Address,
	depositAddr mcrypto.Address,
	transferBack bool,
) error {
	conf := xmrClient.CreateWalletConf(fmt.Sprintf("swap-wallet-claim-%s", id))
	abWalletCli, err := monero.CreateSpendWalletFromKeysAndAddress(conf, kpAB, address, walletScanHeight)
	if err != nil {
		return err
	}

	if transferBack {
		defer abWalletCli.CloseAndRemoveWallet()
	} else {
		abWalletCli.Close()
		log.Infof("monero claimed in account %s with wallet file %s", address, conf.WalletFilePath)
		return nil
	}

	log.Infof("monero claimed in account %s; transferring to deposit account %s",
		address, depositAddr)

	err = mcrypto.ValidateAddress(string(depositAddr), env)
	if err != nil {
		log.Errorf(
			"failed to transfer XMR out of swap wallet, dest address %s is invalid: %s",
			address,
			err,
		)
		return err
	}

	err = waitUntilBalanceUnlocks(ctx, abWalletCli)
	if err != nil {
		return fmt.Errorf("failed to wait for balance to unlock: %w", err)
	}

	transfers, err := abWalletCli.SweepAll(ctx, depositAddr, 0, monero.SweepToSelfConfirmations)
	if err != nil {
		return fmt.Errorf("failed to send funds to deposit account: %w", err)
	}

	for _, transfer := range transfers {
		log.Infof("transferred %s XMR to primary wallet (%s XMR lost to fees)",
			coins.FmtPiconeroAmtAsXMR(transfer.Amount),
			coins.FmtPiconeroAmtAsXMR(transfer.Fee),
		)
	}

	return nil
}

// ClaimMonero claims the XMR located in the wallet controlled by the private keypair `kpAB`.
// If transferBack is true, it sweeps the XMR to `depositAddr`.
func ClaimMonero(
	ctx context.Context,
	env common.Environment,
	id types.Hash,
	xmrClient monero.WalletClient,
	walletScanHeight uint64,
	kpAB *mcrypto.PrivateKeyPair,
	depositAddr mcrypto.Address,
	transferBack bool,
) (mcrypto.Address, error) {
	abAddr := kpAB.PublicKeyPair().Address(env)

	err := ClaimMoneroInAddress(
		ctx, env, id, xmrClient, walletScanHeight, kpAB, abAddr, depositAddr, transferBack,
	)
	if err != nil {
		return "", err
	}

	return abAddr, nil
}

// TODO: Put this in monero package? Unit test.
func waitUntilBalanceUnlocks(ctx context.Context, walletCli monero.WalletClient) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		log.Infof("checking if balance unlocked...")
		balance, err := walletCli.GetBalance(0)
		if err != nil {
			return fmt.Errorf("failed to get balance: %w", err)
		}

		if balance.Balance == balance.UnlockedBalance {
			return nil
		}

		if _, err = monero.WaitForBlocks(ctx, walletCli, int(balance.BlocksToUnlock)); err != nil {
			log.Warnf("waiting for %d monero blocks failed: %s", balance.BlocksToUnlock, err)
		}
	}
}
