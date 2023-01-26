package xmrtaker

import (
	"context"
	"fmt"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/monero"

	eth "github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

func (s *swapState) tryClaim() error {
	if !s.info.Status.IsOngoing() {
		return nil
	}

	skA, err := s.filterForClaim()
	if err != nil {
		return err
	}

	addr, err := s.claimMonero(skA)
	if err != nil {
		return err
	}

	log.Infof("claimed monero: address=%s", addr)
	s.clearNextExpectedEvent(types.CompletedSuccess)
	return nil
}

func (s *swapState) filterForClaim() (*mcrypto.PrivateSpendKey, error) {
	const claimedEvent = "Claimed"

	logs, err := s.ETHClient().Raw().FilterLogs(s.ctx, eth.FilterQuery{
		Addresses: []ethcommon.Address{s.ContractAddr()},
		Topics:    [][]ethcommon.Hash{{claimedTopic}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %w", err)
	}

	if len(logs) == 0 {
		return nil, errNoClaimLogsFound
	}

	var (
		foundLog ethtypes.Log
		found    bool
	)

	for _, log := range logs {
		matches, err := contracts.CheckIfLogIDMatches(log, claimedEvent, s.contractSwapID) //nolint:govet
		if err != nil {
			continue
		}

		if matches {
			foundLog = log
			found = true
			break
		}
	}

	if !found {
		return nil, errNoClaimLogsFound
	}

	sa, err := contracts.GetSecretFromLog(&foundLog, claimedEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret from log: %w", err)
	}

	return sa, nil
}

func (s *swapState) claimMonero(skB *mcrypto.PrivateSpendKey) (mcrypto.Address, error) {
	if !s.info.Status.IsOngoing() {
		return "", errSwapCompleted
	}

	skAB := mcrypto.SumPrivateSpendKeys(skB, s.privkeys.SpendKey())
	vkAB := mcrypto.SumPrivateViewKeys(s.xmrmakerPrivateViewKey, s.privkeys.ViewKey())
	kpAB := mcrypto.NewPrivateKeyPair(skAB, vkAB)
	abAddr := kpAB.PublicKeyPair().Address(s.Env())

	// write keys to file in case something goes wrong
	err := s.Backend.RecoveryDB().PutSharedSwapPrivateKey(s.ID(), kpAB.SpendKey())
	if err != nil {
		return "", err
	}

	conf := s.XMRClient().CreateABWalletConf("xmrtaker-swap-wallet-claim")
	abWalletCli, err := monero.CreateSpendWalletFromKeys(conf, kpAB, s.walletScanHeight)
	if err != nil {
		return "", err
	}

	if s.transferBack {
		defer abWalletCli.CloseAndRemoveWallet()
	} else {
		abWalletCli.Close()
		log.Infof("monero claimed in account %s with wallet file %s", abAddr, conf.WalletFilePath)
		return abAddr, nil
	}

	id := s.ID()
	depositAddr, err := s.XMRDepositAddress(&id)
	if err != nil {
		return "", err
	}

	log.Infof("monero claimed in account %s; transferring to original account %s",
		abAddr, depositAddr)

	err = mcrypto.ValidateAddress(string(depositAddr), s.Env())
	if err != nil {
		log.Errorf("Failed to transfer XMR out of swap wallet, dest address %s is invalid: %s", abAddr, err)
		return "", err
	}

	err = waitUntilBalanceUnlocks(s.ctx, abWalletCli)
	if err != nil {
		return "", fmt.Errorf("failed to wait for balance to unlock: %w", err)
	}

	transfers, err := s.XMRClient().SweepAll(s.ctx, depositAddr, 0, monero.SweepToSelfConfirmations)
	if err != nil {
		return "", fmt.Errorf("failed to send funds to original account: %w", err)
	}
	for _, transfer := range transfers {
		log.Infof("Moved %s XMR claimed from swap to primary wallet (%s XMR lost to fees)",
			coins.FmtPiconeroAmtAsXMR(transfer.Amount), coins.FmtPiconeroAmtAsXMR(transfer.Fee))
	}

	close(s.claimedCh)
	return abAddr, nil
}

// TODO: Put this in monero package?  Unit test.
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
			log.Warnf("Waiting for %d monero blocks failed: %s", balance.BlocksToUnlock, err)
		}
	}
}
