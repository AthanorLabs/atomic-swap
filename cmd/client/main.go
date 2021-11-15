package main

import (
	"os"

	"github.com/urfave/cli"

	logging "github.com/ipfs/go-log"
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

func runClient(c *cli.Context) error {
	return nil
}
