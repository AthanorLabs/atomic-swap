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
				Name:    "make",
				Aliases: []string{"m"},
				Usage:   "mke a swap offer; currently monero holders must be the makers",
				Action:  runMake,
				Flags: []cli.Flag{
					&cli.Float64Flag{
						Name:  "min-amount",
						Usage: "minimum amount to be swapped, in XMR",
					},
					&cli.Float64Flag{
						Name:  "max-amount",
						Usage: "maximum amount to be swapped, in XMR",
					},
					&cli.Float64Flag{
						Name:  "exchange-rate",
						Usage: "desired exchange rate of XMR:ETH, eg. --exchange-rate=0.1 means 10XMR = 1ETH",
					},
					daemonAddrFlag,
				},
			},
			{
				Name:    "take",
				Aliases: []string{"t"},
				Usage:   "initiate a swap by taking an offerl currently only eth holders can be the takers",
				Action:  runTake,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "multiaddr",
						Usage: "peer's multiaddress, as provided by discover",
					},
					&cli.StringFlag{
						Name:  "offer-id",
						Usage: "ID of the offer being taken",
					},
					&cli.Float64Flag{
						Name:  "provides-amount",
						Usage: "amount of coin to send in the swap",
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

	if provides == "" {
		provides = common.ProvidesXMR
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

	for _, o := range res.Offers {
		fmt.Printf("%v\n", o)
	}
	return nil
}

func runMake(ctx *cli.Context) error {
	min := ctx.Float64("min-amount")
	if min == 0 {
		return errors.New("must provide non-zero --min-amount")
	}

	max := ctx.Float64("max-amount")
	if max == 0 {
		return errors.New("must provide non-zero --max-amount")
	}

	exchangeRate := ctx.Float64("exchange-rate")
	if exchangeRate == 0 {
		return errors.New("must provide non-zero --exchange-rate")
	}

	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	c := client.NewClient(endpoint)
	id, err := c.MakeOffer(min, max, exchangeRate)
	if err != nil {
		return err
	}

	fmt.Printf("Published offer with ID %s\n", id)
	return nil
}

func runTake(ctx *cli.Context) error {
	maddr := ctx.String("multiaddr")
	if maddr == "" {
		return errors.New("must provide peer's multiaddress with --multiaddr")
	}

	offerID := ctx.String("offer-id")
	if offerID == "" {
		return errors.New("must provide --offer-id")
	}

	providesAmount := ctx.Float64("provides-amount")
	if providesAmount == 0 {
		return errors.New("must provide --provides-amount")
	}

	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	c := client.NewClient(endpoint)
	ok, received, err := c.TakeOffer(maddr, offerID, providesAmount)
	if err != nil {
		return err
	}

	if ok {
		fmt.Printf("Swap successful, received %v ETH\n", received)
	} else {
		fmt.Printf("Swap failed! Please check swapd logs for additional information.")
	}

	return nil
}
