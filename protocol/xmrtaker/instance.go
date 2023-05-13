// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package xmrtaker

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/ChainSafe/chaindb"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/txsender"
)

var (
	log = logging.Logger("xmrtaker")
)

// Instance implements the functionality that will be used by a user who owns ETH
// and wishes to swap for XMR.
type Instance struct {
	backend backend.Backend
	dataDir string

	noTransferBack bool // leave XMR in per-swap generated wallet

	// non-nil if a swap is currently happening, nil otherwise
	// map of offer IDs -> ongoing swaps
	swapStates map[types.Hash]*swapState
	swapMu     sync.RWMutex // lock for above map
}

// Config contains the configuration values for a new XMRTaker instance.
type Config struct {
	Backend        backend.Backend
	DataDir        string
	NoTransferBack bool
	ExternalSender bool
}

// NewInstance returns a new instance of XMRTaker.
// It accepts an endpoint to a monero-wallet-rpc instance where XMRTaker will generate
// the account in which the XMR will be deposited.
func NewInstance(cfg *Config) (*Instance, error) {
	inst := &Instance{
		backend:    cfg.Backend,
		dataDir:    cfg.DataDir,
		swapStates: make(map[types.Hash]*swapState),
	}

	err := inst.checkForOngoingSwaps()
	if err != nil {
		return nil, err
	}

	return inst, nil
}

func (inst *Instance) checkForOngoingSwaps() error {
	swaps, err := inst.backend.SwapManager().GetOngoingSwaps()
	if err != nil {
		return err
	}

	for _, s := range swaps {
		if s.Provides != coins.ProvidesETH {
			continue
		}

		// in this case, we exited either before locking funds, or before the newSwap tx
		// was included in the chain.
		if s.Status == types.ExpectingKeys {
			txHash, err := inst.backend.RecoveryDB().GetNewSwapTxHash(s.OfferID) //nolint:govet
			if err != nil && errors.Is(err, chaindb.ErrKeyNotFound) {
				// since there was no newSwap tx hash, it means there was never an attempt to lock funds,
				// and we can safely abort the swap.
				log.Infof("found ongoing swap %s in DB, aborting since no funds were locked", s.OfferID)
				err = inst.abortOngoingSwap(s)
				if err != nil {
					return fmt.Errorf("failed to abort ongoing swap %s: %s", s.OfferID, err)
				}

				continue
			} else if err != nil {
				return fmt.Errorf("failed to get newSwap tx hash for ongoing swap %s: %w", s.OfferID, err)
			}

			// we have a newSwap tx hash, so we need to check if it was included in the chain.
			err = inst.refundOrCancelNewSwap(s, txHash)
			if err != nil {
				return fmt.Errorf("failed to refund or cancel swap %s: %w", s.OfferID, err)
			}

			continue
		}

		if s.Status == types.SweepingXMR {
			log.Infof(
				"found ongoing swap %s in DB where XMR was being swept back to the primary account, marking as completed",
				s.OfferID,
			)
			s.Status = types.CompletedSuccess
			err = inst.backend.SwapManager().CompleteOngoingSwap(s)
			if err != nil {
				return fmt.Errorf("failed to mark swap as completed: %w", err)
			}

			continue
		}

		err = inst.createOngoingSwap(s)
		if err != nil {
			log.Errorf("%s", err)
			continue
		}
	}

	return nil
}

func (inst *Instance) abortOngoingSwap(s *swap.Info) error {
	// set status to aborted, delete info from recovery db
	s.Status = types.CompletedAbort
	err := inst.backend.SwapManager().CompleteOngoingSwap(s)
	if err != nil {
		return err
	}

	return inst.backend.RecoveryDB().DeleteSwap(s.OfferID)
}

