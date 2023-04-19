// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package daemon is responsible for assembling, running and cleanly shutting
// down the swap daemon (swapd) and its numerous subcomponents.
package daemon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/ChainSafe/chaindb"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-multierror"
	logging "github.com/ipfs/go-log"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/db"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/xmrmaker"
	"github.com/athanorlabs/atomic-swap/protocol/xmrtaker"
	"github.com/athanorlabs/atomic-swap/rpc"
)

var log = logging.Logger("daemon")

// SwapdConfig provides startup parameters for swapd.
type SwapdConfig struct {
	EnvConf        *common.Config
	MoneroClient   monero.WalletClient
	EthereumClient extethclient.EthClient
	Libp2pPort     uint16
	Libp2pKeyfile  string
	RPCPort        uint16
	IsRelayer      bool
	NoTransferBack bool
}

// RunSwapDaemon assembles and runs a swapd instance blocking until swapd is
// shut down. Typically, shutdown happens because a signal handler cancels the
// passed in context, or when the shutdown RPC method is called.
func RunSwapDaemon(ctx context.Context, conf *SwapdConfig) (err error) {
	// Note: err can be modified in defer blocks, so it needs to be a named return
	//       value above.
	if conf.Libp2pKeyfile == "" {
		conf.Libp2pKeyfile = path.Join(conf.EnvConf.DataDir, common.DefaultLibp2pKeyFileName)
	}

	if conf.EnvConf.SwapCreatorAddr == (ethcommon.Address{}) {
		panic("swap creator address not specified")
	}

	ec := conf.EthereumClient
	chainID := ec.ChainID()

	// Initialize the database first, so the defer statement that closes it
	// will get executed last.
	sdb, err := db.NewDatabase(&chaindb.Config{
		DataDir: path.Join(conf.EnvConf.DataDir, "db"),
	})
	if err != nil {
		return err
	}
	defer func() {
		if dbErr := sdb.Close(); dbErr != nil {
			err = multierror.Append(err, fmt.Errorf("syncing database: %s", dbErr))
		}
	}()

	sm, err := swap.NewManager(sdb)
	if err != nil {
		return err
	}

	hostListenIP := "0.0.0.0"
	if conf.EnvConf.Env == common.Development {
		hostListenIP = "127.0.0.1"
	}

	host, err := net.NewHost(&net.Config{
		Ctx:        ctx,
		DataDir:    conf.EnvConf.DataDir,
		Port:       conf.Libp2pPort,
		KeyFile:    conf.Libp2pKeyfile,
		Bootnodes:  conf.EnvConf.Bootnodes,
		ProtocolID: fmt.Sprintf("%s/%d", net.ProtocolID, chainID.Int64()),
		ListenIP:   hostListenIP,
		IsRelayer:  conf.IsRelayer,
	})
	if err != nil {
		return err
	}
	defer func() {
		if hostErr := host.Stop(); hostErr != nil {
			err = multierror.Append(err, fmt.Errorf("error shutting down peer-to-peer services: %w", hostErr))
		}
	}()

	swapBackend, err := backend.NewBackend(&backend.Config{
		Ctx:             ctx,
		MoneroClient:    conf.MoneroClient,
		EthereumClient:  conf.EthereumClient,
		Environment:     conf.EnvConf.Env,
		SwapCreatorAddr: conf.EnvConf.SwapCreatorAddr,
		SwapManager:     sm,
		RecoveryDB:      sdb.RecoveryDB(),
		Net:             host,
	})
	if err != nil {
		return fmt.Errorf("failed to make backend: %w", err)
	}

	log.Infof("created backend with monero endpoint %s and ethereum endpoint %s",
		swapBackend.XMRClient().Endpoint(),
		conf.EthereumClient.Endpoint(),
	)

	xmrTaker, err := xmrtaker.NewInstance(&xmrtaker.Config{
		Backend:        swapBackend,
		DataDir:        conf.EnvConf.DataDir,
		NoTransferBack: conf.NoTransferBack,
	})
	if err != nil {
		return err
	}

	xmrMaker, err := xmrmaker.NewInstance(&xmrmaker.Config{
		Backend:  swapBackend,
		DataDir:  conf.EnvConf.DataDir,
		Database: sdb,
		Network:  host,
	})
	if err != nil {
		return err
	}

	// connect the maker/taker handlers to the p2p network host
	host.SetHandlers(xmrMaker, swapBackend)
	if err = host.Start(); err != nil {
		return err
	}

	rpcServer, err := rpc.NewServer(&rpc.Config{
		Ctx:             ctx,
		Address:         fmt.Sprintf("127.0.0.1:%d", conf.RPCPort),
		Net:             host,
		XMRTaker:        xmrTaker,
		XMRMaker:        xmrMaker,
		ProtocolBackend: swapBackend,
		RecoveryDB:      sdb.RecoveryDB(),
	})

	log.Infof("starting swapd with data-dir %s", conf.EnvConf.DataDir)
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
