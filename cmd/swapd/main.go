package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"path"
	"strings"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/xmrmaker"
	"github.com/athanorlabs/atomic-swap/protocol/xmrtaker"
	"github.com/athanorlabs/atomic-swap/rpc"

	logging "github.com/ipfs/go-log"
)

const (
	// default libp2p ports
	defaultLibp2pPort         = 9900
	defaultXMRTakerLibp2pPort = 9933
	defaultXMRMakerLibp2pPort = 9934

	// default libp2p key files
	defaultLibp2pKey         = "node.key"
	defaultXMRTakerLibp2pKey = "xmrtaker.key"
	defaultXMRMakerLibp2pKey = "xmrmaker.key"

	// default RPC port
	defaultRPCPort         = 5005
	defaultXMRTakerRPCPort = 5001
	defaultXMRMakerRPCPort = 5002
)

var (
	log = logging.Logger("cmd")

	// Default dev base paths. If SWAP_TEST_DATA_DIR is not defined, it is
	// still safe, there just won't be an intermediate directory and tests
	// could fail from stale data.
	testDataDir            = os.Getenv("SWAP_TEST_DATA_DIR")
	defaultXMRMakerDataDir = path.Join(os.TempDir(), testDataDir, "xmrmaker")
	defaultXMRTakerDataDir = path.Join(os.TempDir(), testDataDir, "xmrtaker")
)

const (
	flagRPCPort    = "rpc-port"
	flagDataDir    = "data-dir"
	flagLibp2pKey  = "libp2p-key"
	flagLibp2pPort = "libp2p-port"
	flagBootnodes  = "bootnodes"

	flagEnv                  = "env"
	flagMoneroDaemonHost     = "monerod-host"
	flagMoneroDaemonPort     = "monerod-port"
	flagMoneroWalletPath     = "wallet-file"
	flagMoneroWalletPassword = "wallet-password"
	flagMoneroWalletPort     = "wallet-port"
	flagEthereumEndpoint     = "ethereum-endpoint"
	flagEthereumPrivKey      = "ethereum-privkey"
	flagEthereumChainID      = "ethereum-chain-id"
	flagContractAddress      = "contract-address"
	flagGasPrice             = "gas-price"
	flagGasLimit             = "gas-limit"
	flagUseExternalSigner    = "external-signer"

	flagDevXMRTaker  = "dev-xmrtaker"
	flagDevXMRMaker  = "dev-xmrmaker"
	flagDeploy       = "deploy"
	flagTransferBack = "transfer-back"

	flagLogLevel = "log-level"
)

