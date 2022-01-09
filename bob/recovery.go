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

func NewRecoveryState(b *Instance, secret *mcrypto.PrivateSpendKey, contractAddr ethcommon.Address) (*recoveryState, error) {
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

type RecoveryResult struct {
	Claimed, Recovered bool
	TxHash             ethcommon.Hash
	MoneroAddress      mcrypto.Address
}

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