// refundOrCancelNewSwap checks if the newSwap tx was included in the chain.
// if it was, it attempts to refund the swap.
// otherwise, it attempts to cancel the swap by sending a zero-value transfer to our own account
// with the same nonce.
func (inst *Instance) refundOrCancelNewSwap(s *swap.Info, txHash ethcommon.Hash) error {
	log.Infof("found ongoing swap %s with status %s in DB, checking to either refund or cancel", s.OfferID, s.Status)

	cancelled, err := inst.maybeCancelNewSwap(txHash)
	if err != nil {
		return fmt.Errorf("failed to maybe cancel newSwap: %w", err)
	}

	if cancelled {
		return nil
	}

	receipt, err := block.WaitForReceipt(inst.backend.Ctx(), inst.backend.ETHClient().Raw(), txHash)
	if err != nil {
		return fmt.Errorf("failed to get newSwap transaction receipt: %w", err)
	}

	if len(receipt.Logs) == 0 {
		return errSwapInstantiationNoLogs
	}

	var t1 *big.Int
	var t2 *big.Int
	for _, log := range receipt.Logs {
		t1, t2, err = contracts.GetTimeoutsFromLog(log)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("timeouts not found in transaction receipt's logs: %w", err)
	}

	// we have a tx hash, so we can assume that the swap is ongoing
	params, err := getNewSwapParametersFromTx(inst.backend.Ctx(), inst.backend.ETHClient().Raw(), txHash)
	if err != nil {
		return fmt.Errorf("failed to get newSwap parameters from tx %s: %w", txHash, err)
	}

	swap := contracts.SwapCreatorSwap{
		Owner:        params.owner,
		Claimer:      params.claimer,
		PubKeyClaim:  params.cmtXMRMaker,
		PubKeyRefund: params.cmtXMRTaker,
		Timeout1:     t1,
		Timeout2:     t2,
		Asset:        params.asset,
		Value:        params.value,
		Nonce:        params.nonce,
	}

	// our secret value
	secret, err := inst.backend.RecoveryDB().GetSwapPrivateKey(s.OfferID)
	if err != nil {
		return fmt.Errorf("failed to get private key for ongoing swap from db with offer id %s: %w",
			s.OfferID, err)
	}

	swapCreator, err := contracts.NewSwapCreator(params.swapCreatorAddr, inst.backend.ETHClient().Raw())
	if err != nil {
		return fmt.Errorf("failed to instantiate SwapCreator contract: %w", err)
	}

	stage, err := swapCreator.Swaps(nil, swap.SwapID())
	if err != nil {
		return fmt.Errorf("failed to get swap stage: %w", err)
	}

	// TODO: if this is not the case, then the swap has probably already been claimed.
	// however, this seems very unlikely to happen, as the swap counterparty will exit
	// the swap as soon as the network stream is closed, and they will likely not have
	// locked XMR or claimed.
	if stage != contracts.StagePending {
		return fmt.Errorf("swap %s is not in pending stage, aborting", s.OfferID)
	}

	// TODO: check for t1/t2? if between t1 and t2, we need to wait for t2

	txOpts, err := inst.backend.ETHClient().TxOpts(inst.backend.Ctx())
	if err != nil {
		return fmt.Errorf("failed to get tx opts: %w", err)
	}

	refundTx, err := swapCreator.Refund(txOpts, swap, [32]byte(common.Reverse(secret.Bytes())))
	if err != nil {
		return fmt.Errorf("failed to create refund tx: %w", err)
	}

	log.Infof("submit refund tx %s for swap %s", refundTx.Hash(), s.OfferID)
	receipt, err = block.WaitForReceipt(inst.backend.Ctx(), inst.backend.ETHClient().Raw(), refundTx.Hash())
	if err != nil {
		return fmt.Errorf("failed to get refund transaction receipt: %w", err)
	}

	log.Infof("refunded swap %s successfully: %s", s.OfferID, common.ReceiptInfo(receipt))

	// set status to refunded
	s.Status = types.CompletedRefund
	return inst.backend.SwapManager().CompleteOngoingSwap(s)
}

