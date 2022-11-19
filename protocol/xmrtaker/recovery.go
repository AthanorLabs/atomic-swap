package xmrtaker

import (
	"context"
	"errors"
	"fmt"

	eth "github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/db"
	"github.com/athanorlabs/atomic-swap/dleq"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	pswap "github.com/athanorlabs/atomic-swap/protocol/swap"
)

var claimedTopic = common.GetTopic(common.ClaimedEventSignature)

type recoveryState struct {
	ss *swapState
}

// NewRecoveryState returns a new *xmrmaker.recoveryState,
// which has methods to either claim ether or reclaim monero from an initiated swap.
func NewRecoveryState(
	b backend.Backend,
	swapID types.Hash,
	dataDir string,
	secret *mcrypto.PrivateSpendKey,
	ethSwapInfo *db.EthereumSwapInfo,
) (*recoveryState, error) {
	kp, err := secret.AsPrivateKeyPair()
	if err != nil {
		return nil, err
	}

	pubkp := kp.PublicKeyPair()

	var sc [32]byte
	copy(sc[:], secret.Bytes())

	// TODO: update to work with ERC20s
	sender, err := b.NewTxSender(types.EthAssetETH.Address(), nil)
	if err != nil {
		return nil, err
	}

	moneroHeight, err := b.RecoveryDB().GetMoneroStartHeight(swapID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(b.Ctx())
	s := &swapState{
		ctx:              ctx,
		cancel:           cancel,
		Backend:          b,
		sender:           sender,
		privkeys:         kp,
		pubkeys:          pubkp,
		dleqProof:        dleq.NewProofWithSecret(sc),
		walletScanHeight: moneroHeight,
		contractSwapID:   ethSwapInfo.SwapID,
		contractSwap:     ethSwapInfo.Swap,
		claimedCh:        make(chan struct{}),
		info:             pswap.NewEmptyInfo(),
		eventCh:          make(chan Event),
	}

	rs := &recoveryState{
		ss: s,
	}

	rs.ss.setTimeouts(ethSwapInfo.Swap.Timeout0, ethSwapInfo.Swap.Timeout1)

	// TODO: scan for events only starting from ethSwapInfo.StartHeight
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
	// TODO: also run runContractEventWatcher to watch for Claimed logs?
	// will address in recovery refactor (#212)
	go rs.ss.runT1ExpirationHandler()

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