var (
	app = &cli.App{
		Name:                 "swapd",
		Usage:                "A program for doing atomic swaps between ETH and XMR",
		Version:              cliutil.GetVersion(),
		Action:               runDaemon,
		EnableBashCompletion: true,
		Suggest:              true,
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:  flagRPCPort,
				Usage: "Port for the daemon RPC server to run on",
				Value: defaultRPCPort,
			},
			&cli.StringFlag{
				Name:  flagDataDir,
				Usage: "Path to store swap artifacts", //nolint:misspell
				Value: "{HOME}/.atomicswap/{ENV}",     // For --help only, actual default replaces variables
			},
			&cli.StringFlag{
				Name:  flagLibp2pKey,
				Usage: "libp2p private key",
				Value: defaultLibp2pKey,
			},
			&cli.UintFlag{
				Name:  flagLibp2pPort,
				Usage: "libp2p port to listen on",
				Value: defaultLibp2pPort,
			},
			&cli.StringFlag{
				Name:  flagEnv,
				Usage: "Environment to use: one of mainnet, stagenet, or dev",
				Value: "dev",
			},
			&cli.StringFlag{
				Name:  flagMoneroDaemonHost,
				Usage: "monerod host",
				Value: "127.0.0.1",
			},
			&cli.UintFlag{
				Name: flagMoneroDaemonPort,
				Usage: fmt.Sprintf("monerod port (--%s=stagenet changes default to %d)",
					flagEnv, common.DefaultMoneroDaemonStagenetPort),
				Value: common.DefaultMoneroDaemonMainnetPort, // at least for now, this is also the dev default
			},
			&cli.StringFlag{
				Name:  flagMoneroWalletPath,
				Usage: "Path to the Monero wallet file, created if missing",
				Value: "{DATA-DIR}/{ENV}/wallet/swap-wallet",
			},
			&cli.StringFlag{
				Name:  flagMoneroWalletPassword,
				Usage: "Password of monero wallet file",
			},
			&cli.UintFlag{
				Name:   flagMoneroWalletPort,
				Usage:  "The port that the internal monero-wallet-rpc instance listens on",
				Hidden: true, // flag is for integration tests and won't be supported long term
			},
			&cli.StringFlag{
				Name:  flagEthereumEndpoint,
				Usage: "Ethereum client endpoint",
			},
			&cli.StringFlag{
				Name:  flagEthereumPrivKey,
				Usage: "File containing a private key as hex string",
			},
			&cli.UintFlag{
				Name:  flagEthereumChainID,
				Usage: "Ethereum chain ID; eg. mainnet=1, goerli=5, ganache=1337",
			},
			&cli.StringFlag{
				Name:  flagContractAddress,
				Usage: "Address of instance of SwapFactory.sol already deployed on-chain; required if running on mainnet",
			},
			&cli.StringSliceFlag{
				Name:    flagBootnodes,
				Aliases: []string{"bn"},
				Usage:   "libp2p bootnode, comma separated if passing multiple to a single flag",
			},
			&cli.UintFlag{
				Name:  flagGasPrice,
				Usage: "Ethereum gas price to use for transactions (in gwei). If not set, the gas price is set via oracle.",
			},
			&cli.UintFlag{
				Name:  flagGasLimit,
				Usage: "Ethereum gas limit to use for transactions. If not set, the gas limit is estimated for each transaction.",
			},
			&cli.BoolFlag{
				Name:  flagDevXMRTaker,
				Usage: "Run in development mode and use ETH provider default values",
			},
			&cli.BoolFlag{
				Name:  flagDevXMRMaker,
				Usage: "Run in development mode and use XMR provider default values",
			},
			&cli.BoolFlag{
				Name:  flagDeploy,
				Usage: "Deploy an instance of the swap contract",
			},
			&cli.BoolFlag{
				Name:  flagTransferBack,
				Usage: "When receiving XMR in a swap, transfer it back to the original wallet.",
			},
			&cli.StringFlag{
				Name:  flagLogLevel,
				Usage: "Set log level: one of [error|warn|info|debug]",
				Value: "info",
			},
			&cli.BoolFlag{
				Name:  flagUseExternalSigner,
				Usage: "Use external signer, for usage with the swap UI",
			},
		},
	}
)

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

type xmrtakerHandler interface {
	rpc.XMRTaker
}

type xmrmakerHandler interface {
	net.Handler
	rpc.XMRMaker
}

type daemon struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func setLogLevels(c *cli.Context) error {
	const (
		levelError = "error"
		levelWarn  = "warn"
		levelInfo  = "info"
		levelDebug = "debug"
	)

	level := c.String(flagLogLevel)
	switch level {
	case levelError, levelWarn, levelInfo, levelDebug:
	default:
		return fmt.Errorf("invalid log level %q", level)
	}

	_ = logging.SetLogLevel("xmrtaker", level)
	_ = logging.SetLogLevel("xmrmaker", level)
	_ = logging.SetLogLevel("common", level)
	_ = logging.SetLogLevel("cmd", level)
	_ = logging.SetLogLevel("net", level)
	_ = logging.SetLogLevel("rpc", level)
	_ = logging.SetLogLevel("monero", level)
	return nil
}

func runDaemon(c *cli.Context) error {
	if err := setLogLevels(c); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := &daemon{
		ctx:    ctx,
		cancel: cancel,
	}

	err := d.make(c)
	if err != nil {
		return err
	}

	d.wait()
	os.Exit(0)
	return nil
}

// expandBootnodes expands the boot nodes passed on the command line that
// can be specified individually with multiple flags, but can also contain
// multiple boot nodes passed to single flag separated by commas.
func expandBootnodes(nodesCLI []string) []string {
	var nodes []string
	for _, n := range nodesCLI {
		splitNodes := strings.Split(n, ",")
		for _, ns := range splitNodes {
			nodes = append(nodes, strings.TrimSpace(ns))
		}
	}
	return nodes
}

