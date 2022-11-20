package xmrtaker

import (
	"fmt"

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

	// write keys to file in case something goes wrong
	if err := s.Backend.RecoveryDB().PutSharedSwapPrivateKey(s.ID(), kpAB, s.Env()); err != nil {
		return "", err
	}

	s.XMRClient().Lock()
	defer s.XMRClient().Unlock()

	addr, err := monero.CreateWallet("xmrtaker-swap-wallet", s.Env(), s.Backend.XMRClient(), kpAB, s.walletScanHeight)
	if err != nil {
		return "", err
	}

	if !s.transferBack {
		log.Infof("monero claimed in account %s", addr)
		return addr, nil
	}

	id := s.ID()
	depositAddr, err := s.XMRDepositAddress(&id)
	if err != nil {
		return "", err
	}

	log.Infof("monero claimed in account %s; transferring to original account %s",
		addr, depositAddr)

	err = mcrypto.ValidateAddress(string(depositAddr), s.Env())
	if err != nil {
		log.Errorf("failed to transfer to original account, address %s is invalid", addr)
		return addr, nil
	}

	err = s.waitUntilBalanceUnlocks()
	if err != nil {
		return "", fmt.Errorf("failed to wait for balance to unlock: %w", err)
	}

	_, err = s.XMRClient().SweepAll(depositAddr, 0)
	if err != nil {
		return "", fmt.Errorf("failed to send funds to original account: %w", err)
	}

	close(s.claimedCh)
	return addr, nil
}

func (s *swapState) waitUntilBalanceUnlocks() error {
	for {
		if s.ctx.Err() != nil {
			return s.ctx.Err()
		}

		log.Infof("checking if balance unlocked...")
		balance, err := s.XMRClient().GetBalance(0)
		if err != nil {
			return fmt.Errorf("failed to get balance: %w", err)
		}

		if balance.Balance == balance.UnlockedBalance {
			return nil
		}
		if _, err = monero.WaitForBlocks(s.ctx, s.XMRClient(), int(balance.BlocksToUnlock)); err != nil {
			log.Warnf("Waiting for %d monero blocks failed: %s", balance.BlocksToUnlock, err)
		}
	}
}
