package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/noot/atomic-swap/cmd/client/client"
	"github.com/noot/atomic-swap/common"

	logging "github.com/ipfs/go-log"
	"github.com/urfave/cli"
)

const (
	defaultSwapdAddress = "http://localhost:5001"
)

var log = logging.Logger("cmd")

var (
	app = &cli.App{
		Name:  "swapcli",
		Usage: "Client for swapd",
		Commands: []cli.Command{
			{
				Name:    "addresses",
				Aliases: []string{"a"},
				Usage:   "list our daemon's libp2p listening addresses",
				Action:  runAddresses,
				Flags: []cli.Flag{
					daemonAddrFlag,
				},
			},
			{
				Name:    "discover",
				Aliases: []string{"d"},
				Usage:   "discover peers who provide a certain coin",
				Action:  runDiscover,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "provides",
						Usage: "coin to find providers for: one of [ETH, XMR]",
					},
					&cli.UintFlag{
						Name:  "search-time",
						Usage: "duration of time to search for, in seconds",
					},
					daemonAddrFlag,
				},
			},
			{
				Name:    "query",
				Aliases: []string{"q"},
				Usage:   "query a peer for details on what they provide",
				Action:  runQuery,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "multiaddr",
						Usage: "peer's multiaddress, as provided by discover",
					},
					daemonAddrFlag,
				},
			},
			{
				Name:    "initiate",
				Aliases: []string{"i"},
				Usage:   "initiate a swap",
				Action:  runInitiate,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "multiaddr",
						Usage: "peer's multiaddress, as provided by discover",
					},
					&cli.StringFlag{
						Name:  "provides",
						Usage: "coin to provide in the swap: one of [ETH, XMR]",
					},
					&cli.Float64Flag{
						Name:  "provides-amount",
						Usage: "amount of coin to send in the swap",
					},
					&cli.Float64Flag{
						Name:  "desired-amount",
						Usage: "amount of coin to receive in the swap",
					},
					daemonAddrFlag,
				},
			},
		},
		Flags: []cli.Flag{daemonAddrFlag},
	}

	daemonAddrFlag = &cli.StringFlag{
		Name:  "daemon-addr",
		Usage: "address of swap daemon; default http://localhost:5001",
	}
)

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func runAddresses(ctx *cli.Context) error {
	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	c := client.NewClient(endpoint)
	addrs, err := c.Addresses()
	if err != nil {
		return err
	}

	fmt.Printf("Listening addresses: %v\n", addrs)
	return nil
}

func runDiscover(ctx *cli.Context) error {
	provides, err := common.NewProvidesCoin(ctx.String("provides"))
	if err != nil {
		return err
	}

	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	searchTime := ctx.Uint("search-time")

	c := client.NewClient(endpoint)
	peers, err := c.Discover(provides, uint64(searchTime))
	if err != nil {
		return err
	}

	for i, peer := range peers {
		fmt.Printf("Peer %d: %v\n", i, peer)
	}

	return nil
}

func runQuery(ctx *cli.Context) error {
	maddr := ctx.String("multiaddr")
	if maddr == "" {
		return errors.New("must provide peer's multiaddress with --multiaddr")
	}

	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	c := client.NewClient(endpoint)
	res, err := c.Query(maddr)
	if err != nil {
		return err
	}

	fmt.Printf("Provides: %v\n", res.Provides)
	fmt.Printf("MaximumAmount: %v\n", res.MaximumAmount)
	fmt.Printf("ExchangeRate (ETH/XMR): %v\n", res.ExchangeRate)

	return nil
}

func runInitiate(ctx *cli.Context) error {
	maddr := ctx.String("multiaddr")
	if maddr == "" {
		return errors.New("must provide peer's multiaddress with --multiaddr")
	}

	provides, err := common.NewProvidesCoin(ctx.String("provides"))
	if err != nil {
		return err
	}

	providesAmount := ctx.Float64("provides-amount")
	if providesAmount == 0 {
		return errors.New("must provide --provides-amount")
	}

	desiredAmount := ctx.Float64("desired-amount")
	if desiredAmount == 0 {
		return errors.New("must provide --desired-amount")
	}

	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	c := client.NewClient(endpoint)
	ok, err := c.Initiate(maddr, provides, providesAmount, desiredAmount)
	if err != nil {
		return err
	}

	var desiredCoin common.ProvidesCoin
	switch provides {
	case common.ProvidesETH:
		desiredCoin = common.ProvidesXMR
	case common.ProvidesXMR:
		desiredCoin = common.ProvidesETH
	}

	if ok {
		fmt.Printf("Swap successful, received %v %s\n", desiredAmount, desiredCoin)
	} else {
		fmt.Printf("Swap failed! Please check swapd logs for additional information.")
	}

	return nil
}