func (d *daemon) make(c *cli.Context) error {
	env, cfg, err := cliutil.GetEnvironment(c.String(flagEnv))
	if err != nil {
		return err
	}

	devXMRMaker := c.Bool(flagDevXMRMaker)
	devXMRTaker := c.Bool(flagDevXMRTaker)
	if devXMRMaker && devXMRTaker {
		return errFlagsMutuallyExclusive(flagDevXMRMaker, flagDevXMRTaker)
	}

	// By default, the chain ID is derived from the `flagEnv` value, but it can be overridden if
	// `flagEthereumChainID` is passed:
	if c.IsSet(flagEthereumChainID) {
		cfg.EthereumChainID = int64(c.Uint(flagEthereumChainID))
	}

	if len(c.StringSlice(flagBootnodes)) > 0 {
		cfg.Bootnodes = expandBootnodes(c.StringSlice(flagBootnodes))
	}

	//
	// Note: Overrides for devXMRTaker/devXMRMaker use "IsSet" instead of checking the value so that
	//       the devXMRTaker/devXMRMaker configurations take precedence over normal default values,
	//       but will not override values explicitly set by the end user.
	//

	libp2pKey := c.String(flagLibp2pKey)
	if !c.IsSet(flagLibp2pKey) {
		switch {
		case devXMRTaker:
			libp2pKey = defaultXMRTakerLibp2pKey
		case devXMRMaker:
			libp2pKey = defaultXMRMakerLibp2pKey
		}
	}

	libp2pPort := uint16(c.Uint(flagLibp2pPort))
	if !c.IsSet(flagLibp2pPort) {
		switch {
		case devXMRTaker:
			libp2pPort = defaultXMRTakerLibp2pPort
		case devXMRMaker:
			libp2pPort = defaultXMRMakerLibp2pPort
		}
	}

	// cfg.DataDir was already defaulted from the `flagEnv` value and `flagDataDir` does
	// not directly set a default value.
	if c.IsSet(flagDataDir) {
		cfg.DataDir = c.String(flagDataDir) // override the value derived from `flagEnv`
	} else {
		// Override in dev scenarios if the value was not explicitly set
		switch {
		case devXMRTaker:
			cfg.DataDir = defaultXMRTakerDataDir
		case devXMRMaker:
			cfg.DataDir = defaultXMRMakerDataDir
		}
	}

	if err = common.MakeDir(cfg.DataDir); err != nil {
		return err
	}

	netCfg := &net.Config{
		Ctx:         d.ctx,
		Environment: env,
		DataDir:     cfg.DataDir,
		EthChainID:  cfg.EthereumChainID,
		Port:        libp2pPort,
		KeyFile:     libp2pKey,
		Bootnodes:   cfg.Bootnodes,
	}
	host, err := net.NewHost(netCfg)
	if err != nil {
		return err
	}

	sm := swap.NewManager()
	backend, err := newBackend(d.ctx, c, env, cfg, devXMRMaker, devXMRTaker, sm, host)
	if err != nil {
		return err
	}
	defer backend.Close()

	a, b, err := getProtocolInstances(c, cfg, backend)
	if err != nil {
		return err
	}

	// connect network to protocol handler
	// handler handles initiated ("taken") swap
	host.SetHandler(b)

	if err = host.Start(); err != nil {
		return err
	}

	rpcPort := uint16(c.Uint(flagRPCPort))
	if !c.IsSet(flagRPCPort) {
		switch {
		case devXMRTaker:
			rpcPort = defaultXMRTakerRPCPort
		case devXMRMaker:
			rpcPort = defaultXMRMakerRPCPort
		}
	}
	listenAddr := fmt.Sprintf("127.0.0.1:%d", rpcPort)

	rpcCfg := &rpc.Config{
		Ctx:             d.ctx,
		Address:         listenAddr,
		Net:             host,
		XMRTaker:        a,
		XMRMaker:        b,
		ProtocolBackend: backend,
	}

	s, err := rpc.NewServer(rpcCfg)
	if err != nil {
		return err
	}

	log.Infof("starting swapd with data-dir %s", cfg.DataDir)
	return s.Start()
}

func errFlagsMutuallyExclusive(flag1, flag2 string) error {
	return fmt.Errorf("flags %q and %q are mutually exclusive", flag1, flag2)
}

