package xmrmaker

import (
	"fmt"
	"sync"

	"github.com/MarinX/monerorpc/wallet"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
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

	net net.Host

	walletFile, walletPassword string

	offerManager *offers.Manager

	swapMu     sync.Mutex // synchronises access to swapStates
	swapStates map[types.Hash]*swapState
}

// Config contains the configuration values for a new XMRMaker instance.
type Config struct {
	Backend                    backend.Backend
	Database                   offers.Database
	DataDir                    string
	WalletFile, WalletPassword string
	ExternalSender             bool
	Network                    net.Host
}

// NewInstance returns a new *xmrmaker.Instance.
// It accepts an endpoint to a monero-wallet-rpc instance where account 0 contains XMRMaker's XMR.
func NewInstance(cfg *Config) (*Instance, error) {
	om, err := offers.NewManager(cfg.DataDir, cfg.Database)
	if err != nil {
		return nil, err
	}

	if om.NumOffers() > 0 {
		// this is blocking if the network service hasn't started yet
		go cfg.Network.Advertise()
	}

	inst := &Instance{
		backend:        cfg.Backend,
		dataDir:        cfg.DataDir,
		walletFile:     cfg.WalletFile,
		walletPassword: cfg.WalletPassword,
		offerManager:   om,
		swapStates:     make(map[types.Hash]*swapState),
		net:            cfg.Network,
	}

	err = inst.checkForOngoingSwaps()
	if err != nil {
		return nil, err
	}

	return inst, nil
}

func (b *Instance) checkForOngoingSwaps() error {
	swaps, err := b.backend.SwapManager().GetOngoingSwaps()
	if err != nil {
		return err
	}

	for _, s := range swaps {
		if s.Status == types.KeysExchanged || s.Status == types.ExpectingKeys {
			// TODO: set status to aborted, delete info from recovery db
			continue
		}

		err = b.createOngoingSwap(s)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Instance) createOngoingSwap(s *swap.Info) error {
	// TODO: check if we have shared secret key in db; if so, recover XMR from that
	// otherwise, create new swap state from recovery info

	offer, err := b.offerManager.GetOfferFromDB(s.ID)
	if err != nil {
		return fmt.Errorf("failed to get offer for ongoing swap, id %s: %s", s.ID, err)
	}

	ethSwapInfo, err := b.backend.RecoveryDB().GetContractSwapInfo(s.ID)
	if err != nil {
		return fmt.Errorf("failed to get offer for ongoing swap, id %s: %s", s.ID, err)
	}

	moneroStartHeight, err := b.backend.RecoveryDB().GetMoneroStartHeight(s.ID)
	if err != nil {
		return fmt.Errorf("failed to get monero start height for ongoing swap, id %s: %s", s.ID, err)
	}

	sk, err := b.backend.RecoveryDB().GetSwapPrivateKey(s.ID)
	if err != nil {
		return fmt.Errorf("failed to get private key for ongoing swap, id %s: %s", s.ID, err)
	}

	b.swapMu.Lock()
	defer b.swapMu.Unlock()
	ss, err := newSwapStateFromOngoing(
		b.backend,
		offer,
		&types.OfferExtra{}, // TODO: store relayer info in db also
		b.offerManager,
		ethSwapInfo.StartNumber,
		moneroStartHeight,
		s,
		sk,
	)
	if err != nil {
		return fmt.Errorf("failed to create new swap state for ongoing swap, id %s: %s", s.ID, err)
	}

	b.swapStates[s.ID] = ss
	return nil
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
	addr, err := b.backend.XMRClient().GetAddress(0)
	if err != nil {
		return "", nil, err
	}
	if err = b.backend.XMRClient().Refresh(); err != nil {
		return "", nil, err
	}
	balance, err := b.backend.XMRClient().GetBalance(0)
	if err != nil {
		return "", nil, err
	}
	return addr.Address, balance, nil
}
