package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/rpcclient"
	"github.com/noot/atomic-swap/rpcclient/wsclient"

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
					&cli.BoolFlag{
						Name:  "subscribe",
						Usage: "subscribe to push notifications about the swap's status",
					},
					&cli.StringFlag{
						Name:  "eth-asset",
						Usage: "Ethereum ERC-20 token address to receive, or the zero address for regular ETH",
					},
					daemonAddrFlag,
				},
			},
			{
				Name:    "take",
				Aliases: []string{"t"},
				Usage:   "initiate a swap by taking an offer; currently only eth holders can be the takers",
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
					&cli.BoolFlag{
						Name:  "subscribe",
						Usage: "subscribe to push notifications about the swap's status",
					},
					daemonAddrFlag,
				},
			},
			{
				Name:   "get-past-swap-ids",
				Usage:  "get past swap IDs",
				Action: runGetPastSwapIDs,
				Flags:  []cli.Flag{daemonAddrFlag},
			},
			{
				Name:   "get-ongoing-swap",
				Usage:  "get information about ongoing swap, if there is one",
				Action: runGetOngoingSwap,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "offer-id",
						Usage: "ID of swap to retrieve info for",
					},
					daemonAddrFlag,
				},
			},
			{
				Name:   "get-past-swap",
				Usage:  "get information about a past swap with the given ID",
				Action: runGetPastSwap,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "offer-id",
						Usage: "ID of swap to retrieve info for",
					},
					daemonAddrFlag,
				},
			},
			{
				Name:   "refund",
				Usage:  "if we are the ETH provider for an ongoing swap, refund it if possible.",
				Action: runRefund,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "offer-id",
						Usage: "ID of swap to retrieve info for",
					},
					daemonAddrFlag,
				},
			},
			{
				Name:   "cancel",
				Usage:  "cancel a ongoing swap if possible.",
				Action: runCancel,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "offer-id",
						Usage: "ID of swap to retrieve info for",
					},
					daemonAddrFlag,
				},
			},
			{
				Name:   "clear-offers",
				Usage:  "clear current offers. if no offer IDs are provided, clears all current offers.",
				Action: runClearOffers,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "offer-ids",
						Usage: "a comma-separated list of offer IDs to delete",
					},
					daemonAddrFlag,
				},
			},
			{
				Name:   "get-stage",
				Usage:  "get the stage of a current swap.",
				Action: runGetStage,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "offer-id",
						Usage: "ID of swap to retrieve info for",
					},
					daemonAddrFlag,
				},
			},
			{
				Name:   "set-swap-timeout",
				Usage:  "set the duration between swap initiation and t0 and t0 and t1, in seconds",
				Action: runSetSwapTimeout,
				Flags: []cli.Flag{
					&cli.UintFlag{
						Name:  "duration",
						Usage: "duration of timeout, in seconds",
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

	c := rpcclient.NewClient(endpoint)
	addrs, err := c.Addresses()
	if err != nil {
		return err
	}

	fmt.Printf("Listening addresses: %v\n", addrs)
	return nil
}

func runDiscover(ctx *cli.Context) error {
	provides, err := types.NewProvidesCoin(ctx.String("provides"))
	if err != nil {
		return err
	}

	if provides == "" {
		provides = types.ProvidesXMR
	}

	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	searchTime := ctx.Uint("search-time")

	c := rpcclient.NewClient(endpoint)
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
		return errNoMultiaddr
	}

	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	c := rpcclient.NewClient(endpoint)
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
		return errNoMinAmount
	}

	max := ctx.Float64("max-amount")
	if max == 0 {
		return errNoMaxAmount
	}

	exchangeRate := ctx.Float64("exchange-rate")
	if exchangeRate == 0 {
		return errNoExchangeRate
	}

	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	if ctx.Bool("subscribe") {
		c, err := wsclient.NewWsClient(context.Background(), endpoint)
		if err != nil {
			return err
		}

		id, statusCh, err := c.MakeOfferAndSubscribe(min, max, types.ExchangeRate(exchangeRate))
		if err != nil {
			return err
		}

		fmt.Printf("Made offer with ID %s\n", id)

		for stage := range statusCh {
			fmt.Printf("> Stage updated: %s\n", stage)
			if !stage.IsOngoing() {
				return nil
			}
		}

		return nil
	}

	ethAssetStr := ctx.String("eth-asset")
	ethAsset := types.EthAsset(common.HexToAddress("0x0000000000000000000000000000000000000000"))
	if ethAssetStr != "" {
		ethAsset = types.EthAsset(common.HexToAddress(ethAssetStr))
	}

	c := rpcclient.NewClient(endpoint)
	id, err := c.MakeOffer(min, max, exchangeRate, ethAsset)
	if err != nil {
		return err
	}

	fmt.Printf("Published offer with ID %s\n", id)
	return nil
}

