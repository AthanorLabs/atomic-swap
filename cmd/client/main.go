package main

import (
	"fmt"
	"os"
	"strings"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/rpcclient"
	"github.com/athanorlabs/atomic-swap/rpcclient/wsclient"
)

const (
	defaultSwapdPort              = 5001
	defaultDiscoverSearchTimeSecs = 12

	flagSwapdPort = "swapd-port"
)

var (
	app = &cli.App{
		Name:  "swapcli",
		Usage: "Client for swapd",
		Commands: []*cli.Command{
			{
				Name:    "addresses",
				Aliases: []string{"a"},
				Usage:   "List our daemon's libp2p listening addresses",
				Action:  runAddresses,
				Flags: []cli.Flag{
					swapdPortFlag,
				},
			},
			{
				Name:    "discover",
				Aliases: []string{"d"},
				Usage:   "Discover peers who provide a certain coin",
				Action:  runDiscover,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "provides",
						Usage: fmt.Sprintf("Coin to find providers for: one of [%s, %s]",
							types.ProvidesXMR, types.ProvidesETH),
						Value: string(types.ProvidesXMR),
					},
					&cli.UintFlag{
						Name:  "search-time",
						Usage: "Duration of time to search for, in seconds",
						Value: defaultDiscoverSearchTimeSecs,
					},
					swapdPortFlag,
				},
			},
			{
				Name:    "query",
				Aliases: []string{"q"},
				Usage:   "Query a peer for details on what they provide",
				Action:  runQuery,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "multiaddr",
						Usage:    "Peer's multiaddress, as provided by discover",
						Required: true,
					},
					swapdPortFlag,
				},
			},
			{
				Name:    "query-all",
				Aliases: []string{"qall"},
				Usage:   "discover peers that provide a certain coin and their offers",
				Action:  runQueryAll,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "provides",
						Usage: fmt.Sprintf("Coin to find providers for: one of [%s, %s]",
							types.ProvidesXMR, types.ProvidesETH),
						Value: string(types.ProvidesXMR),
					},
					&cli.UintFlag{
						Name:  "search-time",
						Usage: "Duration of time to search for, in seconds",
						Value: defaultDiscoverSearchTimeSecs,
					},
					swapdPortFlag,
				},
			},
			{
				Name:    "make",
				Aliases: []string{"m"},
				Usage:   "Make a swap offer; currently monero holders must be the makers",
				Action:  runMake,
				Flags: []cli.Flag{
					&cli.Float64Flag{
						Name:     "min-amount",
						Usage:    "Minimum amount to be swapped, in XMR",
						Required: true,
					},
					&cli.Float64Flag{
						Name:     "max-amount",
						Usage:    "Maximum amount to be swapped, in XMR",
						Required: true,
					},
					&cli.Float64Flag{
						Name:     "exchange-rate",
						Usage:    "Desired exchange rate of XMR:ETH, eg. --exchange-rate=0.1 means 10XMR = 1ETH",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "subscribe",
						Usage: "Subscribe to push notifications about the swap's status",
					},
					&cli.StringFlag{
						Name:  "eth-asset",
						Usage: "Ethereum ERC-20 token address to receive, or the zero address for regular ETH",
					},
					swapdPortFlag,
				},
			},
			{
				Name:    "take",
				Aliases: []string{"t"},
				Usage:   "Initiate a swap by taking an offer; currently only eth holders can be the takers",
				Action:  runTake,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "multiaddr",
						Usage:    "Peer's multiaddress, as provided by discover",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "offer-id",
						Usage:    "ID of the offer being taken",
						Required: true,
					},
					&cli.Float64Flag{
						Name:     "provides-amount",
						Usage:    "Amount of coin to send in the swap",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "subscribe",
						Usage: "Subscribe to push notifications about the swap's status",
					},
					swapdPortFlag,
				},
			},
			{
				Name:   "get-past-swap-ids",
				Usage:  "Get past swap IDs",
				Action: runGetPastSwapIDs,
				Flags:  []cli.Flag{swapdPortFlag},
			},
			{
				Name:   "get-ongoing-swap",
				Usage:  "Get information about ongoing swap, if there is one",
				Action: runGetOngoingSwap,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "offer-id",
						Usage:    "ID of swap to retrieve info for",
						Required: true,
					},
					swapdPortFlag,
				},
			},
			{
				Name:   "get-past-swap",
				Usage:  "Get information about a past swap with the given ID",
				Action: runGetPastSwap,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "offer-id",
						Usage:    "ID of swap to retrieve info for",
						Required: true,
					},
					swapdPortFlag,
				},
			},
			{
				Name:   "refund",
				Usage:  "If we are the ETH provider for an ongoing swap, refund it if possible.",
				Action: runRefund,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "offer-id",
						Usage:    "ID of swap to retrieve info for",
						Required: true,
					},
					swapdPortFlag,
				},
			},
			{
				Name:   "cancel",
				Usage:  "Cancel a ongoing swap if possible.",
				Action: runCancel,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "offer-id",
						Usage: "ID of swap to retrieve info for",
					},
					swapdPortFlag,
				},
			},
			{
				Name:   "clear-offers",
				Usage:  "Clear current offers. If no offer IDs are provided, clears all current offers.",
				Action: runClearOffers,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "offer-ids",
						Usage: "A comma-separated list of offer IDs to delete",
					},
					swapdPortFlag,
				},
			},
			{
				Name:   "get-stage",
				Usage:  "Get the stage of a current swap.",
				Action: runGetStage,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "offer-id",
						Usage:    "ID of swap to retrieve info for",
						Required: true,
					},
					swapdPortFlag,
				},
			},
			{
				Name:   "set-swap-timeout",
				Usage:  "Set the duration between swap initiation and t0 and t0 and t1, in seconds",
				Action: runSetSwapTimeout,
				Flags: []cli.Flag{
					&cli.UintFlag{
						Name:     "duration",
						Usage:    "Duration of timeout, in seconds",
						Required: true,
					},
					swapdPortFlag,
				},
			},
		},
		Flags: []cli.Flag{swapdPortFlag},
	}

	swapdPortFlag = &cli.UintFlag{
		Name:  flagSwapdPort,
		Usage: "RPC port of swap daemon",
		Value: defaultSwapdPort,
	}
)

