package bob

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"

	mcrypto "github.com/noot/atomic-swap/monero/crypto"
)

type recoveryState struct {
	ss *swapState
}

// NewRecoveryState returns a new *bob.recoveryState,
// which has methods to either claim ether or reclaim monero from an initiated swap.
func NewRecoveryState(b *Instance, secret *mcrypto.PrivateSpendKey,
	contractAddr ethcommon.Address) (*recoveryState, error) { //nolint:revive
	txOpts, err := bind.NewKeyedTransactorWithChainID(b.ethPrivKey, b.chainID)
	if err != nil {
		return nil, err
	}

	kp, err := secret.AsPrivateKeyPair()
	if err != nil {
		return nil, err
	}

	pubkp := kp.PublicKeyPair()

	txOpts.GasPrice = b.gasPrice
	txOpts.GasLimit = b.gasLimit

	ctx, cancel := context.WithCancel(b.ctx)
	s := &swapState{
		ctx:      ctx,
		cancel:   cancel,
		bob:      b,
		txOpts:   txOpts,
		privkeys: kp,
		pubkeys:  pubkp,
	}

	if err := s.setContract(contractAddr); err != nil {
		return nil, err
	}
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
	if err := rs.ss.setTimeouts(); err != nil {
		return nil, err
	}

	// check if Alice refunded
	skA, err := rs.ss.filterForRefund()
	if !errors.Is(err, errNoRefundLogsFound) && err != nil {
		return nil, err
	}

	// if Alice refunded, let's get our monero back
	if skA != nil {
		addr, err := rs.ss.reclaimMonero(skA) //nolint:govet
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
		if errors.Is(err, errPastClaimTime) {
			log.Infof(
				"Past the time where we can claim the ether, and the counterparty" +
					"has not yet refunded. Please try running the recovery module again later" +
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
