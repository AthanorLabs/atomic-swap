package xmrtaker

import (
	"fmt"
	"sync"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	"github.com/athanorlabs/atomic-swap/protocol/txsender"

	logging "github.com/ipfs/go-log"
)

const (
	swapDepositWallet = "swap-deposit-wallet"
)

var (
	log = logging.Logger("xmrtaker")
)

// Instance implements the functionality that will be used by a user who owns ETH
// and wishes to swap for XMR.
type Instance struct {
	backend  backend.Backend
	basepath string

	walletFile, walletPassword string
	transferBack               bool // transfer xmr back to original account

	// non-nil if a swap is currently happening, nil otherwise
	// map of offer IDs -> ongoing swaps
	swapStates map[types.Hash]*swapState
	swapMu     sync.Mutex // lock for above map
}

// Config contains the configuration values for a new XMRTaker instance.
type Config struct {
	Backend                                backend.Backend
	Basepath                               string
	MoneroWalletFile, MoneroWalletPassword string
	TransferBack                           bool
	ExternalSender                         bool
}

// NewInstance returns a new instance of XMRTaker.
// It accepts an endpoint to a monero-wallet-rpc instance where XMRTaker will generate
// the account in which the XMR will be deposited.
func NewInstance(cfg *Config) (*Instance, error) {
	var (
		address mcrypto.Address
		err     error
	)

	// if this is set, it transfers all xmr received during swaps back to the given wallet.
	if cfg.TransferBack {
		address, err = getAddress(cfg.Backend, cfg.MoneroWalletFile, cfg.MoneroWalletPassword)
		if err != nil {
			return nil, err
		}
		cfg.Backend.SetBaseXMRDepositAddress(address)
	} else {
		// check that XMRTaker's monero-wallet-cli endpoint has wallet-dir configured
		err = checkWalletDir(cfg.Backend)
		if err != nil {
			return nil, err
		}
	}

	return &Instance{
		backend:        cfg.Backend,
		basepath:       cfg.Basepath,
		walletFile:     cfg.MoneroWalletFile,
		walletPassword: cfg.MoneroWalletPassword,
		swapStates:     make(map[types.Hash]*swapState),
	}, nil
}

func checkWalletDir(walletClient monero.WalletClient) error {
	// don't need to check error here, since if there's no wallet open that's fine
	_ = walletClient.CloseWallet()
	err := walletClient.CreateWallet(swapDepositWallet, "")
	if err != nil {
		return err
	}
	return walletClient.CloseWallet()
}

func getAddress(walletClient monero.WalletClient, file, password string) (mcrypto.Address, error) {
	// open XMR wallet, if it exists
	if file != "" {
		if err := walletClient.OpenWallet(file, password); err != nil {
			return "", err
		}
	} else {
		log.Info("monero wallet file not set; creating wallet swap-deposit-wallet")
		err := walletClient.CreateWallet(swapDepositWallet, "")
		if err != nil {
			if err := walletClient.OpenWallet(swapDepositWallet, ""); err != nil {
				return "", fmt.Errorf("failed to create or open swap deposit wallet: %w", err)
			}
		}
	}

	// get wallet address to deposit funds into at end of swap
	address, err := walletClient.GetAddress(0)
	if err != nil {
		return "", fmt.Errorf("failed to get monero wallet address: %w", err)
	}

	err = walletClient.CloseWallet()
	if err != nil {
		return "", fmt.Errorf("failed to close wallet: %w", err)
	}

	return mcrypto.Address(address.Address), nil
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
