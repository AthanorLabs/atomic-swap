// Package main provides the entrypoint of swapd, a daemon that manages atomic swaps
// between monero and ethereum assets.
package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/ChainSafe/chaindb"
	p2pnet "github.com/athanorlabs/go-p2p-net"
	rnet "github.com/athanorlabs/go-relayer/net"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log"
	"github.com/urfave/cli/v2"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/db"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/xmrmaker"
	"github.com/athanorlabs/atomic-swap/protocol/xmrtaker"
	"github.com/athanorlabs/atomic-swap/rpc"
)

const (
	// default libp2p ports
	defaultLibp2pPort         = 9900
	defaultXMRTakerLibp2pPort = 9933
	defaultXMRMakerLibp2pPort = 9934

	// default RPC port
	defaultRPCPort         = common.DefaultSwapdPort
	defaultXMRTakerRPCPort = defaultRPCPort
	defaultXMRMakerRPCPort = defaultXMRTakerRPCPort + 1
)

var (
	log = logging.Logger("cmd")

	// Default dev base paths. If SWAP_TEST_DATA_DIR is not defined, it is
	// still safe, there just won't be an intermediate directory and tests
	// could fail from stale data.
	testDataDir = os.Getenv("SWAP_TEST_DATA_DIR")
	// MkdirTemp uses os.TempDir() by default if the first argument is an empty string.
	defaultXMRMakerDataDir, _ = os.MkdirTemp("", path.Join(testDataDir, "xmrmaker-*"))
	defaultXMRTakerDataDir, _ = os.MkdirTemp("", path.Join(testDataDir, "xmrtaker-*"))
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
	flagContractAddress      = "contract-address"
	flagGasPrice             = "gas-price"
	flagGasLimit             = "gas-limit"
	flagUseExternalSigner    = "external-signer"

	flagDevXMRTaker      = "dev-xmrtaker"
	flagDevXMRMaker      = "dev-xmrmaker"
	flagDeploy           = "deploy"
	flagForwarderAddress = "forwarder-address"
	flagTransferBack     = "transfer-back"

	flagLogLevel = "log-level"
	flagProfile  = "profile"
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
				Usage: "Path to store swap artifacts",
				Value: "{HOME}/.atomicswap/{ENV}", // For --help only, actual default replaces variables
			},
			&cli.StringFlag{
				Name:  flagLibp2pKey,
				Usage: "libp2p private key",
				Value: fmt.Sprintf("{DATA_DIR}/%s", common.DefaultLibp2pKeyFileName),
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
				Value: fmt.Sprintf("{DATA-DIR}/wallet/%s", common.DefaultMoneroWalletName),
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
				Usage: "File containing ethereum private key as hex, new key is generated if missing",
				Value: fmt.Sprintf("{DATA-DIR}/%s", common.DefaultEthKeyFileName),
			},
			&cli.StringFlag{
				Name:  flagContractAddress,
				Usage: "Address of instance of SwapFactory.sol already deployed on-chain; required if running on mainnet",
			},
			&cli.StringSliceFlag{
				Name:    flagBootnodes,
				Aliases: []string{"bn"},
				Usage:   "libp2p bootnode, comma separated if passing multiple to a single flag",
				EnvVars: []string{"SWAPD_BOOTNODES"},
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
			&cli.StringFlag{
				Name:  flagForwarderAddress,
				Usage: "Specifies the Ethereum address of the trusted forwarder contract when deploying the swap contract. Ignored if --deploy is not passed.", //nolint:lll
			},
			&cli.BoolFlag{
				Name:  flagTransferBack,
				Usage: "Set to false to leave XMR in generated swap wallet instead of moving to primary.",
				Value: true,
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
			&cli.StringFlag{
				Name:   flagProfile,
				Usage:  "BIND_IP:PORT to provide profiling information on",
				Hidden: true, // flag is only for developers
			},
		},
	}
)

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
		os.Exit(1)
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
	ctx       context.Context
	cancel    context.CancelFunc
	database  *db.Database
	host      *net.Host
	rpcServer *rpc.Server

	// this channel is closed once the daemon has started up
	// (but before the RPC server starts, since that blocks)
	startedCh chan struct{}
}

func setLogLevelsFromContext(c *cli.Context) error {
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

	setLogLevels(level)
	return nil
}

