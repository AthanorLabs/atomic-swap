package xmrmaker

import (
	"sync"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/protocol/backend"

	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("xmrmaker")
)

// Instance implements the functionality that will be needed by a user who owns XMR
// and wishes to swap for ETH.
type Instance struct {
	backend backend.Backend
	// ctx      context.Context
	// env      common.Environment
	basepath string

	// client                     monero.Client
	// daemonClient               monero.DaemonClient
	walletFile, walletPassword string

	// ethClient  *ethclient.Client
	// ethPrivKey *ecdsa.PrivateKey
	// callOpts   *bind.CallOpts
	// ethAddress ethcommon.Address
	// chainID    *big.Int
	// gasPrice   *big.Int
	// gasLimit   uint64

	// net net.MessageSender

	offerManager *offerManager
	// swapManager  *swap.Manager

	swapMu    sync.Mutex
	swapState *swapState
}

// Config contains the configuration values for a new XMRMaker instance.
type Config struct {
	Backend backend.Backend
	//Ctx                        context.Context
	Basepath string
	// MoneroWalletEndpoint       string
	// MoneroDaemonEndpoint       string // only needed for development
	WalletFile, WalletPassword string
	// EthereumClient             *ethclient.Client
	// EthereumPrivateKey         *ecdsa.PrivateKey
	// Environment                common.Environment
	// ChainID                    *big.Int
	// GasPrice                   *big.Int
	// SwapManager                *swap.Manager
	// GasLimit                   uint64
}

// NewInstance returns a new *xmrmaker.Instance.
// It accepts an endpoint to a monero-wallet-rpc instance where account 0 contains XMRMaker's XMR.
func NewInstance(cfg *Config) (*Instance, error) {
	// if cfg.Environment == common.Development && cfg.MoneroDaemonEndpoint == "" {
	// 	return nil, errMustProvideDaemonEndpoint
	// }

	// addr := common.EthereumPrivateKeyToAddress(cfg.EthereumPrivateKey)

	// // monero-wallet-rpc client
	// walletClient := monero.NewClient(cfg.MoneroWalletEndpoint)

	// open XMRMaker's XMR wallet
	if cfg.WalletFile != "" {
		if err := cfg.Backend.OpenWallet(cfg.WalletFile, cfg.WalletPassword); err != nil {
			return nil, err
		}
	} else {
		log.Warn("monero wallet-file not set; must be set via RPC call personal_setMoneroWalletFile before making an offer")
	}

	// // this is only used in the monero development environment to generate new blocks
	// var daemonClient monero.DaemonClient
	// if cfg.Environment == common.Development {
	// 	daemonClient = monero.NewClient(cfg.MoneroDaemonEndpoint)
	// }

	return &Instance{
		backend: cfg.Backend,
		//ctx:            cfg.Ctx,
		basepath: cfg.Basepath,
		// env:            cfg.Environment,
		// client:         walletClient,
		// daemonClient:   daemonClient,
		walletFile:     cfg.WalletFile,
		walletPassword: cfg.WalletPassword,
		// ethClient:      cfg.EthereumClient,
		// ethPrivKey:     cfg.EthereumPrivateKey,
		// callOpts: &bind.CallOpts{
		// 	From:    addr,
		// 	Context: cfg.Ctx,
		// },
		// ethAddress:   addr,
		// chainID:      cfg.ChainID,
		offerManager: newOfferManager(cfg.Basepath),
		//swapManager:  cfg.SwapManager,
	}, nil
}

// // SetMessageSender sets the Instance's net.MessageSender interface.
// func (b *Instance) SetMessageSender(n net.MessageSender) {
// 	b.net = n
// }

// SetMoneroWalletFile sets the Instance's current monero wallet file.
func (b *Instance) SetMoneroWalletFile(file, password string) error {
	_ = b.backend.CloseWallet()
	return b.backend.OpenWallet(file, password)
}

func (b *Instance) openWallet() error { //nolint
	return b.backend.OpenWallet(b.walletFile, b.walletPassword)
}

// GetOngoingSwapState ...
func (b *Instance) GetOngoingSwapState() common.SwapState {
	return b.swapState
}
