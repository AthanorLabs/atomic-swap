// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package main provides the entrypoint of the swapd executable, a daemon that
// manages atomic swaps between monero and ethereum assets.
package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"os"
	"path"

	ethcommon "github.com/ethereum/go-ethereum/common"
	logging "github.com/ipfs/go-log"
	"github.com/urfave/cli/v2"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/daemon"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/monero"
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
	flagEthEndpoint          = "eth-endpoint"
	flagEthPrivKey           = "eth-privkey"
	flagContractAddress      = "contract-address"
	flagGasPrice             = "gas-price"
	flagGasLimit             = "gas-limit"
	flagUseExternalSigner    = "external-signer"
	flagRelayer              = "relayer"

	flagDevXMRTaker    = "dev-xmrtaker"
	flagDevXMRMaker    = "dev-xmrmaker"
	flagDeploy         = "deploy"
	flagNoTransferBack = "no-transfer-back"

	flagLogLevel = cliutil.FlagLogLevel
	flagProfile  = "profile"
)

func cliApp() *cli.App {
	return &cli.App{
		Name:                 "swapd",
		Usage:                "A program for doing atomic swaps between ETH and XMR",
		Version:              cliutil.GetVersion(),
		Action:               runDaemon,
		EnableBashCompletion: true,
		Suggest:              true,
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:    flagRPCPort,
				Usage:   "Port for the daemon RPC server to run on",
				Value:   defaultRPCPort,
				EnvVars: []string{"SWAPD_RPC_PORT"},
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
				Name:    flagLibp2pPort,
				Usage:   "libp2p port to listen on",
				Value:   defaultLibp2pPort,
				EnvVars: []string{"SWAPD_LIBP2P_PORT"},
			},
			&cli.StringFlag{
				Name:    flagEnv,
				Usage:   "Environment to use: one of mainnet, stagenet, or dev: Default: mainnet",
				EnvVars: []string{"SWAPD_ENV"},
				Value:   "mainnet",
			},
			&cli.StringFlag{
				Name:    flagMoneroDaemonHost,
				Usage:   "monerod host",
				EnvVars: []string{"SWAPD_MONEROD_HOST"},
			},
			&cli.UintFlag{
				Name: flagMoneroDaemonPort,
				Usage: fmt.Sprintf("monerod port (--%s=stagenet changes default to %d)",
					flagEnv, common.DefaultMoneroDaemonStagenetPort),
				EnvVars: []string{"SWAPD_MONEROD_PORT"},
				Value:   common.DefaultMoneroDaemonMainnetPort, // at least for now, this is also the dev default
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
				Name:    flagEthEndpoint,
				Usage:   "Ethereum client endpoint",
				Aliases: []string{"ethereum-endpoint"},
				EnvVars: []string{"SWAPD_ETH_ENDPOINT"},
			},
			&cli.StringFlag{
				Name:    flagEthPrivKey,
				Usage:   "File containing ethereum private key as hex, new key is generated if missing",
				Aliases: []string{"ethereum-privkey"},
				EnvVars: []string{"SWAPD_ETH_PRIVKEY"},
				Value:   fmt.Sprintf("{DATA-DIR}/%s", common.DefaultEthKeyFileName),
			},
			&cli.StringFlag{
				Name:  flagContractAddress,
				Usage: "Address of instance of SwapCreator.sol already deployed on-chain",
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
			&cli.BoolFlag{
				Name:  flagNoTransferBack,
				Usage: "Leave XMR in generated swap wallet instead of sweeping funds to primary.",
			},
			&cli.StringFlag{
				Name:    flagLogLevel,
				Usage:   "Set log level: one of [error|warn|info|debug]",
				EnvVars: []string{"SWAPD_LOG_LEVEL"},
				Value:   "info",
			},
			&cli.BoolFlag{
				Name:  flagUseExternalSigner,
				Usage: "Use external signer, for usage with the swap UI",
			},
			&cli.BoolFlag{
				Name: flagRelayer,
				Usage: fmt.Sprintf(
					"Relay claims for XMR makers and earn %s ETH (minus gas fees) per transaction",
					coins.RelayerFeeETH.Text('f'),
				),
				Value: false,
			},
			&cli.StringFlag{
				Name:   flagProfile,
				Usage:  "BIND_IP:PORT to provide profiling information on",
				Hidden: true, // flag is only for developers
			},
		},
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go cliutil.SignalHandler(ctx, cancel, log)

	err := cliApp().RunContext(ctx, os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func runDaemon(c *cli.Context) error {
	// Fail if any non-flag arguments were passed
	if c.Args().Present() {
		return fmt.Errorf("unknown command %q", c.Args().First())
	}

	if err := cliutil.SetLogLevelsFromContext(c); err != nil {
		return err
	}

	if err := maybeStartProfiler(c); err != nil {
		return err
	}

	devXMRMaker := c.Bool(flagDevXMRMaker)
	devXMRTaker := c.Bool(flagDevXMRTaker)
	if devXMRMaker && devXMRTaker {
		return errFlagsMutuallyExclusive(flagDevXMRMaker, flagDevXMRTaker)
	}

	envConf, err := getEnvConfig(c, devXMRMaker, devXMRTaker)
	if err != nil {
		return err
	}

	mc, err := createMoneroClient(c, envConf)
	if err != nil {
		return err
	}
	defer mc.Close()

	if err = maybeBackgroundMine(c.Context, devXMRMaker, mc.PrimaryAddress()); err != nil {
		return err
	}

	ec, err := createEthClient(c, envConf)
	if err != nil {
		return err
	}
	defer ec.Close()

	if err = validateOrDeployContracts(c, envConf, ec); err != nil {
		return err
	}

	conf, err := createSwapdConf(c, envConf, mc, ec)
	if err != nil {
		return err
	}

	err = daemon.RunSwapDaemon(c.Context, conf)
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

// getEnvConfig returns the environment specific config, adjusting all values changed by
// command line options.
func getEnvConfig(c *cli.Context, devXMRMaker bool, devXMRTaker bool) (*common.Config, error) {
	if c.IsSet(flagEnv) {
		if c.String(flagEnv) != common.Development.String() && (devXMRMaker || devXMRTaker) {
			return nil, errors.New("--dev-xmrmaker and --dev-xmrtaker are only valid with --env=dev")
		}
	}

	env, err := common.NewEnv(c.String(flagEnv))
	if err != nil {
		return nil, err
	}

	if devXMRMaker || devXMRTaker {
		env = common.Development
	}

	log.Infof("starting swapd, environment: %s", env)
	conf := common.ConfigDefaultsForEnv(env)

	// cfg.DataDir already has a default set, so only override if the user explicitly set the flag
	if c.IsSet(flagDataDir) {
		conf.DataDir = c.String(flagDataDir) // override the value derived from `flagEnv`
		if conf.DataDir == "" {
			return nil, errFlagValueEmpty(flagDataDir)
		}
	} else if env == common.Development {
		// Override in dev scenarios if the value was not explicitly set
		switch {
		case devXMRTaker:
			conf.DataDir = defaultXMRTakerDataDir
		case devXMRMaker:
			conf.DataDir = defaultXMRMakerDataDir
		}
	}
	if err = common.MakeDir(conf.DataDir); err != nil {
		return nil, err
	}

	if c.IsSet(flagBootnodes) {
		conf.Bootnodes = cliutil.ExpandBootnodes(c.StringSlice(flagBootnodes))
	}

	deploy := c.Bool(flagDeploy)
	if deploy {
		if c.IsSet(flagContractAddress) {
			return nil, errFlagsMutuallyExclusive(flagDeploy, flagContractAddress)
		}
		// Zero out the contract address, we'll set its final value after deploying
		conf.SwapCreatorAddr = ethcommon.Address{}
	} else {
		contractAddrStr := c.String(flagContractAddress)
		if contractAddrStr != "" {
			if !ethcommon.IsHexAddress(contractAddrStr) {
				return nil, fmt.Errorf("%q requires a valid ethereum address", flagContractAddress)
			}
			conf.SwapCreatorAddr = ethcommon.HexToAddress(contractAddrStr)
		}

		if conf.SwapCreatorAddr == (ethcommon.Address{}) {
			return nil, fmt.Errorf("flag %q or %q is required for env=%s", flagDeploy, flagContractAddress, env)
		}
	}

	return conf, nil
}

// validateOrDeployContracts validates or deploys the swap creator. The SwapCreatorAddr field
// of envConf should be all zeros if deploying and its value will be replaced by the new deployed
// contract.
func validateOrDeployContracts(c *cli.Context, envConf *common.Config, ec extethclient.EthClient) error {
	deploy := c.Bool(flagDeploy)
	if deploy && envConf.SwapCreatorAddr != (ethcommon.Address{}) {
		panic("contract address should have been zeroed when envConf was initialized")
	}

	swapCreatorAddr, err := getOrDeploySwapCreator(
		c.Context,
		envConf.SwapCreatorAddr,
		envConf.Env,
		envConf.DataDir,
		ec,
	)
	if err != nil {
		return err
	}

	envConf.SwapCreatorAddr = swapCreatorAddr
	return nil
}

func createMoneroClient(c *cli.Context, envConf *common.Config) (monero.WalletClient, error) {
	if c.IsSet(flagMoneroDaemonHost) || c.IsSet(flagMoneroDaemonPort) {
		node := &common.MoneroNode{
			Host: "127.0.0.1",
			Port: common.DefaultMoneroPortFromEnv(envConf.Env),
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
		envConf.MoneroNodes = []*common.MoneroNode{node}
	}

	walletFilePath := envConf.MoneroWalletPath()
	if c.IsSet(flagMoneroWalletPath) {
		walletFilePath = c.String(flagMoneroWalletPath)
		if walletFilePath == "" {
			return nil, errFlagValueEmpty(flagMoneroWalletPath)
		}
	}

	return monero.NewWalletClient(&monero.WalletClientConf{
		Env:                 envConf.Env,
		WalletFilePath:      walletFilePath,
		MonerodNodes:        envConf.MoneroNodes,
		MoneroWalletRPCPath: "", // look for it in "./monero-bin/monero-wallet-rpc" and then the user's path
		WalletPassword:      c.String(flagMoneroWalletPassword),
		WalletPort:          c.Uint(flagMoneroWalletPort),
	})
}

func createEthClient(c *cli.Context, envConf *common.Config) (extethclient.EthClient, error) {
	env := envConf.Env

	ethEndpoint := envConf.EthEndpoint
	if c.IsSet(flagEthEndpoint) {
		ethEndpoint = c.String(flagEthEndpoint)
	}
	if ethEndpoint == "" {
		// Message is mainnet specific, because we have defaults for dev/stagenet
		return nil, errors.New("missing ETH endpoint; completely open endpoints are not reliable for swaps")
	}

	var ethPrivKey *ecdsa.PrivateKey

	useExternalSigner := c.Bool(flagUseExternalSigner)
	if useExternalSigner && c.IsSet(flagEthPrivKey) {
		return nil, errFlagsMutuallyExclusive(flagUseExternalSigner, flagEthPrivKey)
	}

	if !useExternalSigner {
		ethPrivKeyFile := envConf.EthKeyFileName()
		if c.IsSet(flagEthPrivKey) {
			ethPrivKeyFile = c.String(flagEthPrivKey)
			if ethPrivKeyFile == "" {
				return nil, errFlagValueEmpty(flagEthPrivKey)
			}
		}

		devXMRMaker := c.Bool(flagDevXMRMaker)
		devXMRTaker := c.Bool(flagDevXMRTaker)
		if devXMRMaker && devXMRTaker {
			return nil, errFlagsMutuallyExclusive(flagDevXMRMaker, flagDevXMRTaker)
		}

		var err error
		ethPrivKey, err = cliutil.GetEthereumPrivateKey(ethPrivKeyFile, env, devXMRMaker, devXMRTaker)
		if err != nil {
			return nil, err
		}
	}

	extendedEC, err := extethclient.NewEthClient(c.Context, env, ethEndpoint, ethPrivKey)
	if err != nil {
		return nil, err
	}

	// TODO: add configs for different eth testnets + L2 and set gas limit based on those, if not set (#153)
	extendedEC.SetGasPrice(uint64(c.Uint(flagGasPrice)))
	extendedEC.SetGasLimit(uint64(c.Uint(flagGasLimit)))

	return extendedEC, nil
}

func createSwapdConf(
	c *cli.Context,
	envConf *common.Config,
	mc monero.WalletClient,
	ec extethclient.EthClient,
) (*daemon.SwapdConfig, error) {

	libp2pKeyFile := envConf.LibP2PKeyFile()
	if c.IsSet(flagLibp2pKey) {
		libp2pKeyFile = c.String(flagLibp2pKey)
		if libp2pKeyFile == "" {
			return nil, errFlagValueEmpty(flagLibp2pKey)
		}
	}

	libp2pPort := c.Uint(flagLibp2pPort)
	if !c.IsSet(flagLibp2pPort) {
		switch {
		case c.Bool(flagDevXMRMaker):
			libp2pPort = defaultXMRMakerLibp2pPort
		case c.Bool(flagDevXMRTaker):
			libp2pPort = defaultXMRTakerLibp2pPort
		}
	}

	rpcPort := c.Uint(flagRPCPort)
	if !c.IsSet(flagRPCPort) {
		switch {
		case c.Bool(flagDevXMRMaker):
			rpcPort = defaultXMRMakerRPCPort
		case c.Bool(flagDevXMRTaker):
			rpcPort = defaultXMRTakerRPCPort
		}
	}

	return &daemon.SwapdConfig{
		EnvConf:        envConf,
		Libp2pPort:     uint16(libp2pPort),
		Libp2pKeyfile:  libp2pKeyFile,
		RPCPort:        uint16(rpcPort),
		IsRelayer:      c.Bool(flagRelayer),
		NoTransferBack: c.Bool(flagNoTransferBack),
		MoneroClient:   mc,
		EthereumClient: ec,
	}, nil
}

func maybeBackgroundMine(ctx context.Context, devXMRMaker bool, address *mcrypto.Address) error {
	// if we're in dev-xmrmaker mode, start background mining blocks
	// otherwise swaps won't succeed as they'll be waiting for blocks
	if !devXMRMaker {
		return nil
	}

	log.Infof("background mining blocks...")
	go monero.BackgroundMineBlocks(ctx, address)
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
