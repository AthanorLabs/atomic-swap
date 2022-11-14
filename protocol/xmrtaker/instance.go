package xmrtaker

import (
	"sync"

	ethcommon "github.com/ethereum/go-ethereum/common"
	logging "github.com/ipfs/go-log"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
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
		cfg.Backend.SetBaseXMRDepositAddress(cfg.Backend.XMRClient().PrimaryWalletAddress())
	}

	return &Instance{
		backend:    cfg.Backend,
		dataDir:    cfg.DataDir,
		swapStates: make(map[types.Hash]*swapState),
	}, nil
}

// Refund is called by the RPC function swap_refund.
// If it's possible to refund the ongoing swap, it does that, then notifies the counterparty.
func (a *Instance) Refund(offerID types.Hash) (ethcommon.Hash, error) {
	a.swapMu.Lock()
	defer a.swapMu.Unlock()

	s, has := a.swapStates[offerID]
	if !has {
		return ethcommon.Hash{}, errNoOngoingSwap
	}

	return s.doRefund()
}

// GetOngoingSwapState ...
func (a *Instance) GetOngoingSwapState(offerID types.Hash) common.SwapState {
	return a.swapStates[offerID]
}

// ExternalSender returns the *txsender.ExternalSender for a swap, if the swap exists and is using
// and external tx sender
func (a *Instance) ExternalSender(offerID types.Hash) (*txsender.ExternalSender, error) {
	a.swapMu.Lock()
	defer a.swapMu.Unlock()

	s, has := a.swapStates[offerID]
	if !has {
		return nil, errNoOngoingSwap
	}

	es, ok := s.sender.(*txsender.ExternalSender)
	if !ok {
		return nil, errSenderIsNotExternal
	}

	return es, nil
}
