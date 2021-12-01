package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"

	"github.com/noot/atomic-swap/alice"
	"github.com/noot/atomic-swap/bob"
	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/rpc"

	logging "github.com/ipfs/go-log"
)

const (
	// default libp2p ports
	defaultAlicePort = 9933
	defaultBobPort   = 9934

	// defaultExchangeRate is the default ratio of ETH:XMR.
	// TODO; make this a CLI flag, or get it from some price feed.
	defaultExchangeRate = 0.0578261

	// default libp2p key files
	defaultAliceLibp2pKey = "alice.key"
	defaultBobLibp2pKey   = "bob.key"

	// default RPC port
	defaultRPCPort = 5001

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
				Usage: "path to store swap artifacts",
			},
			&cli.StringFlag{
				Name:  "libp2p-key",
				Usage: "libp2p private key",
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
				Name:  "amount", // TODO: remove this and pass it via RPC
				Value: 0,
				Usage: "maximum amount to swap (in smallest units of coin)",
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
		},
	}
)

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func runDaemon(c *cli.Context) error {
	var (
		moneroEndpoint, daemonEndpoint, ethEndpoint, ethPrivKeyFile, ethPrivKey string
		env                                                                     common.Environment
		cfg                                                                     common.Config
	)

	isAlice := c.Bool("alice")
	isBob := c.Bool("bob")

	if !isAlice && !isBob {
		return errors.New("must specify either --alice or --bob")
	}

	if isAlice && isBob {
		return errors.New("must specify only one of --alice or --bob")
	}

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
		return errors.New("--env must be one of mainnet, stagenet, or dev")
	}

	if c.String("monero-endpoint") != "" {
		moneroEndpoint = c.String("monero-endpoint")
	} else {
		if isAlice {
			moneroEndpoint = common.DefaultAliceMoneroEndpoint
		} else {
			moneroEndpoint = common.DefaultBobMoneroEndpoint
		}
	}

	if c.String("ethereum-endpoint") != "" {
		ethEndpoint = c.String("ethereum-endpoint")
	} else {
		ethEndpoint = common.DefaultEthEndpoint
	}

	// TODO: if env isn't development, require a private key
	if c.String("ethereum-privkey") != "" {
		ethPrivKeyFile = c.String("ethereum-privkey")
		key, err := os.ReadFile(filepath.Clean(ethPrivKeyFile))
		if err != nil {
			return fmt.Errorf("failed to read ethereum-privkey file: %w", err)
		}

		if key[len(key)-1] == '\n' {
			key = key[:len(key)-1]
		}

		ethPrivKey = string(key)
	} else {
		if env != common.Development {
			return errors.New("must provide --ethereum-privkey file for non-development environment")
		}

		log.Warn("no ethereum private key file provided, using ganache deterministic key")
		if isAlice {
			ethPrivKey = common.DefaultPrivKeyAlice
		} else {
			ethPrivKey = common.DefaultPrivKeyBob
		}
	}

	chainID := int64(c.Uint("ethereum-chain-id"))
	if chainID == 0 {
		chainID = cfg.EthereumChainID
	}

	if c.String("monero-daemon-endpoint") != "" {
		daemonEndpoint = c.String("monero-daemon-endpoint")
	} else {
		daemonEndpoint = cfg.MoneroDaemonEndpoint
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type Handler interface {
		net.Handler
		rpc.Protocol
		SetMessageSender(net.MessageSender)
	}

	var (
		handler Handler
		err     error
	)
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
		}

		handler, err = alice.NewAlice(aliceCfg)
		if err != nil {
			return err
		}
	case isBob:
		walletFile := c.String("wallet-file")
		if walletFile == "" {
			return errors.New("must provide --wallet-file")
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
		}

		handler, err = bob.NewBob(bobCfg)
		if err != nil {
			return err
		}
	default:
		return errors.New("must specify either --alice or --bob")
	}

	port := uint32(c.Uint("rpc-port"))
	if port == 0 {
		port = defaultRPCPort
	}

	amount := float64(c.Float64("amount"))

	var bootnodes []string
	if c.String("bootnodes") != "" {
		bootnodes = strings.Split(c.String("bootnodes"), ",")
	}

	k := c.String("libp2p-key")
	var libp2pKey string
	switch {
	case k != "":
		libp2pKey = k
	case isAlice:
		libp2pKey = defaultAliceLibp2pKey
	case isBob:
		libp2pKey = defaultBobLibp2pKey
	}

	netCfg := &net.Config{
		Ctx:           ctx,
		Environment:   env,
		ChainID:       chainID,
		Port:          defaultAlicePort,                          // TODO: make flag
		Provides:      []common.ProvidesCoin{common.ProvidesETH}, // TODO: make flag
		MaximumAmount: []float64{amount},
		ExchangeRate:  defaultExchangeRate,
		KeyFile:       libp2pKey,
		Bootnodes:     bootnodes,
		Handler:       handler,
	}

	// TODO: this is ugly
	if c.Bool("bob") {
		netCfg.Port = defaultBobPort
		netCfg.Provides = []common.ProvidesCoin{common.ProvidesXMR}
		port = defaultRPCPort + 1
	}

	host, err := net.NewHost(netCfg)
	if err != nil {
		return err
	}

	// connect network to protocol handler
	handler.SetMessageSender(host)

	if err = host.Start(); err != nil {
		return err
	}

	rpcCfg := &rpc.Config{
		Port:     port,
		Net:      host,
		Protocol: handler,
	}

	s, err := rpc.NewServer(rpcCfg)
	if err != nil {
		return err
	}

	go s.Start()

	wait(ctx)
	return nil
}
