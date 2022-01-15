package main

import (
	"context"
	"math/big"
	"os"
	"strings"

	"github.com/urfave/cli"

	"github.com/noot/atomic-swap/cmd/utils"
	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/protocol/alice"
	"github.com/noot/atomic-swap/protocol/bob"
	"github.com/noot/atomic-swap/rpc"

	logging "github.com/ipfs/go-log"
)

const (
	// default libp2p ports
	defaultLibp2pPort      = 9900
	defaultAliceLibp2pPort = 9933
	defaultBobLibp2pPort   = 9934

	// default libp2p key files
	defaultLibp2pKey      = "node.key"
	defaultAliceLibp2pKey = "alice.key"
	defaultBobLibp2pKey   = "bob.key"

	// default RPC port
	defaultRPCPort      = 5005
	defaultAliceRPCPort = 5001
	defaultBobRPCPort   = 5002
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
		Name:   "swapd",
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
			&cli.BoolFlag{
				Name: "dev-alice",
			},
			&cli.BoolFlag{
				Name: "dev-bob",
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

type aliceHandler interface {
	rpc.Alice
	SetMessageSender(net.MessageSender)
}

type bobHandler interface {
	net.Handler
	rpc.Bob
	SetMessageSender(net.MessageSender)
}

func runDaemon(c *cli.Context) error {
	env, cfg, err := utils.GetEnvironment(c)
	if err != nil {
		return err
	}

	devAlice := c.Bool("dev-alice")
	devBob := c.Bool("dev-bob")

	chainID := int64(c.Uint("ethereum-chain-id"))
	if chainID == 0 {
		chainID = cfg.EthereumChainID
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a, b, err := getProtocolInstances(ctx, c, env, cfg, chainID, devBob)
	if err != nil {
		return err
	}

	var bootnodes []string
	if c.String("bootnodes") != "" {
		bootnodes = strings.Split(c.String("bootnodes"), ",")
	}

	k := c.String("libp2p-key")
	p := uint16(c.Uint("libp2p-port"))
	var (
		libp2pKey  string
		libp2pPort uint16
		rpcPort    uint16
	)

	switch {
	case k != "":
		libp2pKey = k
	case devAlice:
		libp2pKey = defaultAliceLibp2pKey
	case devBob:
		libp2pKey = defaultBobLibp2pKey
	default:
		libp2pKey = defaultLibp2pKey
	}

	switch {
	case p != 0:
		libp2pPort = p
	case devAlice:
		libp2pPort = defaultAliceLibp2pPort
	case devBob:
		libp2pPort = defaultBobLibp2pPort
	default:
		libp2pPort = defaultLibp2pPort
		//	return errors.New("must provide --libp2p-port")
	}

	netCfg := &net.Config{
		Ctx:         ctx,
		Environment: env,
		ChainID:     chainID,
		Port:        libp2pPort,
		KeyFile:     libp2pKey,
		Bootnodes:   bootnodes,
		Handler:     b, // handler handles initiated ("taken") swaps
	}

	host, err := net.NewHost(netCfg)
	if err != nil {
		return err
	}

	// connect network to protocol handlers
	a.SetMessageSender(host)
	b.SetMessageSender(host)

	if err = host.Start(); err != nil {
		return err
	}

	p = uint16(c.Uint("rpc-port"))
	switch {
	case p != 0:
		rpcPort = p
	case devAlice:
		rpcPort = defaultAliceRPCPort
	case devBob:
		rpcPort = defaultBobRPCPort
	default:
		rpcPort = defaultRPCPort
	}

	rpcCfg := &rpc.Config{
		Port:  rpcPort,
		Net:   host,
		Alice: a,
		Bob:   b,
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
		cfg.Basepath,
	)
	wait(ctx)
	return nil
}

func getProtocolInstances(ctx context.Context, c *cli.Context, env common.Environment, cfg common.Config,
	chainID int64, devBob bool) (a aliceHandler, b bobHandler, err error) {
	var (
		moneroEndpoint, daemonEndpoint, ethEndpoint string
	)

	if c.String("monero-endpoint") != "" {
		moneroEndpoint = c.String("monero-endpoint")
	} else if devBob {
		moneroEndpoint = common.DefaultBobMoneroEndpoint
	} else {
		moneroEndpoint = common.DefaultAliceMoneroEndpoint
	}

	if c.String("ethereum-endpoint") != "" {
		ethEndpoint = c.String("ethereum-endpoint")
	} else {
		ethEndpoint = common.DefaultEthEndpoint
	}

	ethPrivKey, err := utils.GetEthereumPrivateKey(c, env, devBob)
	if err != nil {
		return nil, nil, err
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

	a, err = alice.NewInstance(aliceCfg)
	if err != nil {
		return nil, nil, err
	}

	walletFile := c.String("wallet-file")

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

	b, err = bob.NewInstance(bobCfg)
	if err != nil {
		return nil, nil, err
	}

	log.Info("created swap protocol module with monero endpoint %s and ethereum endpoint %s",
		moneroEndpoint,
		ethEndpoint,
	)
	return a, b, nil
}
