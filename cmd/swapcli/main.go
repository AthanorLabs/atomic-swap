// Package main provides the entrypoint of swapcli, an executable for interacting with a
// local swapd instance from the command line.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/skip2/go-qrcode"
	"github.com/urfave/cli/v2"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/relayer"
	"github.com/athanorlabs/atomic-swap/rpcclient"
	"github.com/athanorlabs/atomic-swap/rpcclient/wsclient"
)

const (
	defaultDiscoverSearchTimeSecs = 12

	flagSwapdPort      = "swapd-port"
	flagMinAmount      = "min-amount"
	flagMaxAmount      = "max-amount"
	flagPeerID         = "peer-id"
	flagOfferID        = "offer-id"
	flagOfferIDs       = "offer-ids"
	flagExchangeRate   = "exchange-rate"
	flagProvides       = "provides"
	flagProvidesAmount = "provides-amount"
	flagRelayerFee     = "relayer-fee"
	flagSearchTime     = "search-time"
	flagSubscribe      = "subscribe"
)

var (
	minRelayerFee = coins.NewWeiAmount(relayer.DefaultRelayerFee).AsEther()
	maxRelayerFee = apd.New(1, 0) // 1 ETH

	app = &cli.App{
		Name:                 "swapcli",
		Usage:                "Client for swapd",
		Version:              cliutil.GetVersion(),
		EnableBashCompletion: true,
		Suggest:              true,
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
				Name:    "peers",
				Aliases: []string{"p"},
				Usage:   "List peers that are currently connected",
				Action:  runPeers,
				Flags: []cli.Flag{
					swapdPortFlag,
				},
			},
			{
				Name:    "balances",
				Aliases: []string{"b"},
				Usage:   "Show our monero and ethereum account balances",
				Action:  runBalances,
				Flags: []cli.Flag{
					swapdPortFlag,
				},
			},
			{
				Name:   "eth-address",
				Usage:  "Show our ethereum address with its QR code",
				Action: runETHAddress,
				Flags: []cli.Flag{
					swapdPortFlag,
				},
			},
			{
				Name:   "xmr-address",
				Usage:  "Show our Monero address with its QR code",
				Action: runXMRAddress,
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
						Name: flagProvides,
						Usage: fmt.Sprintf("Coin to find providers for: one of [%s, %s]",
							coins.ProvidesXMR, coins.ProvidesETH),
						Value: string(coins.ProvidesXMR),
					},
					&cli.Uint64Flag{
						Name:  flagSearchTime,
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
						Name:     flagPeerID,
						Usage:    "Peer's ID, as provided by discover",
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
						Name: flagProvides,
						Usage: fmt.Sprintf("Coin to find providers for: one of [%s, %s]",
							coins.ProvidesXMR, coins.ProvidesETH),
						Value: string(coins.ProvidesXMR),
					},
					&cli.Uint64Flag{
						Name:  flagSearchTime,
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
					&cli.StringFlag{
						Name:     flagMinAmount,
						Usage:    "Minimum amount to be swapped, in XMR",
						Required: true,
					},
					&cli.StringFlag{
						Name:     flagMaxAmount,
						Usage:    "Maximum amount to be swapped, in XMR",
						Required: true,
					},
					&cli.StringFlag{
						Name:     flagExchangeRate,
						Usage:    "Desired exchange rate of XMR:ETH, eg. --exchange-rate=0.1 means 10XMR = 1ETH",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  flagSubscribe,
						Usage: "Subscribe to push notifications about the swap's status",
					},
					&cli.StringFlag{
						Name:  "eth-asset",
						Usage: "Ethereum ERC-20 token address to receive, or the zero address for regular ETH",
					},
					&cli.StringFlag{
						Name: flagRelayerFee,
						Usage: "Fee to pay the relayer in ETH if you have insufficient funds to claim:" +
							" eg. --relayer-fee=0.009 to pay 0.0009 ETH",
						Value: minRelayerFee.Text('f'),
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
						Name:     flagPeerID,
						Usage:    "Peer's ID, as provided by discover",
						Required: true,
					},
					&cli.StringFlag{
						Name:     flagOfferID,
						Usage:    "ID of the offer being taken",
						Required: true,
					},
					&cli.StringFlag{
						Name:     flagProvidesAmount,
						Usage:    "Amount of coin to send in the swap",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  flagSubscribe,
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
				Usage:  "Get information about ongoing swap(s).",
				Action: runGetOngoingSwap,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  flagOfferID,
						Usage: "ID of swap to retrieve info for",
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
						Name:     flagOfferID,
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
						Name:     flagOfferID,
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
						Name:  flagOfferID,
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
						Name:  flagOfferIDs,
						Usage: "A comma-separated list of offer IDs to delete",
					},
					swapdPortFlag,
				},
			},
			{
				Name:   "get-offers",
				Usage:  "Get all current offers.",
				Action: runGetOffers,
				Flags: []cli.Flag{
					swapdPortFlag,
				},
			},
			{
				Name:   "get-status",
				Usage:  "Get the status of a current swap.",
				Action: runGetStatus,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     flagOfferID,
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
			{
				Name:   "suggested-exchange-rate",
				Usage:  "Returns the current mainnet exchange rate based on ETH/USD and XMR/USD price feeds.",
				Action: runSuggestedExchangeRate,
				Flags:  []cli.Flag{swapdPortFlag},
			},
			{
				Name:   "get-swap-timeout",
				Usage:  "Get the duration between swap initiation and t0 and t0 and t1, in seconds",
				Action: runGetSwapTimeout,
				Flags: []cli.Flag{
					swapdPortFlag,
				},
			},
		},
	}

	swapdPortFlag = &cli.UintFlag{
		Name:    flagSwapdPort,
		Aliases: []string{"p"},
		Usage:   "RPC port of swap daemon",
		Value:   common.DefaultSwapdPort,
		EnvVars: []string{"SWAPD_PORT"},
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
	return rpcclient.NewClient(ctx.Context, endpoint)
}

