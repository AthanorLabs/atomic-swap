// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package rpc provides the HTTP server for incoming JSON-RPC and websocket requests to
// swapd from the local host. The answers to these queries come from 3 subsystems: net,
// personal and swap.
package rpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/MarinX/monerorpc/wallet"
	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/txsender"
)

const (
	DaemonNamespace   = "daemon"   //nolint:revive
	DatabaseNamespace = "database" //nolint:revive
	NetNamespace      = "net"      //nolint:revive
	PersonalName      = "personal" //nolint:revive
	SwapNamespace     = "swap"     //nolint:revive
)

var log = logging.Logger("rpc")

// Server represents the JSON-RPC server
type Server struct {
	ctx        context.Context
	listener   net.Listener
	httpServer *http.Server
}

// Config ...
type Config struct {
	Ctx             context.Context
	Env             common.Environment
	Address         string // "IP:port"
	Net             Net
	XMRTaker        XMRTaker        // nil on bootnodes
	XMRMaker        XMRMaker        // nil on bootnodes
	ProtocolBackend ProtocolBackend // nil on bootnodes
	RecoveryDB      RecoveryDB      // nil on bootnodes
	Namespaces      map[string]struct{}
	IsBootnodeOnly  bool
}

// AllNamespaces returns a map with all RPC namespaces set for usage in the config.
func AllNamespaces() map[string]struct{} {
	return map[string]struct{}{
		DaemonNamespace:   {},
		DatabaseNamespace: {},
		NetNamespace:      {},
		PersonalName:      {},
		SwapNamespace:     {},
	}
}

