package main

import (
	"os"

	logging "github.com/ipfs/go-log"
	"github.com/urfave/cli"
)

var log = logging.Logger("cmd")

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

func runClient(ctx *cli.Context) error {
	return nil
}