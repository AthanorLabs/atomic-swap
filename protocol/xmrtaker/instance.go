package xmrtaker

import (
	"fmt"
	"sync"

	ethcommon "github.com/ethereum/go-ethereum/common"
	logging "github.com/ipfs/go-log"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
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

	transferBack bool // transfer xmr back to original account

	// non-nil if a swap is currently happening, nil otherwise
	// map of offer IDs -> ongoing swaps
	swapStates map[types.Hash]*swapState
	swapMu     sync.Mutex // lock for above map
}

// Config contains the configuration values for a new XMRTaker instance.
type Config struct {
	Backend        backend.Backend
	DataDir        string
	TransferBack   bool
	ExternalSender bool
}

// NewInstance returns a new instance of XMRTaker.
// It accepts an endpoint to a monero-wallet-rpc instance where XMRTaker will generate
// the account in which the XMR will be deposited.
func NewInstance(cfg *Config) (*Instance, error) {
	// if this is set, it transfers all xmr received during swaps back to the given wallet.
	if cfg.TransferBack {
		cfg.Backend.SetBaseXMRDepositAddress(cfg.Backend.XMRClient().PrimaryAddress())
	}

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

		if s.Status == types.KeysExchanged || s.Status == types.ExpectingKeys {
			// TODO: set status to aborted, delete info from recovery db
			continue
		}

		err = inst.createOngoingSwap(s)
		if err != nil {
			return err
		}
	}

	return nil
}

func (inst *Instance) createOngoingSwap(s *swap.Info) error {
	// check if we have shared secret key in db; if so, claim XMR from that
	// otherwise, create new swap state from recovery info
	skB, err := inst.backend.RecoveryDB().GetCounterpartySwapPrivateKey(s.ID)
	if err == nil {
		return inst.completeSwap(s, skB)
	}

	ethSwapInfo, err := inst.backend.RecoveryDB().GetContractSwapInfo(s.ID)
	if err != nil {
		return fmt.Errorf("failed to get offer for ongoing swap, id %s: %s", s.ID, err)
	}

	sk, err := inst.backend.RecoveryDB().GetSwapPrivateKey(s.ID)
	if err != nil {
		return fmt.Errorf("failed to get private key for ongoing swap, id %s: %s", s.ID, err)
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
		inst.transferBack,
		ethSwapInfo,
		kp,
	)
	if err != nil {
		return fmt.Errorf("failed to create new swap state for ongoing swap, id %s: %s", s.ID, err)
	}

	inst.swapStates[s.ID] = ss

	go func() {
		<-ss.done
		inst.swapMu.Lock()
		defer inst.swapMu.Unlock()
		delete(inst.swapStates, s.ID)
	}()

	return nil
}

func (inst *Instance) completeSwap(s *swap.Info, skB *mcrypto.PrivateSpendKey) error {
	// fetch our swap private spend key
	skA, err := inst.backend.RecoveryDB().GetSwapPrivateKey(s.ID)
	if err != nil {
		return err
	}

	// fetch our swap private view key
	vkA, err := skA.View()
	if err != nil {
		return err
	}

	// fetch counterparty's private view key
	_, vkB, err := inst.backend.RecoveryDB().GetXMRMakerSwapKeys(s.ID)
	if err != nil {
		return err
	}

	kpAB := pcommon.GetClaimKeypair(
		skA, skB,
		vkA, vkB,
	)

	_, err = pcommon.ClaimMonero(
		inst.backend.Ctx(),
		inst.backend.Env(),
		s.ID,
		inst.backend.XMRClient(),
		s.MoneroStartHeight,
		kpAB,
		inst.backend.XMRClient().PrimaryAddress(),
		inst.transferBack,
	)
	if err != nil {
		return err
	}

	return nil
}

// Refund is called by the RPC function swap_refund.
// If it's possible to refund the ongoing swap, it does that, then notifies the counterparty.
func (inst *Instance) Refund(offerID types.Hash) (ethcommon.Hash, error) {
	inst.swapMu.Lock()
	defer inst.swapMu.Unlock()

	s, has := inst.swapStates[offerID]
	if !has {
		return ethcommon.Hash{}, errNoOngoingSwap
	}

	return s.doRefund()
}

// GetOngoingSwapState ...
func (inst *Instance) GetOngoingSwapState(offerID types.Hash) common.SwapState {
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