func newWSClient(ctx *cli.Context) (wsclient.WsClient, error) {
	swapdPort := ctx.Uint(flagSwapdPort)
	endpoint := fmt.Sprintf("ws://127.0.0.1:%d/ws", swapdPort)
	return wsclient.NewWsClient(ctx.Context, endpoint)
}

func runAddresses(ctx *cli.Context) error {
	c := newRRPClient(ctx)
	resp, err := c.Addresses()
	if err != nil {
		return err
	}

	fmt.Println("Local listening multi-addresses:")
	for i, a := range resp.Addrs {
		fmt.Printf("%d: %s\n", i+1, a)
	}
	if len(resp.Addrs) == 0 {
		fmt.Println("[none]")
	}
	return nil
}

func runPeers(ctx *cli.Context) error {
	c := newRRPClient(ctx)
	resp, err := c.Peers()
	if err != nil {
		return err
	}

	fmt.Println("Connected peer multi-addresses:")
	for i, a := range resp.Addrs {
		fmt.Printf("%d: %s\n", i+1, a)
	}
	if len(resp.Addrs) == 0 {
		fmt.Println("[none]")
	}
	return nil
}

func runBalances(ctx *cli.Context) error {
	c := newRRPClient(ctx)
	balances, err := c.Balances()
	if err != nil {
		return err
	}

	fmt.Printf("Ethereum address: %s\n", balances.EthAddress)
	fmt.Printf("ETH Balance: %s\n", balances.WeiBalance.AsEther().Text('f'))
	fmt.Println()
	fmt.Printf("Monero address: %s\n", balances.MoneroAddress)
	fmt.Printf("XMR Balance: %s\n", balances.PiconeroBalance.AsMoneroString())
	fmt.Printf("Unlocked XMR balance: %s\n",
		balances.PiconeroUnlockedBalance.AsMoneroString())
	fmt.Printf("Blocks to unlock: %d\n", balances.BlocksToUnlock)
	return nil
}

func runETHAddress(ctx *cli.Context) error {
	c := newRRPClient(ctx)
	balances, err := c.Balances()
	if err != nil {
		return err
	}
	fmt.Printf("Ethereum address: %s\n", balances.EthAddress)
	code, err := qrcode.New(balances.EthAddress.String(), qrcode.Medium)
	if err != nil {
		return err
	}
	fmt.Println(code.ToString(false))
	return nil
}

func runXMRAddress(ctx *cli.Context) error {
	c := newRRPClient(ctx)
	balances, err := c.Balances()
	if err != nil {
		return err
	}
	fmt.Printf("Monero address: %s\n", balances.MoneroAddress)
	code, err := qrcode.New(balances.MoneroAddress.String(), qrcode.Medium)
	if err != nil {
		return err
	}
	fmt.Println(code.ToString(true))
	return nil
}

func runDiscover(ctx *cli.Context) error {
	c := newRRPClient(ctx)
	provides := ctx.String(flagProvides)
	peerIDs, err := c.Discover(provides, ctx.Uint64(flagSearchTime))
	if err != nil {
		return err
	}

	for i, peerID := range peerIDs {
		fmt.Printf("Peer %d: %v\n", i, peerID)
	}
	if len(peerIDs) == 0 {
		fmt.Println("[none]")
	}

	return nil
}

