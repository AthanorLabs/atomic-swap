// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package xmrtaker

import (
	"fmt"

	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"

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
		matches, err := contracts.CheckIfLogIDMatches(log, claimedTopic, s.contractSwapID) //nolint:govet
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

	sa, err := contracts.GetSecretFromLog(&foundLog, claimedTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret from log: %w", err)
	}

	return sa, nil
}

func (s *swapState) claimMonero(skB *mcrypto.PrivateSpendKey) (*mcrypto.Address, error) {
	if !s.info.Status.IsOngoing() {
		return nil, errSwapCompleted
	}

	// write counterparty swap privkey to disk in case something goes wrong
	err := s.Backend.RecoveryDB().PutCounterpartySwapPrivateKey(s.ID(), skB)
	if err != nil {
		return nil, err
	}

	id := s.ID()
	depositAddr := s.XMRDepositAddress(&id)
	if s.noTransferBack {
		depositAddr = nil
	}

	kpAB := pcommon.GetClaimKeypair(
		skB, s.privkeys.SpendKey(),
		s.xmrmakerPrivateViewKey, s.privkeys.ViewKey(),
	)

	err = pcommon.ClaimMonero(
		s.ctx,
		s.Env(),
		s.info.ID,
		s.XMRClient(),
		s.walletScanHeight,
		kpAB,
		depositAddr,
		s.noTransferBack,
	)
	if err != nil {
		return nil, err
	}

	close(s.claimedCh)
	log.Infof("monero claimed and swept to original account %s", depositAddr)
	return kpAB.PublicKeyPair().Address(s.Env()), nil
}