func (inst *Instance) maybeCancelNewSwap(txHash ethcommon.Hash) (bool, error) {
	tx, isPending, err := inst.backend.ETHClient().Raw().TransactionByHash(inst.backend.Ctx(), txHash)
	if err != nil {
		return false, err
	}

	if !isPending {
		// tx is already included, so we can't cancel it
		return false, nil
	}

	log.Infof("newSwap tx %s is still pending, attempting to cancel", tx.Hash())

	// just double the gas price for now, this is higher than needed for a replacement tx though
	gasPrice := new(big.Int).Mul(tx.GasPrice(), big.NewInt(2))
	cancelTx, err := inst.backend.ETHClient().CancelTxWithNonce(inst.backend.Ctx(), tx.Nonce(), gasPrice)
	if err != nil {
		return false, fmt.Errorf("failed to create or send cancel tx: %w", err)
	}

	log.Infof("submit cancel tx %s", cancelTx)
	receipt, err := block.WaitForReceipt(inst.backend.Ctx(), inst.backend.ETHClient().Raw(), cancelTx)
	if err != nil {
		return false, fmt.Errorf("failed to get cancel transaction receipt: %w", err)
	}

	// TODO: check for receipt success; there's still a case newSwap might be included
	log.Infof("cancelled newSwap tx %s successfully: %s", tx.Hash(), common.ReceiptInfo(receipt))
	return true, nil
}

func (inst *Instance) createOngoingSwap(s *swap.Info) error {
	log.Infof("found ongoing swap %s with status %s in DB, restarting swap", s.OfferID, s.Status)

	// check if we have shared secret key in db; if so, claim XMR from that
	// otherwise, create new swap state from recovery info
	skB, err := inst.backend.RecoveryDB().GetCounterpartySwapPrivateKey(s.OfferID)
	if err == nil {
		return inst.completeSwap(s, skB)
	}

	ethSwapInfo, err := inst.backend.RecoveryDB().GetContractSwapInfo(s.OfferID)
	if err != nil {
		return fmt.Errorf("failed to get contract info for ongoing swap from db with offer id %s: %w", s.OfferID, err)
	}

	sk, err := inst.backend.RecoveryDB().GetSwapPrivateKey(s.OfferID)
	if err != nil {
		return fmt.Errorf("failed to get private key for ongoing swap from db with offer id %s: %w",
			s.OfferID, err)
	}

	kp, err := sk.AsPrivateKeyPair()
	if err != nil {
		return err
	}

	inst.swapMu.Lock()
	defer inst.swapMu.Unlock()
	ss, err := newSwapStateFromOngoing(
		inst.backend,
		s,
		inst.noTransferBack,
		ethSwapInfo,
		kp,
	)
	if err != nil {
		return fmt.Errorf("failed to create new swap state for ongoing swap, offer id %s: %w", s.OfferID, err)
	}

	inst.swapStates[s.OfferID] = ss

	go func() {
		<-ss.done
		inst.swapMu.Lock()
		defer inst.swapMu.Unlock()
		delete(inst.swapStates, s.OfferID)
	}()

	return nil
}

