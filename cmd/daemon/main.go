package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"

	"github.com/noot/atomic-swap/alice"
	"github.com/noot/atomic-swap/bob"
	"github.com/noot/atomic-swap/common"
	recovery "github.com/noot/atomic-swap/monero/recover"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/rpc"

	logging "github.com/ipfs/go-log"
)

const (
	// default libp2p ports
	defaultAliceLibp2pPort = 9933
	defaultBobLibp2pPort   = 9934

	// defaultExchangeRate is the default ratio of ETH:XMR.
	// TODO; make this a CLI flag, or get it from some price feed.
	defaultExchangeRate = 0.0578261

	// default libp2p key files
	defaultAliceLibp2pKey = "alice.key"
	defaultBobLibp2pKey   = "bob.key"

	// default RPC port
	defaultAliceRPCPort = 5001
	defaultBobRPCPort   = 5002

	defaultEnvironment = common.Development
)

var (
	log = logging.Logger("cmd")
	_   = logging.SetLogLevel("alice", "debug")
	_   = logging.SetLogLevel("bob", "debug")
	_   = logging.SetLogLevel("common", "debug")
	_   = logging.SetLogLevel("cmd", "debug")
	_   = logging.SetLogLevel("net", "debug")
	_   = logging.SetLogLevel("rpc", "debug")
)

var (
	app = &cli.App{
		Name:   "atomic-swap",
		Usage:  "A program for doing atomic swaps between ETH and XMR",
		Action: runDaemon,
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:  "rpc-port",
				Usage: "port for the daemon RPC server to run on; default 5001",
			},
			&cli.StringFlag{
				Name:  "basepath",
				Usage: "path to store swap artefacts",
			},
			&cli.StringFlag{
				Name:  "libp2p-key",
				Usage: "libp2p private key",
			},
			&cli.UintFlag{
				Name:  "libp2p-port",
				Usage: "libp2p port to listen on",
			},
			&cli.BoolFlag{
				Name:  "alice",
				Usage: "run as Alice (have ETH, want XMR)",
			},
			&cli.BoolFlag{
				Name:  "bob",
				Usage: "run as Bob (have XMR, want ETH)",
			},
			&cli.StringFlag{
				Name:  "wallet-file",
				Usage: "filename of wallet file containing XMR to be swapped; required if running as Bob",
			},
			&cli.StringFlag{
				Name:  "wallet-password",
				Usage: "password of wallet file containing XMR to be swapped",
			},
			&cli.StringFlag{
				Name:  "env",
				Usage: "environment to use: one of mainnet, stagenet, or dev",
			},
			&cli.Float64Flag{
				Name:  "max-amount", // TODO: remove this and pass it via RPC
				Value: 0,
				Usage: "maximum amount to swap (in standard units of coin)",
			},
			&cli.StringFlag{
				Name:  "monero-endpoint",
				Usage: "monero-wallet-rpc endpoint",
			},
			&cli.StringFlag{
				Name:  "monero-daemon-endpoint",
				Usage: "monerod RPC endpoint",
			},
			&cli.StringFlag{
				Name:  "ethereum-endpoint",
				Usage: "ethereum client endpoint",
			},
			&cli.StringFlag{
				Name:  "ethereum-privkey",
				Usage: "file containing a private key hex string",
			},
			&cli.UintFlag{
				Name:  "ethereum-chain-id",
				Usage: "ethereum chain ID; eg. mainnet=1, ropsten=3, rinkeby=4, goerli=5, ganache=1337",
			},
			&cli.StringFlag{
				Name:  "bootnodes",
				Usage: "comma-separated string of libp2p bootnodes",
			},
			&cli.UintFlag{
				Name:  "gas-price",
				Usage: "ethereum gas price to use for transactions (in gwei). if not set, the gas price is set via oracle.",
			},
			&cli.UintFlag{
				Name:  "gas-limit",
				Usage: "ethereum gas limit to use for transactions. if not set, the gas limit is estimated for each transaction.",
			},
		},
	}
)

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

type protocolHandler interface {
	net.Handler
	rpc.Protocol
	SetMessageSender(net.MessageSender)
}

