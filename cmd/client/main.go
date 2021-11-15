package main

import (
	"fmt"
	"os"
	"time"

	"github.com/noot/atomic-swap/net"

	"github.com/gdamore/tcell/v2"
	logging "github.com/ipfs/go-log"
	"github.com/rivo/tview"
	"github.com/urfave/cli"
)

var log = logging.Logger("cmd")

const (
	defaultDaemonAddress = "http://localhost:5001"
)

var (
	app = &cli.App{
		Name:   "swapcli",
		Usage:  "Client for swapd",
		Action: runClient,
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:  "daemon-addr",
				Usage: "address of swap daemon; default http://localhost:5001",
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

type listHandlerFunc func(index int, mainText string, secondaryText string, shortcut rune)

func runClient(c *cli.Context) error {
	app := tview.NewApplication()
	list := tview.NewList().
		AddItem("Query", "Some explanatory text", 'a', func() {
			queryFunc(app)
		}).
		AddItem("List item 2", "Some explanatory text", 'b', nil).
		AddItem("List item 3", "Some explanatory text", 'c', nil).
		AddItem("List item 4", "Some explanatory text", 'd', nil).
		AddItem("Quit", "Press to exit", 'q', func() {
			app.Stop()
		})
	if err := app.SetRoot(list, true).SetFocus(list).Run(); err != nil {
		panic(err)
	}

	return nil
}

func queryFunc(app *tview.Application) {
	queryResultPrimitive := tview.NewBox()

	pages := tview.NewPages()
	pages = pages.AddAndSwitchToPage(
		"query",
		tview.NewModal().SetText("coin to query network and find providers for").
			AddButtons([]string{net.ProvidesETH, net.ProvidesXMR}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {

				pages.HidePage("query").AddAndSwitchToPage("queryResult", queryResultPrimitive, false)

				app.SetAfterDrawFunc(func(screen tcell.Screen) {
					tview.PrintSimple(screen, "querying daemon for peers...", 0, 0)
					go queryDaemon(net.ProvidesCoin(buttonLabel), screen)
				})

			}),
		false,
	)
	app.SetRoot(pages, true).SetFocus(pages)
}

func queryDaemon(coin net.ProvidesCoin, screen tcell.Screen) {
	i := 0
	for {
		time.Sleep(time.Second)
		tview.PrintSimple(screen, fmt.Sprintf("%d", i), 1, i+1)
		i++

		if i == 10 {
			return
		}
	}
}