func setLogLevels(level string) {
	_ = logging.SetLogLevel("xmrtaker", level)
	_ = logging.SetLogLevel("xmrmaker", level)
	_ = logging.SetLogLevel("coins", level)
	_ = logging.SetLogLevel("common", level)
	_ = logging.SetLogLevel("contracts", level)
	_ = logging.SetLogLevel("cmd", level)
	_ = logging.SetLogLevel("extethclient", level)
	_ = logging.SetLogLevel("monero", level)
	_ = logging.SetLogLevel("net", level)
	_ = logging.SetLogLevel("offers", level)
	_ = logging.SetLogLevel("pricefeed", level)
	_ = logging.SetLogLevel("rpc", level)

}

func runDaemon(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()
	go signalHandler(ctx, cancel)

	if err := setLogLevelsFromContext(c); err != nil {
		return err
	}

	if err := maybeStartProfiler(c); err != nil {
		return err
	}

	d := newEmptyDaemon(ctx, cancel)
	if err := d.make(c); err != nil {
		log.Errorf("RPC/Websocket server exited: %s", err)
		cancel()
	}
	if err := d.stop(); err != nil {
		log.Warnf("Cleanup error: %s", err)
	}

	return nil
}

func newEmptyDaemon(ctx context.Context, cancel context.CancelFunc) *daemon {
	return &daemon{
		ctx:       ctx,
		cancel:    cancel,
		startedCh: make(chan struct{}),
	}
}

func (d *daemon) stop() error {
	var hostErr, rpcErr, dbErr error
	d.cancel()

	if d.host != nil {
		if hostErr = d.host.Stop(); hostErr != nil {
			hostErr = fmt.Errorf("shutting down peer-to-peer services: %w", hostErr)
		}
	}

	if d.rpcServer != nil {
		if rpcErr = d.rpcServer.Stop(); rpcErr != nil {
			rpcErr = fmt.Errorf("shutting down RPC/Websockets service: %s", rpcErr)
		}
	}

	if d.database != nil {
		if dbErr = d.database.Close(); dbErr != nil {
			dbErr = fmt.Errorf("syncing database: %s", dbErr)
		}
	}

	// Making sure the database is synced is the most important task, so we don't want to
	// skip closing it if errors happen when stopping other services. We also want to
	// close services that may modify the database before closing the database. Lastly, if
	// we get multiple errors and need to chose which one to propagate upwards, a database
	// error should be prioritised first.
	switch {
	case dbErr != nil:
		return dbErr
	case rpcErr != nil:
		return rpcErr
	case hostErr != nil:
		return hostErr
	default:
		return nil
	}
}

