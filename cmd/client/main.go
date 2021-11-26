package main

import (
	"os"

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
				Name:    "discover",
				Aliases: []string{"d"},
				Usage:   "discover peers who provide a certain coin",
				Action:  runDiscover,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "provides",
						Usage: "coin to find providers for: one of [ETH, XMR]",
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
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

func runDiscover(ctx *cli.Context) error {
	return nil
}

func runQuery(ctx *cli.Context) error {
	return nil
}

func runInitiate(ctx *cli.Context) error {
	return nil
}
