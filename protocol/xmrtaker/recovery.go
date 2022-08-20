package xmrtaker

import (
	"context"
	"errors"
	"fmt"

	eth "github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/dleq"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	"github.com/athanorlabs/atomic-swap/swapfactory"
)

// TODO: don't hard-code this
var claimedTopic = ethcommon.HexToHash("0x38d6042dbdae8e73a7f6afbabd3fbe0873f9f5ed3cd71294591c3908c2e65fee")

type recoveryState struct {
	ss *swapState
}

// NewRecoveryState returns a new *xmrmaker.recoveryState,
// which has methods to either claim ether or reclaim monero from an initiated swap.
func NewRecoveryState(b backend.Backend, basePath string, secret *mcrypto.PrivateSpendKey,
	contractSwapID [32]byte, contractSwap swapfactory.SwapFactorySwap) (*recoveryState, error) {
	kp, err := secret.AsPrivateKeyPair()
	if err != nil {
		return nil, err
	}

	pubkp := kp.PublicKeyPair()

	var sc [32]byte
	copy(sc[:], secret.Bytes())

	ctx, cancel := context.WithCancel(b.Ctx())
	s := &swapState{
		ctx:            ctx,
		cancel:         cancel,
		Backend:        b,
		privkeys:       kp,
		pubkeys:        pubkp,
		dleqProof:      dleq.NewProofWithSecret(sc),
		contractSwapID: contractSwapID,
		contractSwap:   contractSwap,
		infoFile:       pcommon.GetSwapRecoveryFilepath(basePath),
		claimedCh:      make(chan struct{}),
	}

	rs := &recoveryState{
		ss: s,
	}

	rs.ss.setTimeouts(contractSwap.Timeout0, contractSwap.Timeout1)
	return rs, nil
}

// RecoveryResult represents the result of a recovery operation.
// If the ether was refunded, Refunded is set to true and the TxHash is set.
// If the monero was claimed, Claimed is set to true and the MoneroAddress is set.
type RecoveryResult struct {
	Refunded, Claimed bool
	TxHash            ethcommon.Hash
	MoneroAddress     mcrypto.Address
}

// ClaimOrRefund either claims the monero or recovers the ether returning a *RecoveryResult.
func (rs *recoveryState) ClaimOrRefund() (*RecoveryResult, error) {
	// check if XMRMaker claimed
	skA, err := rs.ss.filterForClaim()
	if !errors.Is(err, errNoClaimLogsFound) && err != nil {
		return nil, err
	}

	// if XMRMaker claimed, let's get our monero
	if skA != nil {
		vkA, err := skA.View() //nolint:govet
		if err != nil {
			return nil, err
		}

		rs.ss.setXMRMakerKeys(skA.Public(), vkA, nil)

		addr, err := rs.ss.claimMonero(skA)
		if err != nil {
			return nil, err
		}

		return &RecoveryResult{
			Claimed:       true,
			MoneroAddress: addr,
		}, nil
	}

	// otherwise, let's try to refund
	txHash, err := rs.ss.tryRefund()
	if err != nil {
		return nil, err
	}

	return &RecoveryResult{
		Refunded: true,
		TxHash:   txHash,
	}, nil
}

func (s *swapState) filterForClaim() (*mcrypto.PrivateSpendKey, error) {
	const claimedEvent = "Claimed"

	logs, err := s.FilterLogs(s.ctx, eth.FilterQuery{
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
		matches, err := swapfactory.CheckIfLogIDMatches(log, claimedEvent, s.contractSwapID) //nolint:govet
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

	sa, err := swapfactory.GetSecretFromLog(&foundLog, claimedEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret from log: %w", err)
	}

	return sa, nil
}
