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

	logging "github.com/ipfs/go-log"
)

const (
	// default libp2p ports
	defaultAlicePort = 9933
	defaultBobPort   = 9934

	// default libp2p key files
	defaultAliceLibp2pKey = "alice.key"
	defaultBobLibp2pKey   = "bob.key"
)

var (
	log = logging.Logger("cmd")
	_   = logging.SetLogLevel("alice", "debug")
	_   = logging.SetLogLevel("bob", "debug")
	_   = logging.SetLogLevel("cmd", "debug")
	_   = logging.SetLogLevel("net", "debug")
)

var (
	app = &cli.App{
		Name:   "atomic-swap",
		Usage:  "A program for doing atomic swaps between ETH and XMR",
		Action: startAction,
		Flags: []cli.Flag{
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
				Usage: "amount to swap (in smallest units of chain)",
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
	amount := c.Uint("amount")
	if amount == 0 {
		return errors.New("must specify amount")
	}

	if c.Bool("alice") {
		if err := runAlice(c, amount); err != nil {
			return err
		}

		return nil
	}

	if c.Bool("bob") {
		if err := runBob(c, amount); err != nil {
			return err
		}

		return nil
	}

	return errors.New("must specify either --alice or --bob")
}

func runAlice(c *cli.Context, amount uint) error {
	var (
		moneroEndpoint, ethEndpoint, ethPrivKey string
	)

	if c.String("monero-endpoint") != "" {
		moneroEndpoint = c.String("monero-endpoint")
	} else {
		moneroEndpoint = common.DefaultAliceMoneroEndpoint
	}

	if c.String("ethereum-endpoint") != "" {
		ethEndpoint = c.String("ethereum-endpoint")
	} else {
		ethEndpoint = common.DefaultEthEndpoint
	}

	if c.String("ethereum-privkey") != "" {
		ethPrivKey = c.String("ethereum-privkey")
	} else {
		log.Warn("no ethereum private key provided, using ganache deterministic key at index 0")
		ethPrivKey = common.DefaultPrivKeyAlice
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	alice, err := alice.NewAlice(ctx, moneroEndpoint, ethEndpoint, ethPrivKey)
	if err != nil {
		return err
	}

	log.Debug("instantiated Alice session")

	var bootnodes []string
	if c.String("bootnodes") != "" {
		bootnodes = strings.Split(c.String("bootnodes"), ",")
	}

	host, err := net.NewHost(ctx, defaultAlicePort, "XMR", amount, defaultAliceLibp2pKey, bootnodes)
	if err != nil {
		return err
	}

	n := &node{
		ctx:    ctx,
		cancel: cancel,
		alice:  alice,
		host:   host,
		amount: amount,
	}

	return n.doProtocolAlice()
}

func runBob(c *cli.Context, amount uint) error {
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bob, err := bob.NewBob(ctx, moneroEndpoint, daemonEndpoint, ethEndpoint, ethPrivKey)
	if err != nil {
		return err
	}

	log.Debug("instantiated Bob session")

	var bootnodes []string
	if c.String("bootnodes") != "" {
		bootnodes = strings.Split(c.String("bootnodes"), ",")
	}

	host, err := net.NewHost(ctx, defaultBobPort, "ETH", amount, defaultBobLibp2pKey, bootnodes)
	if err != nil {
		return err
	}

	n := &node{
		ctx:    ctx,
		cancel: cancel,
		bob:    bob,
		host:   host,
		amount: amount,
	}

	return n.doProtocolBob()
}
