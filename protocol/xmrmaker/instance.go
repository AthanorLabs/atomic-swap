package xmrmaker

import (
	"fmt"
	"sync"

	"github.com/MarinX/monerorpc/wallet"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/xmrmaker/offers"

	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("xmrmaker")
)

// Host contains required network functionality.
type Host interface {
	Advertise([]string)
}

// Instance implements the functionality that will be needed by a user who owns XMR
// and wishes to swap for ETH.
type Instance struct {
	backend backend.Backend
	dataDir string

	net Host

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
	Network                    Host
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
		go cfg.Network.Advertise([]string{string(coins.ProvidesXMR)})
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

func (inst *Instance) checkForOngoingSwaps() error {
	swaps, err := inst.backend.SwapManager().GetOngoingSwaps()
	if err != nil {
		return err
	}

	for _, s := range swaps {
		if s.Provides != coins.ProvidesXMR {
			continue
		}

		if s.Status == types.KeysExchanged || s.Status == types.ExpectingKeys {
			// for these two cases, no funds have been locked, so we can safely
			// abort the swap.
			err = inst.abortOngoingSwap(s)
			if err != nil {
				return fmt.Errorf("failed to abort ongoing swap: %w", err)
			}

			continue
		}

		err = inst.createOngoingSwap(s)
		if err != nil {
			return err
		}
	}

	return nil
}

func (inst *Instance) abortOngoingSwap(s swap.Info) error {
	// set status to aborted, delete info from recovery db
	s.Status = types.CompletedAbort
	err := inst.backend.SwapManager().CompleteOngoingSwap(&s)
	if err != nil {
		return err
	}

	return inst.backend.RecoveryDB().DeleteSwap(s.ID)
}

func (inst *Instance) createOngoingSwap(s swap.Info) error {
	// check if we have shared secret key in db; if so, recover XMR from that
	// otherwise, create new swap state from recovery info
	sharedKey, err := inst.backend.RecoveryDB().GetSharedSwapPrivateKey(s.ID)
	if err == nil {
		kp, err := sharedKey.AsPrivateKeyPair() //nolint:govet
		if err != nil {
			return err
		}

		inst.backend.XMRClient().Lock()
		defer inst.backend.XMRClient().Unlock()

		// TODO: do we want to transfer this back to the original account?
		addr, err := monero.CreateWallet(
			"xmrmaker-swap-wallet",
			inst.backend.Env(),
			inst.backend.XMRClient(),
			kp,
			s.MoneroStartHeight,
		)
		if err != nil {
			return err
		}

		log.Infof("refunded XMR from swap %s: wallet addr is %s", s.ID, addr)
		return nil
	}

	offer, err := inst.offerManager.GetOfferFromDB(s.ID)
	if err != nil {
		return fmt.Errorf("failed to get offer for ongoing swap, id %s: %s", s.ID, err)
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

	relayerInfo, err := inst.backend.RecoveryDB().GetSwapRelayerInfo(s.ID)
	if err != nil {
		// we can ignore the error; if the key doesn't exist,
		// then no relayer was set for this swap.
		relayerInfo = &types.OfferExtra{}
	}

	ss, err := newSwapStateFromOngoing(
		inst.backend,
		offer,
		relayerInfo,
		inst.offerManager,
		ethSwapInfo,
		&s,
		kp,
	)
	if err != nil {
		return fmt.Errorf("failed to create new swap state for ongoing swap, id %s: %s", s.ID, err)
	}

	inst.swapMu.Lock()
	inst.swapStates[s.ID] = ss
	inst.swapMu.Unlock()

	go func() {
		<-ss.done
		inst.swapMu.Lock()
		defer inst.swapMu.Unlock()
		delete(inst.swapStates, offer.ID)
	}()

	return nil
}

// GetOngoingSwapState ...
func (inst *Instance) GetOngoingSwapState(id types.Hash) common.SwapState {
	inst.swapMu.Lock()
	defer inst.swapMu.Unlock()

	return inst.swapStates[id]
}

// GetMoneroBalance returns the primary wallet address, and current balance of the user's monero
// wallet.
func (inst *Instance) GetMoneroBalance() (string, *wallet.GetBalanceResponse, error) {
	addr, err := inst.backend.XMRClient().GetAddress(0)
	if err != nil {
		return "", nil, err
	}
	if err = inst.backend.XMRClient().Refresh(); err != nil {
		return "", nil, err
	}
	balance, err := inst.backend.XMRClient().GetBalance(0)
	if err != nil {
		return "", nil, err
	}
	return addr.Address, balance, nil
}
