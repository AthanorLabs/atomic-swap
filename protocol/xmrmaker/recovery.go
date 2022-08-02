package xmrmaker

import (
	"context"
	"errors"

	ethcommon "github.com/ethereum/go-ethereum/common"

	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	"github.com/noot/atomic-swap/dleq"
	pcommon "github.com/noot/atomic-swap/protocol"
	"github.com/noot/atomic-swap/protocol/backend"
	"github.com/noot/atomic-swap/swapfactory"
)

type recoveryState struct {
	ss *swapState
}

// NewRecoveryState returns a new *xmrmaker.recoveryState,
// which has methods to either claim ether or reclaim monero from an initiated swap.
func NewRecoveryState(b backend.Backend, basePath string, secret *mcrypto.PrivateSpendKey,
	contractAddr ethcommon.Address,
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
	}

	if err := s.setContract(contractAddr); err != nil {
		return nil, err
	}

	s.setTimeouts(contractSwap.Timeout0, contractSwap.Timeout1)
	return &recoveryState{
		ss: s,
	}, nil
}

// RecoveryResult represents the result of a recovery operation.
// If the ether was claimed, Claimed is set to true and the TxHash is set.
// If the monero was recovered, Recovered is set to true and the MoneroAddress is set.
type RecoveryResult struct {
	Claimed, Recovered bool
	TxHash             ethcommon.Hash
	MoneroAddress      mcrypto.Address
}

// ClaimOrRecover either claims ether or recovers monero by creating a wallet.
// It returns a *RecoveryResult.
func (rs *recoveryState) ClaimOrRecover() (*RecoveryResult, error) {
	// check if XMRTaker refunded
	skA, err := rs.ss.filterForRefund()
	if !errors.Is(err, errNoRefundLogsFound) && err != nil {
		return nil, err
	}

	// if XMRTaker refunded, let's get our monero back
	if skA != nil {
		kpA, err := skA.AsPrivateKeyPair() //nolint:govet
		if err != nil {
			return nil, err
		}

		rs.ss.setXMRTakerPublicKeys(kpA.PublicKeyPair(), nil)
		addr, err := rs.ss.reclaimMonero(skA)
		if err != nil {
			return nil, err
		}

		return &RecoveryResult{
			Recovered:     true,
			MoneroAddress: addr,
		}, nil
	}

	// otherwise, let's try to claim
	txHash, err := rs.ss.tryClaim()
	if err != nil {
		if errors.Is(err, errClaimPastTime) {
			log.Infof(
				"Past the time where we can claim the ether, and the counterparty " +
					"has not yet refunded. Please try running the recovery module again later " +
					"and hopefully the counterparty will have refunded by then.",
			)
		}

		return nil, err
	}

	return &RecoveryResult{
		Claimed: true,
		TxHash:  txHash,
	}, nil
}
