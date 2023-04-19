// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package main provides the entrypoint of the bootnode executable,
// a node that is only used to bootstrap the p2p network and does not run
// any swap services.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/rpc"

	"github.com/hashicorp/go-multierror"
	logging "github.com/ipfs/go-log"
	"github.com/urfave/cli/v2"
)

const (
	defaultLibp2pPort = 9900
	defaultRPCPort    = common.DefaultSwapdPort

	flagDataDir    = "data-dir"
	flagLibp2pKey  = "libp2p-key"
	flagLibp2pPort = "libp2p-port"
	flagBootnodes  = "bootnodes"
	flagRPCPort    = "rpc-port"
	flagEnv        = "env"
)

var log = logging.Logger("cmd")

func cliApp() *cli.App {
	return &cli.App{
		Name:                 "bootnode",
		Usage:                "A bootnode for the atomic swap p2p network.",
		Version:              cliutil.GetVersion(),
		Action:               runBootnode,
		EnableBashCompletion: true,
		Suggest:              true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagDataDir,
				Usage: "Path to store swap artifacts",
				Value: "{HOME}/.atomicswap/{ENV}", // For --help only, actual default replaces variables
			},
			&cli.StringFlag{
				Name:  flagLibp2pKey,
				Usage: "libp2p private key",
				Value: fmt.Sprintf("{DATA_DIR}/%s", common.DefaultLibp2pKeyFileName),
			},
			&cli.UintFlag{
				Name:  flagLibp2pPort,
				Usage: "libp2p port to listen on",
				Value: defaultLibp2pPort,
			},
			&cli.StringSliceFlag{
				Name:    flagBootnodes,
				Aliases: []string{"bn"},
				Usage:   "libp2p bootnode, comma separated if passing multiple to a single flag",
				EnvVars: []string{"SWAPD_BOOTNODES"},
			},
			&cli.UintFlag{
				Name:  flagRPCPort,
				Usage: "Port for the bootnode RPC server to run on",
				Value: defaultRPCPort,
			},
			&cli.StringFlag{
				Name:  flagEnv,
				Usage: "Environment to use: one of mainnet, stagenet, or dev",
				Value: "dev",
			},
			&cli.StringFlag{
				Name:  cliutil.FlagLogLevel,
				Usage: "Set log level: one of [error|warn|info|debug]",
				Value: "info",
			},
		},
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go signalHandler(ctx, cancel)

	err := cliApp().RunContext(ctx, os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func runBootnode(c *cli.Context) error {
	// Fail if any non-flag arguments were passed
	if c.Args().Present() {
		return fmt.Errorf("unknown command %q", c.Args().First())
	}

	if err := cliutil.SetLogLevelsFromContext(c); err != nil {
		return err
	}

	config, err := getEnvConfig(c)
	if err != nil {
		return err
	}

	libp2pKeyFile := config.LibP2PKeyFile()
	if c.IsSet(flagLibp2pKey) {
		libp2pKeyFile = c.String(flagLibp2pKey)
		if libp2pKeyFile == "" {
			return errFlagValueEmpty(flagLibp2pKey)
		}
	}

	if libp2pKeyFile == "" {
		libp2pKeyFile = path.Join(config.DataDir, common.DefaultLibp2pKeyFileName)
	}

	libp2pPort := uint16(c.Uint(flagLibp2pPort))

	hostListenIP := "0.0.0.0"
	if config.Env == common.Development {
		hostListenIP = "127.0.0.1"
	}

	host, err := net.NewHost(&net.Config{
		Ctx:        c.Context,
		DataDir:    config.DataDir,
		Port:       libp2pPort,
		KeyFile:    libp2pKeyFile,
		Bootnodes:  config.Bootnodes,
		ProtocolID: fmt.Sprintf("%s/%d", net.ProtocolID, config.EthereumChainID.Int64()),
		ListenIP:   hostListenIP,
		IsRelayer:  false,
		IsBootnode: true,
	})
	if err != nil {
		return err
	}
	defer func() {
		if hostErr := host.Stop(); hostErr != nil {
			err = multierror.Append(err, fmt.Errorf("error shutting down peer-to-peer services: %w", hostErr))
		}
	}()

	if err = host.Start(); err != nil {
		return err
	}

	rpcPort := c.Uint(flagRPCPort)
	rpcServer, err := rpc.NewServer(&rpc.Config{
		Ctx:        c.Context,
		Address:    fmt.Sprintf("127.0.0.1:%d", rpcPort),
		Net:        host,
		Namespaces: map[string]struct{}{rpc.DaemonNamespace: {}, rpc.NetNamespace: {}},
	})

	log.Infof("starting bootnode with data-dir %s", config.DataDir)
	err = rpcServer.Start()

	if errors.Is(err, http.ErrServerClosed) {
		// Remove the error for a clean program exit, as ErrServerClosed only
		// happens when the server is told to shut down
		err = nil
	}

	// err can get set in defer blocks, so return err or use an empty
	// return statement below (not nil)
	return err
}

func getEnvConfig(c *cli.Context) (*common.Config, error) {
	env, err := common.NewEnv(c.String(flagEnv))
	if err != nil {
		return nil, err
	}
	conf := common.ConfigDefaultsForEnv(env)

	// cfg.DataDir already has a default set, so only override if the user explicitly set the flag
	if c.IsSet(flagDataDir) {
		conf.DataDir = c.String(flagDataDir) // override the value derived from `flagEnv`
		if conf.DataDir == "" {
			return nil, errFlagValueEmpty(flagDataDir)
		}
	}

	if err = common.MakeDir(conf.DataDir); err != nil {
		return nil, err
	}

	if c.IsSet(flagBootnodes) {
		conf.Bootnodes = cliutil.ExpandBootnodes(c.StringSlice(flagBootnodes))
	}

	return conf, nil
}

func errFlagValueEmpty(flag string) error {
	return fmt.Errorf("flag %q requires a non-empty value", flag)
}