// NewServer ...
func NewServer(cfg *Config) (*Server, error) {
	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(NewCodec(), "application/json")

	serverCtx, serverCancel := context.WithCancel(cfg.Ctx)
	var swapCreatorAddr *ethcommon.Address
	if !cfg.IsBootnodeOnly {
		addr := cfg.ProtocolBackend.SwapCreatorAddr()
		swapCreatorAddr = &addr
	}
	daemonService := NewDaemonService(serverCancel, cfg.Env, swapCreatorAddr)
	err := rpcServer.RegisterService(daemonService, "daemon")
	if err != nil {
		return nil, err
	}

	var swapManager swap.Manager
	if !cfg.IsBootnodeOnly {
		swapManager = cfg.ProtocolBackend.SwapManager()
	}

	var netService *NetService
	for ns := range cfg.Namespaces {
		switch ns {
		case DaemonNamespace:
			continue
		case DatabaseNamespace:
			err = rpcServer.RegisterService(NewDatabaseService(cfg.RecoveryDB), DatabaseNamespace)
		case NetNamespace:
			netService = NewNetService(cfg.Net, cfg.XMRTaker, cfg.XMRMaker, swapManager, cfg.IsBootnodeOnly)
			err = rpcServer.RegisterService(netService, NetNamespace)
		case PersonalName:
			err = rpcServer.RegisterService(NewPersonalService(serverCtx, cfg.XMRMaker, cfg.ProtocolBackend), PersonalName)
		case SwapNamespace:
			err = rpcServer.RegisterService(
				NewSwapService(
					serverCtx,
					swapManager,
					cfg.XMRTaker,
					cfg.XMRMaker,
					cfg.Net,
					cfg.ProtocolBackend,
					cfg.RecoveryDB,
				),
				SwapNamespace,
			)
		default:
			err = fmt.Errorf("unknown namespace %s", ns)
		}
	}
	if err != nil {
		serverCancel()
		return nil, err
	}

	wsServer := newWsServer(serverCtx, swapManager, netService, cfg.ProtocolBackend, cfg.XMRTaker)

	lc := net.ListenConfig{}
	ln, err := lc.Listen(serverCtx, "tcp", cfg.Address)
	if err != nil {
		serverCancel()
		return nil, err
	}

	reg, err := NewPrometheusRegistry()
	if err != nil {
		return nil, err
	}

	if !cfg.IsBootnodeOnly {
		SetupMetrics(serverCtx, reg, cfg.Net, cfg.ProtocolBackend, cfg.XMRMaker)
	}
	r := mux.NewRouter()
	r.Handle("/", rpcServer)
	r.Handle("/ws", wsServer)
	if !cfg.IsBootnodeOnly {
		r.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	}
	headersOk := handlers.AllowedHeaders([]string{"content-type", "username", "password"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	server := &http.Server{
		Addr:              ln.Addr().String(),
		ReadHeaderTimeout: time.Second,
		Handler:           handlers.CORS(headersOk, methodsOk, originsOk)(r),
		BaseContext: func(listener net.Listener) context.Context {
			return serverCtx
		},
	}

	return &Server{
		ctx:        serverCtx,
		listener:   ln,
		httpServer: server,
	}, nil
}

// Port returns the localhost port used for HTTP and websocket requests
func (s *Server) Port() uint16 {
	return uint16(s.listener.Addr().(*net.TCPAddr).Port)
}

// Start starts the JSON-RPC and Websocket server.
func (s *Server) Start() error {
	if s.ctx.Err() != nil {
		return s.ctx.Err()
	}

	log.Infof("Starting RPC/websockets server on 127.0.0.1:%d", s.Port())

	serverErr := make(chan error, 1)
	go func() {
		// Serve never returns nil. It returns http.ErrServerClosed if it was terminated
		// by the Shutdown.
		serverErr <- s.httpServer.Serve(s.listener)
	}()

	select {
	case <-s.ctx.Done():
		// Shutdown below is passed a closed context, which means it will shut down
		// immediately without servicing already connected clients.
		err := s.httpServer.Shutdown(s.ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Warnf("http server shutdown errored: %s", err)
		}
		// We shut down because the context was cancelled, so that's the error to return
		return s.ctx.Err()
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("RPC server failed: %s", err)
		} else {
			log.Info("RPC server shut down")
		}
		return err
	}
}

// Stop the JSON-RPC and websockets server. If server's context is not cancelled, a
// graceful shutdown happens where existing connections are serviced until disconnected.
// If the context is cancelled, the shutdown is immediate.
func (s *Server) Stop() error {
	return s.httpServer.Shutdown(s.ctx)
}

// Protocol represents the functions required by the rpc service into the protocol handler.
type Protocol interface {
	Provides() coins.ProvidesCoin
	GetOngoingSwapState(types.Hash) common.SwapState
}

// ProtocolBackend represents protocol/backend.Backend
type ProtocolBackend interface {
	Ctx() context.Context
	Env() common.Environment
	SetSwapTimeout(timeout time.Duration)
	SwapTimeout() time.Duration
	SwapManager() swap.Manager
	SwapCreatorAddr() ethcommon.Address
	SetXMRDepositAddress(*mcrypto.Address, types.Hash)
	ClearXMRDepositAddress(types.Hash)
	ETHClient() extethclient.EthClient
	TransferXMR(to *mcrypto.Address, amount *coins.PiconeroAmount) (string, error)
	SweepXMR(to *mcrypto.Address) ([]string, error)
	TransferETH(to ethcommon.Address, amount *coins.WeiAmount, gasLimit *uint64) (*ethtypes.Receipt, error)
	SweepETH(to ethcommon.Address) (*ethtypes.Receipt, error)
}

// XMRTaker ...
type XMRTaker interface {
	Protocol
	InitiateProtocol(peerID peer.ID, providesAmount *apd.Decimal, offer *types.Offer) (common.SwapState, error)
	ExternalSender(offerID types.Hash) (*txsender.ExternalSender, error)
}

// XMRMaker ...
type XMRMaker interface {
	Protocol
	MakeOffer(offer *types.Offer, useRelayer bool) (*types.OfferExtra, error)
	GetOffers() []*types.Offer
	ClearOffers([]types.Hash) error
	GetMoneroBalance() (*mcrypto.Address, *wallet.GetBalanceResponse, error)
}