// completeSwap is called in the case where we find an ongoing swap in the db on startup,
// and the swap already has the counterparty's swap secret stored.
// In this case, we simply claim the XMR, as we have both secrets required.
// It's unlikely for this case to ever be hit, unless the daemon was shut down in-between
// us finding the counterparty's secret and claiming the XMR.
//
// Note: this will use the current value of `noTransferBack` (verses whatever value
// was set when the swap was started). It will also only only recover to the primary
// wallet address, not whatever address was used when the swap was started.
func (inst *Instance) completeSwap(s *swap.Info, skB *mcrypto.PrivateSpendKey) error {
	// fetch our swap private spend key
	skA, err := inst.backend.RecoveryDB().GetSwapPrivateKey(s.OfferID)
	if err != nil {
		return err
	}

	// fetch our swap private view key
	vkA, err := skA.View()
	if err != nil {
		return err
	}

	// fetch counterparty's private view key
	_, vkB, err := inst.backend.RecoveryDB().GetCounterpartySwapKeys(s.OfferID)
	if err != nil {
		return err
	}

	kpAB := pcommon.GetClaimKeypair(
		skA, skB,
		vkA, vkB,
	)

	err = pcommon.ClaimMonero(
		inst.backend.Ctx(),
		inst.backend.Env(),
		s,
		inst.backend.XMRClient(),
		kpAB,
		inst.backend.XMRClient().PrimaryAddress(),
		inst.noTransferBack,
		inst.backend.SwapManager(),
	)
	if err != nil {
		return err
	}

	s.Status = types.CompletedSuccess
	err = inst.backend.SwapManager().CompleteOngoingSwap(s)
	if err != nil {
		return fmt.Errorf("failed to mark swap %s as completed: %w", s.OfferID, err)
	}

	return nil
}

// GetOngoingSwapState ...
func (inst *Instance) GetOngoingSwapState(offerID types.Hash) common.SwapState {
	inst.swapMu.RLock()
	defer inst.swapMu.RUnlock()
	return inst.swapStates[offerID]
}

// ExternalSender returns the *txsender.ExternalSender for a swap, if the swap exists and is using
// and external tx sender
func (inst *Instance) ExternalSender(offerID types.Hash) (*txsender.ExternalSender, error) {
	inst.swapMu.Lock()
	defer inst.swapMu.Unlock()

	s, has := inst.swapStates[offerID]
	if !has {
		return nil, errNoOngoingSwap
	}

	es, ok := s.sender.(*txsender.ExternalSender)
	if !ok {
		return nil, errSenderIsNotExternal
	}

	return es, nil
}

type newSwapParameters struct {
	swapCreatorAddr ethcommon.Address
	owner           ethcommon.Address
	claimer         ethcommon.Address
	cmtXMRMaker     [32]byte
	cmtXMRTaker     [32]byte
	asset           ethcommon.Address
	value           *big.Int
	nonce           *big.Int
}

func getNewSwapParametersFromTx(
	ctx context.Context,
	ec *ethclient.Client,
	txHash ethcommon.Hash,
) (*newSwapParameters, error) {
	var newSwapTopic = common.GetTopic(common.NewSwapFunctionSignature)

	tx, _, err := ec.TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, err
	}

	if tx.To() == nil {
		return nil, fmt.Errorf("invalid transaction: to address is nil")
	}

	data := tx.Data()
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid transaction data: too short")
	}

	m, err := contracts.SwapCreatorParsedABI.MethodById(newSwapTopic[:4] /*data[:4]*/)
	if err != nil {
		return nil, err
	}

	newSwapInputs := make(map[string]any)

	err = m.Inputs.UnpackIntoMap(newSwapInputs, data[4:])
	if err != nil {
		return nil, err
	}

	owner := newSwapInputs["_owner"].(ethcommon.Address)
	claimer := newSwapInputs["_claimer"].(ethcommon.Address)
	cmtXMRMaker := newSwapInputs["_pubKeyClaim"].([32]byte)
	cmtXMRTaker := newSwapInputs["_pubKeyRefund"].([32]byte)
	asset := newSwapInputs["_asset"].(ethcommon.Address)
	value := newSwapInputs["_value"].(*big.Int)
	nonce := newSwapInputs["_nonce"].(*big.Int)

	return &newSwapParameters{
		swapCreatorAddr: *tx.To(),
		owner:           owner,
		claimer:         claimer,
		cmtXMRMaker:     cmtXMRMaker,
		cmtXMRTaker:     cmtXMRTaker,
		asset:           asset,
		value:           value,
		nonce:           nonce,
	}, nil
}