func runTake(ctx *cli.Context) error {
	maddr := ctx.String("multiaddr")
	if maddr == "" {
		return errNoMultiaddr
	}

	offerID := ctx.String("offer-id")
	if offerID == "" {
		return errNoOfferID
	}

	providesAmount := ctx.Float64("provides-amount")
	if providesAmount == 0 {
		return errNoProvidesAmount
	}

	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	if ctx.Bool("subscribe") {
		c, err := wsclient.NewWsClient(context.Background(), endpoint)
		if err != nil {
			return err
		}

		statusCh, err := c.TakeOfferAndSubscribe(maddr, offerID, providesAmount)
		if err != nil {
			return err
		}

		fmt.Printf("Initiated swap with ID %s\n", offerID)

		for stage := range statusCh {
			fmt.Printf("> Stage updated: %s\n", stage)
			if !stage.IsOngoing() {
				return nil
			}
		}

		return nil
	}

	c := rpcclient.NewClient(endpoint)
	err := c.TakeOffer(maddr, offerID, providesAmount)
	if err != nil {
		return err
	}

	fmt.Printf("Initiated swap with ID %s\n", offerID)
	return nil
}

func runGetPastSwapIDs(ctx *cli.Context) error {
	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	c := rpcclient.NewClient(endpoint)
	ids, err := c.GetPastSwapIDs()
	if err != nil {
		return err
	}

	fmt.Printf("Past swap IDs: %v\n", ids)
	return nil
}

func runGetOngoingSwap(ctx *cli.Context) error {
	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	offerID := ctx.String("offer-id")
	if offerID == "" {
		return errNoOfferID
	}

	c := rpcclient.NewClient(endpoint)
	info, err := c.GetOngoingSwap(offerID)
	if err != nil {
		return err
	}

	fmt.Printf("Provided: %s\n ProvidedAmount: %v\n ReceivedAmount: %v\n ExchangeRate: %v\n Status: %s\n",
		info.Provided,
		info.ProvidedAmount,
		info.ReceivedAmount,
		info.ExchangeRate,
		info.Status,
	)
	return nil
}

func runGetPastSwap(ctx *cli.Context) error {
	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	offerID := ctx.String("offer-id")
	if offerID == "" {
		return errNoOfferID
	}

	c := rpcclient.NewClient(endpoint)
	info, err := c.GetPastSwap(offerID)
	if err != nil {
		return err
	}

	fmt.Printf("Provided: %s\n ProvidedAmount: %v\n ReceivedAmount: %v\n ExchangeRate: %v\n Status: %s\n",
		info.Provided,
		info.ProvidedAmount,
		info.ReceivedAmount,
		info.ExchangeRate,
		info.Status,
	)
	return nil
}

func runRefund(ctx *cli.Context) error {
	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	offerID := ctx.String("offer-id")
	if offerID == "" {
		return errNoOfferID
	}

	c := rpcclient.NewClient(endpoint)
	resp, err := c.Refund(offerID)
	if err != nil {
		return err
	}

	fmt.Printf("Refunded successfully, transaction hash: %s\n", resp.TxHash)
	return nil
}

func runCancel(ctx *cli.Context) error {
	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	offerID := ctx.String("offer-id")
	if offerID == "" {
		return errNoOfferID
	}

	c := rpcclient.NewClient(endpoint)
	resp, err := c.Cancel(offerID)
	if err != nil {
		return err
	}

	fmt.Printf("Cancelled successfully, exit status: %s\n", resp)
	return nil
}

func runClearOffers(ctx *cli.Context) error {
	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	c := rpcclient.NewClient(endpoint)

	ids := ctx.String("offer-ids")
	if ids == "" {
		err := c.ClearOffers(nil)
		if err != nil {
			return err
		}

		fmt.Printf("Cleared all offers successfully.\n")
		return nil
	}

	err := c.ClearOffers(strings.Split(ids, ","))
	if err != nil {
		return err
	}

	fmt.Printf("Cleared offers successfully: %s\n", ids)
	return nil
}

func runGetStage(ctx *cli.Context) error {
	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	offerID := ctx.String("offer-id")
	if offerID == "" {
		return errNoOfferID
	}

	c := rpcclient.NewClient(endpoint)
	resp, err := c.GetStage(offerID)
	if err != nil {
		return err
	}

	fmt.Printf("Stage=%s: %s\n", resp.Stage, resp.Info)
	return nil
}

func runSetSwapTimeout(ctx *cli.Context) error {
	duration := ctx.Uint("duration")

	endpoint := ctx.String("daemon-addr")
	if endpoint == "" {
		endpoint = defaultSwapdAddress
	}

	c := rpcclient.NewClient(endpoint)
	err := c.SetSwapTimeout(uint64(duration))
	if err != nil {
		return err
	}

	fmt.Printf("Set timeout duration to %ds", duration)
	return nil
}