func runQuery(ctx *cli.Context) error {
	peerID, err := peer.Decode(ctx.String(flagPeerID))
	if err != nil {
		return errInvalidFlagValue(flagPeerID, err)
	}

	c := newRRPClient(ctx)
	res, err := c.Query(peerID)
	if err != nil {
		return err
	}

	for i, o := range res.Offers {
		err = printOffer(o, i, "")
		if err != nil {
			return err
		}
	}
	return nil
}

func runQueryAll(ctx *cli.Context) error {
	provides, err := providesStrToVal(ctx.String(flagProvides))
	if err != nil {
		return err
	}

	searchTime := ctx.Uint64(flagSearchTime)

	c := newRRPClient(ctx)
	peerOffers, err := c.QueryAll(provides, searchTime)
	if err != nil {
		return err
	}

	for i, po := range peerOffers {
		if i > 0 {
			fmt.Println("---")
		}
		fmt.Printf("Peer %d:\n", i)
		fmt.Printf("  Peer ID: %v\n", po.PeerID)
		fmt.Printf("  Offers:\n")
		for j, o := range po.Offers {
			err = printOffer(o, j, "    ")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func runMake(ctx *cli.Context) error {
	min, err := cliutil.ReadUnsignedDecimalFlag(ctx, flagMinAmount)
	if err != nil {
		return err
	}

	max, err := cliutil.ReadUnsignedDecimalFlag(ctx, flagMaxAmount)
	if err != nil {
		return err
	}

	exchangeRateDec, err := cliutil.ReadUnsignedDecimalFlag(ctx, flagExchangeRate)
	if err != nil {
		return err
	}
	exchangeRate := coins.ToExchangeRate(exchangeRateDec)
	// TODO: How to handle this if the other asset is not ETH?
	otherMin, err := exchangeRate.ToETH(min)
	if err != nil {
		return err
	}
	otherMax, err := exchangeRate.ToETH(max)
	if err != nil {
		return err
	}

	ethAssetStr := ctx.String("eth-asset")
	ethAsset := types.EthAssetETH
	if ethAssetStr != "" {
		ethAsset = types.EthAsset(ethcommon.HexToAddress(ethAssetStr))
	}

	c := newRRPClient(ctx)

	relayerFee, err := cliutil.ReadUnsignedDecimalFlag(ctx, flagRelayerFee)
	if err != nil {
		return err
	}
	if relayerFee.Cmp(minRelayerFee) < 0 || relayerFee.Cmp(maxRelayerFee) > 0 {
		return errRelayerFeeOutOfRange
	}

	printOfferSummary := func(offerResp *rpctypes.MakeOfferResponse) {
		fmt.Println("Published:")
		fmt.Printf("\tOffer ID:  %s\n", offerResp.OfferID)
		fmt.Printf("\tPeer ID:   %s\n", offerResp.PeerID)
		fmt.Printf("\tTaker Min: %s %s\n", otherMin.Text('f'), ethAsset)
		fmt.Printf("\tTaker Max: %s %s\n", otherMax.Text('f'), ethAsset)
	}

	if ctx.Bool(flagSubscribe) {
		wsc, err := newWSClient(ctx) //nolint:govet
		if err != nil {
			return err
		}
		defer wsc.Close()

		resp, statusCh, err := wsc.MakeOfferAndSubscribe(
			min,
			max,
			exchangeRate,
			ethAsset,
			relayerFee,
		)
		if err != nil {
			return err
		}

		printOfferSummary(resp)

		for stage := range statusCh {
			fmt.Printf("%s > Stage updated: %s\n", time.Now().Format(common.TimeFmtSecs), stage)
			if !stage.IsOngoing() {
				return nil
			}
		}

		return nil
	}

	resp, err := c.MakeOffer(min, max, exchangeRate, ethAsset, relayerFee)
	if err != nil {
		return err
	}

	printOfferSummary(resp)
	return nil
}

func runTake(ctx *cli.Context) error {
	peerID, err := peer.Decode(ctx.String(flagPeerID))
	if err != nil {
		return errInvalidFlagValue(flagPeerID, err)
	}
	offerID, err := types.HexToHash(ctx.String(flagOfferID))
	if err != nil {
		return errInvalidFlagValue(flagOfferID, err)
	}

	providesAmount, err := cliutil.ReadUnsignedDecimalFlag(ctx, flagProvidesAmount)
	if err != nil {
		return err
	}

	if ctx.Bool(flagSubscribe) {
		wsc, err := newWSClient(ctx)
		if err != nil {
			return err
		}
		defer wsc.Close()

		statusCh, err := wsc.TakeOfferAndSubscribe(peerID, offerID, providesAmount)
		if err != nil {
			return err
		}

		fmt.Printf("Initiated swap with offer ID %s\n", offerID)

		for stage := range statusCh {
			fmt.Printf("%s > Stage updated: %s\n", time.Now().Format(common.TimeFmtSecs), stage)
			if !stage.IsOngoing() {
				return nil
			}
		}

		return nil
	}

	c := newRRPClient(ctx)
	if err := c.TakeOffer(peerID, offerID, providesAmount); err != nil {
		return err
	}

	fmt.Printf("Initiated swap with offer ID %s\n", offerID)
	return nil
}

func runGetPastSwapIDs(ctx *cli.Context) error {
	c := newRRPClient(ctx)
	ids, err := c.GetPastSwapIDs()
	if err != nil {
		return err
	}

	fmt.Println("Past swap offer IDs:")
	if len(ids) == 0 {
		fmt.Println("[none]")
		return nil
	}

	for i, id := range ids {
		if i > 0 {
			fmt.Printf("---\n")
		}

		fmt.Printf("ID: %s\n", id.ID)
		fmt.Printf("Start time: %s\n", id.StartTime.Format(common.TimeFmtSecs))
		fmt.Printf("End time: %s\n", id.EndTime.Format(common.TimeFmtSecs))
	}

	return nil
}

func runGetOngoingSwap(ctx *cli.Context) error {
	offerID := ctx.String(flagOfferID)

	c := newRRPClient(ctx)
	resp, err := c.GetOngoingSwap(offerID)
	if err != nil {
		return err
	}

	fmt.Println("Ongoing swaps:")
	if len(resp.Swaps) == 0 {
		fmt.Println("[none]")
		return nil
	}

	for i, info := range resp.Swaps {
		if i > 0 {
			fmt.Printf("---\n")
		}

		receivedCoin := "ETH"
		if info.Provided == coins.ProvidesETH {
			receivedCoin = "XMR"
		}

		fmt.Printf("ID: %s\n", info.ID)
		fmt.Printf("Start time: %s\n", info.StartTime.Format(common.TimeFmtSecs))
		fmt.Printf("Provided: %s %s\n", info.ProvidedAmount.Text('f'), info.Provided)
		fmt.Printf("Receiving: %s %s\n", info.ExpectedAmount.Text('f'), receivedCoin)
		fmt.Printf("Exchange Rate: %s ETH/XMR\n", info.ExchangeRate)
		fmt.Printf("Status: %s\n", info.Status)
		if info.Timeout0 != nil && info.Timeout1 != nil {
			fmt.Printf("First timeout: %s\n", info.Timeout0.Format(common.TimeFmtSecs))
			fmt.Printf("Second timeout: %s\n", info.Timeout1.Format(common.TimeFmtSecs))
		}
	}

	return nil
}

func runGetPastSwap(ctx *cli.Context) error {
	offerID := ctx.String(flagOfferID)

	c := newRRPClient(ctx)
	info, err := c.GetPastSwap(offerID)
	if err != nil {
		return err
	}

	receivedCoin := "ETH"
	if info.Provided == coins.ProvidesETH {
		receivedCoin = "XMR"
	}

	fmt.Printf("Start time: %s\n", info.StartTime.Format(common.TimeFmtSecs))
	fmt.Printf("End time: %s\n", info.EndTime.Format(common.TimeFmtSecs))
	fmt.Printf("Provided: %s %s\n", info.ProvidedAmount.Text('f'), info.Provided)
	fmt.Printf("Receiving: %s %s\n", info.ExpectedAmount.Text('f'), receivedCoin)
	fmt.Printf("Exchange Rate: %s ETH/XMR\n", info.ExchangeRate)
	fmt.Printf("Status: %s\n", info.Status)

	return nil
}

func runRefund(ctx *cli.Context) error {
	offerID := ctx.String(flagOfferID)

	c := newRRPClient(ctx)
	resp, err := c.Refund(offerID)
	if err != nil {
		return err
	}

	fmt.Printf("Refunded successfully, transaction hash: %s\n", resp.TxHash)
	return nil
}

func runCancel(ctx *cli.Context) error {
	offerID, err := types.HexToHash(ctx.String(flagOfferID))
	if err != nil {
		return errInvalidFlagValue(flagOfferID, err)
	}

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

	ids := ctx.String(flagOfferIDs)
	if ids == "" {
		err := c.ClearOffers(nil)
		if err != nil {
			return err
		}

		fmt.Printf("Cleared all offers successfully.\n")
		return nil
	}

	var offerIDs []types.Hash
	for _, offerIDStr := range strings.Split(ids, ",") {
		id, err := types.HexToHash(strings.TrimSpace(offerIDStr))
		if err != nil {
			return errInvalidFlagValue(flagOfferIDs, err)
		}
		offerIDs = append(offerIDs, id)
	}
	err := c.ClearOffers(offerIDs)
	if err != nil {
		return err
	}

	fmt.Printf("Cleared offers successfully: %s\n", ids)
	return nil
}

func runGetOffers(ctx *cli.Context) error {
	c := newRRPClient(ctx)
	resp, err := c.GetOffers()
	if err != nil {
		return err
	}

	fmt.Println("Peer ID (self):", resp.PeerID)
	fmt.Println("Offers:")
	for i, offer := range resp.Offers {
		err = printOffer(offer, i, "  ")
		if err != nil {
			return err
		}
	}
	if len(resp.Offers) == 0 {
		fmt.Println("[no offers]")
	}

	return nil
}

func runGetStatus(ctx *cli.Context) error {
	offerID, err := types.HexToHash(ctx.String(flagOfferID))
	if err != nil {
		return errInvalidFlagValue(flagOfferID, err)
	}

	c := newRRPClient(ctx)
	resp, err := c.GetStatus(offerID)
	if err != nil {
		return err
	}

	fmt.Printf("Start time: %s\n", resp.StartTime.Format(common.TimeFmtSecs))
	fmt.Printf("Status=%s: %s\n", resp.Status, resp.Description)
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

	fmt.Printf("Set timeout duration to %d seconds\n", duration)
	return nil
}

func runGetSwapTimeout(ctx *cli.Context) error {
	c := newRRPClient(ctx)
	resp, err := c.GetSwapTimeout()
	if err != nil {
		return err
	}

	fmt.Printf("Swap timeout duration: %d seconds\n", resp.Timeout)
	return nil
}

func runSuggestedExchangeRate(ctx *cli.Context) error {
	c := newRRPClient(ctx)
	resp, err := c.SuggestedExchangeRate()
	if err != nil {
		return err
	}

	fmt.Printf("Exchange rate: %s\n", resp.ExchangeRate)
	fmt.Printf("XMR/USD Price: %-13s (%s)\n", resp.XMRPrice, resp.XMRUpdatedAt)
	fmt.Printf("ETH/USD Price: %-13s (%s)\n", resp.ETHPrice, resp.ETHUpdatedAt)

	return nil
}

func printOffer(o *types.Offer, index int, indent string) error {
	if index > 0 {
		fmt.Printf("%s---\n", indent)
	}

	xRate := o.ExchangeRate
	minETH, err := xRate.ToETH(o.MinAmount)
	if err != nil {
		return err
	}
	maxETH, err := xRate.ToETH(o.MaxAmount)
	if err != nil {
		return err
	}

	fmt.Printf("%sOffer ID: %s\n", indent, o.ID)
	fmt.Printf("%sProvides: %s\n", indent, o.Provides)
	fmt.Printf("%sTakes: %s\n", indent, o.EthAsset)
	fmt.Printf("%sExchange Rate: %s %s/%s\n", indent, o.ExchangeRate, o.EthAsset, o.Provides)
	fmt.Printf("%sMaker Min: %s %s\n", indent, o.MinAmount.Text('f'), o.Provides)
	fmt.Printf("%sMaker Max: %s %s\n", indent, o.MaxAmount.Text('f'), o.Provides)
	fmt.Printf("%sTaker Min: %s %s\n", indent, minETH.Text('f'), o.EthAsset)
	fmt.Printf("%sTaker Max: %s %s\n", indent, maxETH.Text('f'), o.EthAsset)
	return nil
}

func providesStrToVal(providesStr string) (coins.ProvidesCoin, error) {
	var provides coins.ProvidesCoin

	// The provides flag value defaults to XMR, but the user can still specify the empty
	// string explicitly, which they can do to search the empty DHT namespace for all
	// peers. `NewProvidesCoin` gives an error if you pass the empty string, so we
	// special case the empty string.
	if providesStr == "" {
		return provides, nil
	}
	return coins.NewProvidesCoin(providesStr)
}
