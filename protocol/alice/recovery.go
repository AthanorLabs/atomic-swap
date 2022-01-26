package alice

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	eth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"

	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	"github.com/noot/atomic-swap/dleq"
	"github.com/noot/atomic-swap/swapfactory"
)

var claimedTopic = ethcommon.HexToHash("0xd5a2476fc450083bbb092dd3f4be92698ffdc2d213e6f1e730c7f44a52f1ccfc")

var (
	errNoClaimLogsFound = errors.New("no Claimed logs found")
)

type recoveryState struct {
	ss           *swapState
	contractAddr ethcommon.Address
}

// NewRecoveryState returns a new *bob.recoveryState,
// which has methods to either claim ether or reclaim monero from an initiated swap.
func NewRecoveryState(a *Instance, secret *mcrypto.PrivateSpendKey,
	contractAddr ethcommon.Address, contractSwapID *big.Int) (*recoveryState, error) { //nolint:revive
	txOpts, err := bind.NewKeyedTransactorWithChainID(a.ethPrivKey, a.chainID)
	if err != nil {
		return nil, err
	}

	kp, err := secret.AsPrivateKeyPair()
	if err != nil {
		return nil, err
	}

	pubkp := kp.PublicKeyPair()

	txOpts.GasPrice = a.gasPrice
	txOpts.GasLimit = a.gasLimit

	var sc [32]byte
	copy(sc[:], secret.Bytes())

	ctx, cancel := context.WithCancel(a.ctx)
	s := &swapState{
		ctx:            ctx,
		cancel:         cancel,
		alice:          a,
		txOpts:         txOpts,
		privkeys:       kp,
		pubkeys:        pubkp,
		dleqProof:      dleq.NewProofWithSecret(sc),
		contractSwapID: contractSwapID,
	}

	rs := &recoveryState{
		ss: s,
	}

	if err := rs.setContract(contractAddr); err != nil {
		return nil, err
	}

	if err := rs.ss.setTimeouts(); err != nil {
		return nil, err
	}

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

// ClaimOrRecover either claims ether or recovers monero by creating a wallet.
// It returns a *RecoveryResult.
func (rs *recoveryState) ClaimOrRefund() (*RecoveryResult, error) {
	// check if Bob claimed
	skA, err := rs.filterForClaim()
	if !errors.Is(err, errNoClaimLogsFound) && err != nil {
		return nil, err
	}

	log.Info("found claim log??", skA != nil)

	// if Bob claimed, let's get our monero
	if skA != nil {
		vkA, err := skA.View() //nolint:govet
		if err != nil {
			return nil, err
		}

		rs.ss.setBobKeys(skA.Public(), vkA, nil)

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

// setContract sets the contract in which Alice has locked her ETH.
func (rs *recoveryState) setContract(address ethcommon.Address) error {
	var err error
	rs.contractAddr = address
	rs.ss.alice.contract, err = swapfactory.NewSwapFactory(address, rs.ss.alice.ethClient)
	return err
}

func (rs *recoveryState) filterForClaim() (*mcrypto.PrivateSpendKey, error) {
	logs, err := rs.ss.alice.ethClient.FilterLogs(rs.ss.ctx, eth.FilterQuery{
		Addresses: []ethcommon.Address{rs.contractAddr},
		Topics:    [][]ethcommon.Hash{{claimedTopic}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %w", err)
	}

	if len(logs) == 0 {
		return nil, errNoClaimLogsFound
	}

	sa, err := swapfactory.GetSecretFromLog(&logs[0], "Claimed")
	if err != nil {
		return nil, fmt.Errorf("failed to get secret from log: %w", err)
	}

	return sa, nil
}