func newBackend(
	ctx context.Context,
	c *cli.Context,
	env common.Environment,
	cfg common.Config,
	devXMRMaker bool,
	devXMRTaker bool,
	sm swap.Manager,
	net net.Host,
) (backend.Backend, error) {
	var (
		ethEndpoint string
		ethPrivKey  *ecdsa.PrivateKey
	)

	if c.String(flagEthereumEndpoint) != "" {
		ethEndpoint = c.String(flagEthereumEndpoint)
	} else {
		ethEndpoint = common.DefaultEthEndpoint
	}

	useExternalSigner := c.Bool(flagUseExternalSigner)
	ethPrivKeyFile := c.String(flagEthereumPrivKey)
	if useExternalSigner && ethPrivKeyFile != "" {
		return nil, errFlagsMutuallyExclusive(flagUseExternalSigner, flagEthereumPrivKey)
	}

	if !useExternalSigner {
		var err error
		if ethPrivKey, err = cliutil.GetEthereumPrivateKey(ethPrivKeyFile, env, devXMRMaker, devXMRTaker); err != nil {
			return nil, err
		}
	}

	// TODO: add configs for different eth testnets + L2 and set gas limit based on those, if not set (#153)
	var gasPrice *big.Int
	if c.Uint(flagGasPrice) != 0 {
		gasPrice = big.NewInt(int64(c.Uint(flagGasPrice)))
	}

	contractAddrStr := c.String(flagContractAddress)
	if contractAddrStr != "" {
		// We check the contract code at the address later, so we don't need
		// to tightly validate the address here.
		cfg.ContractAddress = ethcommon.HexToAddress(contractAddrStr)
	}

	ec, err := ethclient.Dial(ethEndpoint)
	if err != nil {
		return nil, err
	}

	deploy := c.Bool(flagDeploy)
	if deploy {
		if c.IsSet(flagContractAddress) {
			return nil, errFlagsMutuallyExclusive(flagDeploy, flagContractAddress)
		}
		// Zero out any default contract address in the config, so we deploy
		cfg.ContractAddress = ethcommon.Address{}
	}

	chainID := big.NewInt(cfg.EthereumChainID)
	contract, contractAddr, err :=
		getOrDeploySwapFactory(ctx, cfg.ContractAddress, env, cfg.DataDir, chainID, ethPrivKey, ec)
	if err != nil {
		return nil, err
	}

	// For the monero wallet related values, keep the default config values unless the end
	// use explicitly set the flag.
	if c.IsSet(flagMoneroDaemonHost) {
		cfg.MoneroDaemonHost = c.String(flagMoneroDaemonHost)
	}
	if c.IsSet(flagMoneroDaemonPort) {
		cfg.MoneroDaemonPort = c.Uint(flagMoneroDaemonPort)
	}
	walletFilePath := cfg.MoneroWalletPath()
	if c.IsSet(flagMoneroWalletPath) {
		walletFilePath = c.String(flagMoneroWalletPath)
	}
	mc, err := monero.NewWalletClient(&monero.WalletClientConf{
		Env:                 env,
		WalletFilePath:      walletFilePath,
		MonerodPort:         cfg.MoneroDaemonPort,
		MonerodHost:         cfg.MoneroDaemonHost,
		MoneroWalletRPCPath: "", // look for it in "monero-bin/monero-wallet-rpc" and then the user's path
		WalletPassword:      c.String(flagMoneroWalletPassword),
		WalletPort:          c.Uint(flagMoneroWalletPort),
	})
	if err != nil {
		return nil, err
	}

	bcfg := &backend.Config{
		Ctx:                 ctx,
		MoneroClient:        mc,
		EthereumClient:      ec,
		EthereumPrivateKey:  ethPrivKey,
		Environment:         env,
		ChainID:             chainID,
		GasPrice:            gasPrice,
		GasLimit:            uint64(c.Uint(flagGasLimit)),
		SwapManager:         sm,
		SwapContract:        contract,
		SwapContractAddress: contractAddr,
		Net:                 net,
	}

	b, err := backend.NewBackend(bcfg)
	if err != nil {
		mc.Close()
		return nil, fmt.Errorf("failed to make backend: %w", err)
	}

	log.Infof("created backend with monero endpoint %s and ethereum endpoint %s",
		mc.Endpoint(),
		ethEndpoint,
	)

	return b, nil
}

func getProtocolInstances(c *cli.Context, cfg common.Config,
	b backend.Backend) (xmrtakerHandler, xmrmakerHandler, error) {
	walletFile := c.String("wallet-file")

	// empty password is ok
	walletPassword := c.String("wallet-password")

	xmrtakerCfg := &xmrtaker.Config{
		Backend:              b,
		DataDir:              cfg.DataDir,
		MoneroWalletFile:     walletFile,
		MoneroWalletPassword: walletPassword,
		TransferBack:         c.Bool(flagTransferBack),
	}

	xmrtaker, err := xmrtaker.NewInstance(xmrtakerCfg)
	if err != nil {
		return nil, nil, err
	}

	xmrmakerCfg := &xmrmaker.Config{
		Backend:        b,
		DataDir:        cfg.DataDir,
		WalletFile:     walletFile,
		WalletPassword: walletPassword,
	}

	xmrmaker, err := xmrmaker.NewInstance(xmrmakerCfg)
	if err != nil {
		return nil, nil, err
	}

	return xmrtaker, xmrmaker, nil
}