func runDaemon(c *cli.Context) error {
	isAlice := c.Bool("alice")
	isBob := c.Bool("bob")

	if isAlice && isBob {
		return errors.New("must specify only one of --alice or --bob")
	}

	env, cfg, err := getEnvironment(c)
	if err != nil {
		return err
	}

	chainID := int64(c.Uint("ethereum-chain-id"))
	if chainID == 0 {
		chainID = cfg.EthereumChainID
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler, err := getProtocolHandler(ctx, c, env, cfg, chainID, isAlice, isBob)
	if err != nil {
		return err
	}

	amount := c.Float64("max-amount")

	var bootnodes []string
	if c.String("bootnodes") != "" {
		bootnodes = strings.Split(c.String("bootnodes"), ",")
	}

	k := c.String("libp2p-key")
	p := uint16(c.Uint("libp2p-port"))
	var (
		libp2pKey  string
		libp2pPort uint16
		provides   []common.ProvidesCoin
		rpcPort    uint16
	)

	switch {
	case k != "":
		libp2pKey = k
	case isAlice:
		libp2pKey = defaultAliceLibp2pKey
	case isBob:
		libp2pKey = defaultBobLibp2pKey
	}

	switch {
	case p != 0:
		libp2pPort = p
	case isAlice:
		libp2pPort = defaultAliceLibp2pPort
	case isBob:
		libp2pPort = defaultBobLibp2pPort
	default:
		return errors.New("must provide --libp2p-port")
	}

	switch {
	case isAlice:
		provides = []common.ProvidesCoin{common.ProvidesETH}
	case isBob:
		provides = []common.ProvidesCoin{common.ProvidesXMR}
	}

	netCfg := &net.Config{
		Ctx:           ctx,
		Environment:   env,
		ChainID:       chainID,
		Port:          libp2pPort,
		Provides:      provides,
		MaximumAmount: []float64{amount},
		ExchangeRate:  defaultExchangeRate,
		KeyFile:       libp2pKey,
		Bootnodes:     bootnodes,
		Handler:       handler,
	}

	host, err := net.NewHost(netCfg)
	if err != nil {
		return err
	}

	// connect network to protocol handler
	if isAlice || isBob {
		handler.SetMessageSender(host)
	}

	if err = host.Start(); err != nil {
		return err
	}

	p = uint16(c.Uint("rpc-port"))
	switch {
	case p != 0:
		rpcPort = p
	case isAlice:
		rpcPort = defaultAliceRPCPort
	case isBob:
		rpcPort = defaultBobRPCPort
	default:
		return errors.New("must provide --rpc-port")
	}

	mr, err := getRecoverer(c, env, isAlice, isBob)
	if err != nil {
		return err
	}

	rpcCfg := &rpc.Config{
		Port:            rpcPort,
		Net:             host,
		Protocol:        handler,
		MoneroRecoverer: mr,
	}

	s, err := rpc.NewServer(rpcCfg)
	if err != nil {
		return err
	}

	errCh := s.Start()
	go func() {
		select {
		case <-ctx.Done():
			return
		case err := <-errCh:
			log.Errorf("failed to start RPC server: %s", err)
			os.Exit(1)
		}
	}()

	log.Info("started swapd with basepath %d",
		basepath,
	)
	wait(ctx)
	return nil
}

func getEnvironment(c *cli.Context) (env common.Environment, cfg common.Config, err error) {
	switch c.String("env") {
	case "mainnet":
		env = common.Mainnet
		cfg = common.MainnetConfig
	case "stagenet":
		env = common.Stagenet
		cfg = common.StagenetConfig
	case "dev":
		env = common.Development
		cfg = common.DevelopmentConfig
	case "":
		env = defaultEnvironment
		cfg = common.DevelopmentConfig
	default:
		return 0, common.Config{}, errors.New("--env must be one of mainnet, stagenet, or dev")
	}

	return env, cfg, nil
}

func getEthereumPrivateKey(c *cli.Context, env common.Environment, isAlice bool) (ethPrivKey string, err error) {
	if c.String("ethereum-privkey") != "" {
		ethPrivKeyFile := c.String("ethereum-privkey")
		key, err := os.ReadFile(filepath.Clean(ethPrivKeyFile))
		if err != nil {
			return "", fmt.Errorf("failed to read ethereum-privkey file: %w", err)
		}

		if key[len(key)-1] == '\n' {
			key = key[:len(key)-1]
		}

		ethPrivKey = string(key)
	} else {
		if env != common.Development {
			return "", errors.New("must provide --ethereum-privkey file for non-development environment")
		}

		log.Warn("no ethereum private key file provided, using ganache deterministic key")
		if isAlice {
			ethPrivKey = common.DefaultPrivKeyAlice
		} else {
			ethPrivKey = common.DefaultPrivKeyBob
		}
	}

	return ethPrivKey, nil
}

func getRecoverer(c *cli.Context, env common.Environment, isAlice, isBob bool) (rpc.MoneroRecoverer, error) {
	var (
		moneroEndpoint, ethEndpoint string
	)

	if c.String("monero-endpoint") != "" {
		moneroEndpoint = c.String("monero-endpoint")
	} else {
		switch {
		case isAlice:
			moneroEndpoint = common.DefaultAliceMoneroEndpoint
		case isBob:
			moneroEndpoint = common.DefaultBobMoneroEndpoint
		default:
			moneroEndpoint = common.DefaultAliceMoneroEndpoint
		}
	}

	if c.String("ethereum-endpoint") != "" {
		ethEndpoint = c.String("ethereum-endpoint")
	} else {
		ethEndpoint = common.DefaultEthEndpoint
	}

	log.Info("created recovery module with monero endpoint %s and ethereum endpoint %s",
		moneroEndpoint,
		ethEndpoint,
	)
	return recovery.NewRecoverer(env, moneroEndpoint, ethEndpoint)
}

func getProtocolHandler(ctx context.Context, c *cli.Context, env common.Environment, cfg common.Config,
	chainID int64, isAlice, isBob bool) (handler protocolHandler, err error) {
	var (
		moneroEndpoint, daemonEndpoint, ethEndpoint string
	)

	if c.String("monero-endpoint") != "" {
		moneroEndpoint = c.String("monero-endpoint")
	} else {
		switch {
		case isAlice:
			moneroEndpoint = common.DefaultAliceMoneroEndpoint
		case isBob:
			moneroEndpoint = common.DefaultBobMoneroEndpoint
		default:
			moneroEndpoint = common.DefaultAliceMoneroEndpoint
		}
	}

	if c.String("ethereum-endpoint") != "" {
		ethEndpoint = c.String("ethereum-endpoint")
	} else {
		ethEndpoint = common.DefaultEthEndpoint
	}

	ethPrivKey, err := getEthereumPrivateKey(c, env, isAlice)
	if err != nil {
		return nil, err
	}

	if c.String("monero-daemon-endpoint") != "" {
		daemonEndpoint = c.String("monero-daemon-endpoint")
	} else {
		daemonEndpoint = cfg.MoneroDaemonEndpoint
	}

	// TODO: add configs for different eth testnets + L2 and set gas limit based on those, if not set
	var gasPrice *big.Int
	if c.Uint("gas-price") != 0 {
		gasPrice = big.NewInt(int64(c.Uint("gas-price")))
	}

	switch {
	case isAlice:
		aliceCfg := &alice.Config{
			Ctx:                  ctx,
			Basepath:             cfg.Basepath,
			MoneroWalletEndpoint: moneroEndpoint,
			EthereumEndpoint:     ethEndpoint,
			EthereumPrivateKey:   ethPrivKey,
			Environment:          env,
			ChainID:              chainID,
			GasPrice:             gasPrice,
			GasLimit:             uint64(c.Uint("gas-limit")),
		}

		handler, err = alice.NewAlice(aliceCfg)
		if err != nil {
			return nil, err
		}
	case isBob:
		walletFile := c.String("wallet-file")
		if walletFile == "" {
			return nil, errors.New("must provide --wallet-file")
		}

		// empty password is ok
		walletPassword := c.String("wallet-password")

		bobCfg := &bob.Config{
			Ctx:                  ctx,
			Basepath:             cfg.Basepath,
			MoneroWalletEndpoint: moneroEndpoint,
			MoneroDaemonEndpoint: daemonEndpoint,
			WalletFile:           walletFile,
			WalletPassword:       walletPassword,
			EthereumEndpoint:     ethEndpoint,
			EthereumPrivateKey:   ethPrivKey,
			Environment:          env,
			ChainID:              chainID,
			GasPrice:             gasPrice,
			GasLimit:             uint64(c.Uint("gas-limit")),
		}

		handler, err = bob.NewBob(bobCfg)
		if err != nil {
			return nil, err
		}
	}

	log.Info("created swap protocol module with monero endpoint %s and ethereum endpoint %s",
		moneroEndpoint,
		ethEndpoint,
	)
	return handler, nil
}
