package xmrmaker

import (
	"sync"

	"github.com/MarinX/monerorpc/wallet"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/db"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	"github.com/athanorlabs/atomic-swap/protocol/xmrmaker/offers"

	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("xmrmaker")
)

// Instance implements the functionality that will be needed by a user who owns XMR
// and wishes to swap for ETH.
type Instance struct {
	backend backend.Backend
	dataDir string

	walletFile, walletPassword string

	offerManager *offers.Manager

	swapMu     sync.Mutex // synchronises access to swapStates
	swapStates map[types.Hash]*swapState
}

// Config contains the configuration values for a new XMRMaker instance.
type Config struct {
	Backend                    backend.Backend
	Database                   *db.Database
	DataDir                    string
	WalletFile, WalletPassword string
	ExternalSender             bool
}

// NewInstance returns a new *xmrmaker.Instance.
// It accepts an endpoint to a monero-wallet-rpc instance where account 0 contains XMRMaker's XMR.
func NewInstance(cfg *Config) (*Instance, error) {
	om, err := offers.NewManager(cfg.DataDir, cfg.Database)
	if err != nil {
		return nil, err
	}

	return &Instance{
		backend:        cfg.Backend,
		dataDir:        cfg.DataDir,
		walletFile:     cfg.WalletFile,
		walletPassword: cfg.WalletPassword,
		offerManager:   om,
		swapStates:     make(map[types.Hash]*swapState),
	}, nil
}

// GetOngoingSwapState ...
func (b *Instance) GetOngoingSwapState(id types.Hash) common.SwapState {
	b.swapMu.Lock()
	defer b.swapMu.Unlock()

	return b.swapStates[id]
}

// GetMoneroBalance returns the primary wallet address, and current balance of the user's monero
// wallet.
func (b *Instance) GetMoneroBalance() (string, *wallet.GetBalanceResponse, error) {
	addr, err := b.backend.GetAddress(0)
	if err != nil {
		return "", nil, err
	}
	if err = b.backend.Refresh(); err != nil {
		return "", nil, err
	}
	balance, err := b.backend.GetBalance(0)
	if err != nil {
		return "", nil, err
	}
	return addr.Address, balance, nil
}
