// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package bootnode is responsible for assembling, running and cleanly shutting
// down a swap bootnode.
package bootnode

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"

	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/rpc"

	"github.com/hashicorp/go-multierror"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("bootnode")

// Config provides the configuration for a bootnode.
type Config struct {
	DataDir         string
	Bootnodes       []string
	HostListenIP    string
	Libp2pPort      uint16
	Libp2pKeyFile   string
	RPCPort         uint16
	EthereumChainID *big.Int
}

// RunBootnode assembles and runs a bootnode instance, blocking until the node is
// shut down. Typically, shutdown happens because a signal handler cancels the
// passed in context, or when the shutdown RPC method is called.
func RunBootnode(ctx context.Context, cfg *Config) error {
	host, err := net.NewHost(&net.Config{
		Ctx:            ctx,
		DataDir:        cfg.DataDir,
		Port:           cfg.Libp2pPort,
		KeyFile:        cfg.Libp2pKeyFile,
		Bootnodes:      cfg.Bootnodes,
		ProtocolID:     fmt.Sprintf("%s/%d", net.ProtocolID, cfg.EthereumChainID),
		ListenIP:       cfg.HostListenIP,
		IsRelayer:      false,
		IsBootnodeOnly: true,
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

	rpcServer, err := rpc.NewServer(&rpc.Config{
		Ctx:     ctx,
		Address: fmt.Sprintf("127.0.0.1:%d", cfg.RPCPort),
		Net:     host,
		Namespaces: map[string]struct{}{
			rpc.DaemonNamespace: {},
			rpc.NetNamespace:    {},
		},
		IsBootnodeOnly: true,
	})
	if err != nil {
		return err
	}

	log.Infof("starting bootnode with data-dir %s", cfg.DataDir)
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
