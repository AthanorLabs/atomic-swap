package main

import (
	"context"
	"errors"
	"os"
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
)

var (
	log = logging.Logger("cmd")
	_   = logging.SetLogLevel("alice", "debug")
	_   = logging.SetLogLevel("bob", "debug")
	_   = logging.SetLogLevel("cmd", "debug")
	_   = logging.SetLogLevel("net", "debug")
	_   = logging.SetLogLevel("rpc", "debug")
)

var (
	app = &cli.App{
		Name:   "atomic-swap",
		Usage:  "A program for doing atomic swaps between ETH and XMR",
		Action: startAction,
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:  "rpc-port",
				Usage: "port for the daemon RPC server to run on; default 5001",
			},
			&cli.BoolFlag{
				Name:  "alice",
				Usage: "run as Alice (have ETH, want XMR)",
			},
			&cli.BoolFlag{
				Name:  "bob",
				Usage: "run as Bob (have XMR, want ETH)",
			},
			&cli.UintFlag{
				Name:  "amount",
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
				Usage: "ethereum private key hex string", // TODO: change this to a file
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
		log.Debug(err)
		os.Exit(1)
	}
}

func startAction(c *cli.Context) error {
	log.Debug("starting...")
	return runDaemon(c)

	// amount := uint64(c.Uint("amount"))
	// if amount == 0 {
	// 	return errors.New("must specify amount")
	// }

	// if c.Bool("alice") {
	// 	if err := runAlice(c, amount); err != nil {
	// 		return err
	// 	}

	// 	return nil
	// }

	// if c.Bool("bob") {
	// 	if err := runBob(c, amount); err != nil {
	// 		return err
	// 	}

	// 	return nil
	// }

	// return errors.New("must specify either --alice or --bob")
}