func (d *daemon) make(c *cli.Context) error { //nolint:gocyclo
	env, err := common.NewEnv(c.String(flagEnv))
	if err != nil {
		return err
	}
	cfg := common.ConfigDefaultsForEnv(env)

	devXMRMaker := c.Bool(flagDevXMRMaker)
	devXMRTaker := c.Bool(flagDevXMRTaker)
	if devXMRMaker && devXMRTaker {
		return errFlagsMutuallyExclusive(flagDevXMRMaker, flagDevXMRTaker)
	}

	// cfg.DataDir already has a default set, so only override if the user explicitly set the flag
	if c.IsSet(flagDataDir) {
		cfg.DataDir = c.String(flagDataDir) // override the value derived from `flagEnv`
		if cfg.DataDir == "" {
			return errFlagValueEmpty(flagDataDir)
		}
	} else if env == common.Development {
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

	if c.IsSet(flagBootnodes) {
		cfg.Bootnodes = cliutil.ExpandBootnodes(c.StringSlice(flagBootnodes))
	}

	libp2pKey := cfg.LibP2PKeyFile()
	if c.IsSet(flagLibp2pKey) {
		libp2pKey = c.String(flagLibp2pKey)
		if libp2pKey == "" {
			return errFlagValueEmpty(flagLibp2pKey)
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

	ethEndpoint := common.DefaultEthEndpoint
	if c.String(flagEthereumEndpoint) != "" {
		ethEndpoint = c.String(flagEthereumEndpoint)
	}
	ec, err := ethclient.Dial(ethEndpoint)
	if err != nil {
		return err
	}
	defer ec.Close()
	chainID, err := ec.ChainID(d.ctx)
	if err != nil {
		return err
	}

	listenIP := "0.0.0.0"
	if env == common.Development {
		listenIP = "127.0.0.1"
	}

	netCfg := &p2pnet.Config{
		Ctx:        d.ctx,
		DataDir:    cfg.DataDir,
		Port:       libp2pPort,
		KeyFile:    libp2pKey,
		Bootnodes:  cfg.Bootnodes,
		ProtocolID: fmt.Sprintf("%s/%d", net.ProtocolID, chainID.Int64()),
		ListenIP:   listenIP,
	}

	host, err := net.NewHost(netCfg)
	if err != nil {
		return err
	}
	d.host = host
	relayerHost := setupRelayerNetwork(d.ctx, host.P2pHost())

	dbCfg := &chaindb.Config{
		DataDir: path.Join(cfg.DataDir, "db"),
	}

	sdb, err := db.NewDatabase(dbCfg)
	if err != nil {
		return err
	}
	d.database = sdb

	sm, err := swap.NewManager(sdb)
	if err != nil {
		return err
	}

	swapBackend, err := newBackend(
		d.ctx,
		c,
		env,
		cfg,
		devXMRMaker,
		devXMRTaker,
		sm,
		host,
		ec,
		sdb.RecoveryDB(),
		relayerHost,
	)
	if err != nil {
		return err
	}

	defer swapBackend.XMRClient().Close()
	log.Infof("created backend with monero endpoint %s and ethereum endpoint %s",
		swapBackend.XMRClient().Endpoint(),
		ethEndpoint,
	)

	a, b, err := getProtocolInstances(c, cfg, swapBackend, sdb, host)
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
		ProtocolBackend: swapBackend,
	}

	s, err := rpc.NewServer(rpcCfg)
	if err != nil {
		return err
	}
	d.rpcServer = s

	err = maybeBackgroundMine(d.ctx, devXMRMaker, swapBackend)
	if err != nil {
		return err
	}

	close(d.startedCh)

	log.Infof("starting swapd with data-dir %s", cfg.DataDir)
	err = s.Start()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func maybeBackgroundMine(ctx context.Context, devXMRMaker bool, b backend.Backend) error {
	// if we're in dev-xmrmaker mode, start background mining blocks
	// otherwise swaps won't succeed as they'll be waiting for blocks
	if !devXMRMaker {
		return nil
	}

	addr, err := b.XMRClient().GetAddress(0)
	if err != nil {
		return err
	}

	log.Infof("background mining blocks...")
	go monero.BackgroundMineBlocks(ctx, addr.Address)
	return nil
}

func errFlagsMutuallyExclusive(flag1, flag2 string) error {
	return fmt.Errorf("flags %q and %q are mutually exclusive", flag1, flag2)
}

func errFlagValueEmpty(flag string) error {
	return fmt.Errorf("flag %q requires a non-empty value", flag)
}

func errFlagValueZero(flag string) error {
	return fmt.Errorf("flag %q requires a non-zero value", flag)
}

func newBackend(
	ctx context.Context,
	c *cli.Context,
	env common.Environment,
	cfg *common.Config,
	devXMRMaker bool,
	devXMRTaker bool,
	sm swap.Manager,
	net *net.Host,
	ec *ethclient.Client,
	rdb *db.RecoveryDB,
	rhost *rnet.Host,
) (backend.Backend, error) {
	var (
		ethPrivKey *ecdsa.PrivateKey
	)

	useExternalSigner := c.Bool(flagUseExternalSigner)
	if useExternalSigner && c.IsSet(flagEthereumPrivKey) {
		return nil, errFlagsMutuallyExclusive(flagUseExternalSigner, flagEthereumPrivKey)
	}

	if !useExternalSigner {
		ethPrivKeyFile := cfg.EthKeyFileName()
		if c.IsSet(flagEthereumPrivKey) {
			ethPrivKeyFile = c.String(flagEthereumPrivKey)
			if ethPrivKeyFile == "" {
				return nil, errFlagValueEmpty(flagEthereumPrivKey)
			}
		}
		var err error
		if ethPrivKey, err = cliutil.GetEthereumPrivateKey(ethPrivKeyFile, env, devXMRMaker, devXMRTaker); err != nil {
			return nil, err
		}
	}

	extendedEC, err := extethclient.NewEthClient(ctx, env, ec, ethPrivKey)
	if err != nil {
		return nil, err
	}
	// TODO: add configs for different eth testnets + L2 and set gas limit based on those, if not set (#153)
	extendedEC.SetGasPrice(uint64(c.Uint(flagGasPrice)))
	extendedEC.SetGasLimit(uint64(c.Uint(flagGasLimit)))

	deploy := c.Bool(flagDeploy)
	if deploy {
		if c.IsSet(flagContractAddress) {
			return nil, errFlagsMutuallyExclusive(flagDeploy, flagContractAddress)
		}
		// Zero out any default contract address in the config, so we deploy
		cfg.ContractAddress = ethcommon.Address{}
	} else {
		contractAddrStr := c.String(flagContractAddress)
		if contractAddrStr != "" {
			if !ethcommon.IsHexAddress(contractAddrStr) {
				return nil, fmt.Errorf("%q is not a valid contract address", contractAddrStr)
			}
			cfg.ContractAddress = ethcommon.HexToAddress(contractAddrStr)
		}

		if bytes.Equal(cfg.ContractAddress.Bytes(), ethcommon.Address{}.Bytes()) {
			return nil, fmt.Errorf("flag %q or %q is required for env=%s", flagDeploy, flagContractAddress, env)
		}
	}

	// forwarderAddress is set only if we're deploying the swap contract, and the --forwarder-address
	// flag is set. otherwise, if we're deploying and the flag isn't set, we also deploy the forwarder.
	var forwarderAddress ethcommon.Address
	forwarderAddressStr := c.String(flagForwarderAddress)
	if deploy && forwarderAddressStr != "" {
		ok := ethcommon.IsHexAddress(forwarderAddressStr)
		if !ok {
			return nil, errors.New("forwarder-address is invalid")
		}

		forwarderAddress = ethcommon.HexToAddress(forwarderAddressStr)
	} else if !deploy && forwarderAddressStr != "" {
		log.Warnf("forwarder-address is unused")
	}

	contract, contractAddr, err := getOrDeploySwapFactory(
		ctx,
		cfg.ContractAddress,
		env,
		cfg.DataDir,
		ethPrivKey,
		ec,
		forwarderAddress,
	)
	if err != nil {
		return nil, err
	}

	if c.IsSet(flagMoneroDaemonHost) || c.IsSet(flagMoneroDaemonPort) {
		node := &common.MoneroNode{
			Host: "127.0.0.1",
			Port: common.DefaultMoneroPortFromEnv(env),
		}
		if c.IsSet(flagMoneroDaemonHost) {
			node.Host = c.String(flagMoneroDaemonHost)
			if node.Host == "" {
				return nil, errFlagValueEmpty(flagMoneroDaemonHost)
			}
		}
		if c.IsSet(flagMoneroDaemonPort) {
			node.Port = c.Uint(flagMoneroDaemonPort)
			if node.Port == 0 {
				return nil, errFlagValueZero(flagMoneroDaemonPort)
			}
		}
		cfg.MoneroNodes = []*common.MoneroNode{node}
	}

	walletFilePath := cfg.MoneroWalletPath()
	if c.IsSet(flagMoneroWalletPath) {
		walletFilePath = c.String(flagMoneroWalletPath)
		if walletFilePath == "" {
			return nil, errFlagValueEmpty(flagMoneroWalletPath)
		}
	}
	mc, err := monero.NewWalletClient(&monero.WalletClientConf{
		Env:                 env,
		WalletFilePath:      walletFilePath,
		MonerodNodes:        cfg.MoneroNodes,
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
		EthereumClient:      extendedEC,
		Environment:         env,
		SwapManager:         sm,
		SwapContract:        contract,
		SwapContractAddress: contractAddr,
		Net:                 net,
		RecoveryDB:          rdb,
		RelayerHost:         rhost,
	}

	b, err := backend.NewBackend(bcfg)
	if err != nil {
		mc.Close()
		return nil, fmt.Errorf("failed to make backend: %w", err)
	}

	return b, nil
}

func getProtocolInstances(
	c *cli.Context,
	cfg *common.Config,
	b backend.Backend,
	db *db.Database,
	host *net.Host,
) (xmrtakerHandler, xmrmakerHandler, error) {
	walletFilePath := cfg.MoneroWalletPath()
	if c.IsSet(flagMoneroWalletPath) {
		walletFilePath = c.String(flagMoneroWalletPath)
		if walletFilePath == "" {
			return nil, nil, errFlagValueEmpty(flagMoneroWalletPath)
		}
	}

	// empty password is ok
	walletPassword := c.String(flagMoneroWalletPassword)

	xmrtakerCfg := &xmrtaker.Config{
		Backend:      b,
		DataDir:      cfg.DataDir,
		TransferBack: c.Bool(flagTransferBack),
	}

	xmrTaker, err := xmrtaker.NewInstance(xmrtakerCfg)
	if err != nil {
		return nil, nil, err
	}

	xmrMakerCfg := &xmrmaker.Config{
		Backend:        b,
		DataDir:        cfg.DataDir,
		Database:       db,
		WalletFile:     walletFilePath,
		WalletPassword: walletPassword,
		Network:        host,
	}

	xmrMaker, err := xmrmaker.NewInstance(xmrMakerCfg)
	if err != nil {
		return nil, nil, err
	}

	return xmrTaker, xmrMaker, nil
}

func setupRelayerNetwork(
	ctx context.Context,
	host rnet.P2pnetHost,
) *rnet.Host {
	cfg := &rnet.Config{
		Context:   ctx,
		IsRelayer: false,
	}

	return rnet.NewHostFromP2pHost(cfg, host)
}