func main() {
	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func newRRPClient(ctx *cli.Context) *rpcclient.Client {
	swapdPort := ctx.Uint(flagSwapdPort)
	endpoint := fmt.Sprintf("http://127.0.0.1:%d", swapdPort)
	return rpcclient.NewClient(endpoint)
}

func newWSClient(ctx *cli.Context) (wsclient.WsClient, error) {
	swapdPort := ctx.Uint(flagSwapdPort)
	endpoint := fmt.Sprintf("ws://127.0.0.1:%d/ws", swapdPort)
	return wsclient.NewWsClient(ctx.Context, endpoint)
}

func runAddresses(ctx *cli.Context) error {
	c := newRRPClient(ctx)
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

	searchTime := ctx.Uint("search-time")

	c := newRRPClient(ctx)
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

	c := newRRPClient(ctx)
	res, err := c.Query(maddr)
	if err != nil {
		return err
	}

	for _, o := range res.Offers {
		fmt.Printf("%v\n", o)
	}
	return nil
}

func runQueryAll(ctx *cli.Context) error {
	provides, err := types.NewProvidesCoin(ctx.String("provides"))
	if err != nil {
		return err
	}

	searchTime := ctx.Uint("search-time")

	c := newRRPClient(ctx)
	peers, err := c.QueryAll(provides, uint64(searchTime))
	if err != nil {
		return err
	}

	for i, peer := range peers {
		fmt.Printf("Peer %d:\n", i)
		fmt.Printf("\tMultiaddress: %v\n", peer.Peer)
		fmt.Printf("\tOffers:\n")
		for _, o := range peer.Offers {
			fmt.Printf("\t%v\n", o)
		}
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
	otherMin := min * exchangeRate
	otherMax := max * exchangeRate

	ethAssetStr := ctx.String("eth-asset")
	ethAsset := types.EthAssetETH
	if ethAssetStr != "" {
		ethAsset = types.EthAsset(ethcommon.HexToAddress(ethAssetStr))
	}

	c := newRRPClient(ctx)
	ourAddresses, err := c.Addresses()
	if err != nil {
		return err
	}

	printOffferSummary := func(offerID string) {
		fmt.Printf("Published offer with ID: %s\n", offerID)
		fmt.Printf("On addresses: %v\n", ourAddresses)
		fmt.Printf("Takers can provide between %s to %s %s\n",
			common.FmtFloat(otherMin), common.FmtFloat(otherMax), ethAsset)
	}

	if ctx.Bool("subscribe") {
		wsc, err := newWSClient(ctx) //nolint:govet
		if err != nil {
			return err
		}
		defer wsc.Close()

		id, statusCh, err := wsc.MakeOfferAndSubscribe(min, max, types.ExchangeRate(exchangeRate), ethAsset)
		if err != nil {
			return err
		}

		printOffferSummary(id)

		for stage := range statusCh {
			fmt.Printf("> Stage updated: %s\n", stage)
			if !stage.IsOngoing() {
				return nil
			}
		}

		return nil
	}

	id, err := c.MakeOffer(min, max, exchangeRate, ethAsset)
	if err != nil {
		return err
	}

	printOffferSummary(id)
	return nil
}

func runTake(ctx *cli.Context) error {
	maddr := ctx.String("multiaddr")
	offerID := ctx.String("offer-id")
	providesAmount := ctx.Float64("provides-amount")
	if providesAmount == 0 {
		return errNoProvidesAmount
	}

	if ctx.Bool("subscribe") {
		wsc, err := newWSClient(ctx)
		if err != nil {
			return err
		}
		defer wsc.Close()

		statusCh, err := wsc.TakeOfferAndSubscribe(maddr, offerID, providesAmount)
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

	c := newRRPClient(ctx)
	err := c.TakeOffer(maddr, offerID, providesAmount)
	if err != nil {
		return err
	}

	fmt.Printf("Initiated swap with ID %s\n", offerID)
	return nil
}

func runGetPastSwapIDs(ctx *cli.Context) error {
	c := newRRPClient(ctx)
	ids, err := c.GetPastSwapIDs()
	if err != nil {
		return err
	}

	fmt.Printf("Past swap IDs: %v\n", ids)
	return nil
}

func runGetOngoingSwap(ctx *cli.Context) error {
	offerID := ctx.String("offer-id")

	c := newRRPClient(ctx)
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
	offerID := ctx.String("offer-id")

	c := newRRPClient(ctx)
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
	offerID := ctx.String("offer-id")

	c := newRRPClient(ctx)
	resp, err := c.Refund(offerID)
	if err != nil {
		return err
	}

	fmt.Printf("Refunded successfully, transaction hash: %s\n", resp.TxHash)
	return nil
}

func runCancel(ctx *cli.Context) error {
	offerID := ctx.String("offer-id")

	c := newRRPClient(ctx)
	resp, err := c.Cancel(offerID)
	if err != nil {
		return err
	}

	fmt.Printf("Cancelled successfully, exit status: %s\n", resp)
	return nil
}

func runClearOffers(ctx *cli.Context) error {
	c := newRRPClient(ctx)

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
	offerID := ctx.String("offer-id")

	c := newRRPClient(ctx)
	resp, err := c.GetStage(offerID)
	if err != nil {
		return err
	}

	fmt.Printf("Stage=%s: %s\n", resp.Stage, resp.Info)
	return nil
}

func runSetSwapTimeout(ctx *cli.Context) error {
	duration := ctx.Uint("duration")
	if duration == 0 {
		return errNoDuration
	}

	c := newRRPClient(ctx)
	err := c.SetSwapTimeout(uint64(duration))
	if err != nil {
		return err
	}

	fmt.Printf("Set timeout duration to %ds\n", duration)
	return nil
}