func runDaemon(c *cli.Context) error {
	var (
		moneroEndpoint, daemonEndpoint, ethEndpoint, ethPrivKey string
	)

	if c.String("monero-endpoint") != "" {
		moneroEndpoint = c.String("monero-endpoint")
	} else {
		moneroEndpoint = common.DefaultBobMoneroEndpoint
	}

	if c.String("ethereum-endpoint") != "" {
		ethEndpoint = c.String("ethereum-endpoint")
	} else {
		ethEndpoint = common.DefaultEthEndpoint
	}

	if c.String("ethereum-privkey") != "" {
		ethPrivKey = c.String("ethereum-privkey")
	} else {
		log.Warn("no ethereum private key provided, using ganache deterministic key at index 1")
		ethPrivKey = common.DefaultPrivKeyBob
	}

	if c.String("monero-daemon-endpoint") != "" {
		daemonEndpoint = c.String("monero-daemon-endpoint")
	} else {
		daemonEndpoint = common.DefaultDaemonEndpoint
	}

	port := uint32(c.Uint("rpc-port"))
	if port == 0 {
		port = defaultRPCPort
	}

	amount := uint64(c.Uint("amount"))
	if amount == 0 {
		return errors.New("must specify maximum provided amount")
	}

	var bootnodes []string
	if c.String("bootnodes") != "" {
		bootnodes = strings.Split(c.String("bootnodes"), ",")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	netCfg := &net.Config{
		Ctx:           ctx,
		Port:          defaultAlicePort,                    // TODO: make flag
		Provides:      []net.ProvidesCoin{net.ProvidesETH}, // TODO: make flag
		MaximumAmount: []uint64{amount},
		ExchangeRate:  defaultExchangeRate,
		KeyFile:       defaultAliceLibp2pKey, // TODO: make flag
		Bootnodes:     bootnodes,
	}

	if c.Bool("bob") {
		netCfg.Port = defaultBobPort
		netCfg.Provides = []net.ProvidesCoin{net.ProvidesXMR}
		netCfg.KeyFile = defaultBobLibp2pKey
		port = defaultRPCPort + 1
	}

	host, err := net.NewHost(netCfg)
	if err != nil {
		return err
	}

	if err = host.Start(); err != nil {
		return err
	}

	var protocol rpc.Protocol
	switch {
	case c.Bool("alice"):
		protocol, err = alice.NewAlice(ctx, moneroEndpoint, ethEndpoint, ethPrivKey)
		if err != nil {
			return err
		}
	case c.Bool("bob"):
		protocol, err = bob.NewBob(ctx, moneroEndpoint, daemonEndpoint, ethEndpoint, ethPrivKey)
		if err != nil {
			return err
		}
	default:
		return errors.New("must specify either --alice or --bob")
	}

	cfg := &rpc.Config{
		Port:     port,
		Net:      host,
		Protocol: protocol,
	}

	s, err := rpc.NewServer(cfg)
	if err != nil {
		return err
	}

	go s.Start()

	wait(ctx)
	return nil
}

// func runAlice(c *cli.Context, amount uint64) error {
// 	var (
// 		moneroEndpoint, ethEndpoint, ethPrivKey string
// 	)

// 	if c.String("monero-endpoint") != "" {
// 		moneroEndpoint = c.String("monero-endpoint")
// 	} else {
// 		moneroEndpoint = common.DefaultAliceMoneroEndpoint
// 	}

// 	if c.String("ethereum-endpoint") != "" {
// 		ethEndpoint = c.String("ethereum-endpoint")
// 	} else {
// 		ethEndpoint = common.DefaultEthEndpoint
// 	}

// 	if c.String("ethereum-privkey") != "" {
// 		ethPrivKey = c.String("ethereum-privkey")
// 	} else {
// 		log.Warn("no ethereum private key provided, using ganache deterministic key at index 0")
// 		ethPrivKey = common.DefaultPrivKeyAlice
// 	}

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	alice, err := alice.NewAlice(ctx, moneroEndpoint, ethEndpoint, ethPrivKey)
// 	if err != nil {
// 		return err
// 	}

// 	log.Debug("instantiated Alice session")

// 	var bootnodes []string
// 	if c.String("bootnodes") != "" {
// 		bootnodes = strings.Split(c.String("bootnodes"), ",")
// 	}

// 	netCfg := &net.Config{
// 		Ctx:           ctx,
// 		Port:          defaultAlicePort,
// 		Provides:      []net.ProvidesCoin{net.ProvidesETH},
// 		MaximumAmount: []uint64{amount},
// 		ExchangeRate:  defaultExchangeRate,
// 		KeyFile:       defaultAliceLibp2pKey,
// 		Bootnodes:     bootnodes,
// 	}

// 	host, err := net.NewHost(netCfg)
// 	if err != nil {
// 		return err
// 	}

// 	n := &node{
// 		ctx:    ctx,
// 		cancel: cancel,
// 		alice:  alice,
// 		host:   host,
// 		amount: amount,
// 	}

// 	return n.doProtocolAlice()
// }

// func runBob(c *cli.Context, amount uint64) error {
// 	var (
// 		moneroEndpoint, daemonEndpoint, ethEndpoint, ethPrivKey string
// 	)

// 	if c.String("monero-endpoint") != "" {
// 		moneroEndpoint = c.String("monero-endpoint")
// 	} else {
// 		moneroEndpoint = common.DefaultBobMoneroEndpoint
// 	}

// 	if c.String("ethereum-endpoint") != "" {
// 		ethEndpoint = c.String("ethereum-endpoint")
// 	} else {
// 		ethEndpoint = common.DefaultEthEndpoint
// 	}

// 	if c.String("ethereum-privkey") != "" {
// 		ethPrivKey = c.String("ethereum-privkey")
// 	} else {
// 		log.Warn("no ethereum private key provided, using ganache deterministic key at index 1")
// 		ethPrivKey = common.DefaultPrivKeyBob
// 	}

// 	if c.String("monero-daemon-endpoint") != "" {
// 		daemonEndpoint = c.String("monero-daemon-endpoint")
// 	} else {
// 		daemonEndpoint = common.DefaultDaemonEndpoint
// 	}

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	bob, err := bob.NewBob(ctx, moneroEndpoint, daemonEndpoint, ethEndpoint, ethPrivKey)
// 	if err != nil {
// 		return err
// 	}

// 	log.Debug("instantiated Bob session")

// 	var bootnodes []string
// 	if c.String("bootnodes") != "" {
// 		bootnodes = strings.Split(c.String("bootnodes"), ",")
// 	}

// 	netCfg := &net.Config{
// 		Ctx:           ctx,
// 		Port:          defaultBobPort,
// 		Provides:      []net.ProvidesCoin{net.ProvidesXMR},
// 		MaximumAmount: []uint64{amount},
// 		ExchangeRate:  defaultExchangeRate,
// 		KeyFile:       defaultBobLibp2pKey,
// 		Bootnodes:     bootnodes,
// 	}

// 	host, err := net.NewHost(netCfg)
// 	if err != nil {
// 		return err
// 	}

// 	n := &node{
// 		ctx:    ctx,
// 		cancel: cancel,
// 		bob:    bob,
// 		host:   host,
// 		amount: amount,
// 	}

// 	return n.doProtocolBob()
// }
