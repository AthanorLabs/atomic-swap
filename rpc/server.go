// Copyright 2023 Athanor Labs (ON)
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
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p/core/peer"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/txsender"
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
	Address         string // "IP:port"
	Net             Net
	XMRTaker        XMRTaker
	XMRMaker        XMRMaker
	ProtocolBackend ProtocolBackend
}

// NewServer ...
func NewServer(cfg *Config) (*Server, error) {
	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(NewCodec(), "application/json")

	ns := NewNetService(cfg.Net, cfg.XMRTaker, cfg.XMRMaker, cfg.ProtocolBackend.SwapManager())
	if err := rpcServer.RegisterService(ns, "net"); err != nil {
		return nil, err
	}

	err := rpcServer.RegisterService(NewPersonalService(cfg.Ctx, cfg.XMRMaker, cfg.ProtocolBackend), "personal")
	if err != nil {
		return nil, err
	}

	swapService := NewSwapService(
		cfg.Ctx,
		cfg.ProtocolBackend.SwapManager(),
		cfg.XMRTaker,
		cfg.XMRMaker,
		cfg.Net,
		cfg.ProtocolBackend,
	)
	if err = rpcServer.RegisterService(swapService, "swap"); err != nil {
		return nil, err
	}

	wsServer := newWsServer(cfg.Ctx, cfg.ProtocolBackend.SwapManager(), ns, cfg.ProtocolBackend, cfg.XMRTaker)

	lc := net.ListenConfig{}
	ln, err := lc.Listen(cfg.Ctx, "tcp", cfg.Address)
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()
	r.Handle("/", rpcServer)
	r.Handle("/ws", wsServer)

	headersOk := handlers.AllowedHeaders([]string{"content-type", "username", "password"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	server := &http.Server{
		Addr:              ln.Addr().String(),
		ReadHeaderTimeout: time.Second,
		Handler:           handlers.CORS(headersOk, methodsOk, originsOk)(r),
		BaseContext: func(listener net.Listener) context.Context {
			return cfg.Ctx
		},
	}

	s := &Server{
		ctx:        cfg.Ctx,
		listener:   ln,
		httpServer: server,
	}

	if err = rpcServer.RegisterService(NewDaemonService(s, cfg), "daemon"); err != nil {
		return nil, err
	}

	return s, nil
}

// HttpURL returns the URL used for HTTP requests
func (s *Server) HttpURL() string { //nolint:revive
	return fmt.Sprintf("http://%s", s.httpServer.Addr)
}

// WsURL returns the URL used for websocket requests
func (s *Server) WsURL() string {
	return fmt.Sprintf("ws://%s/ws", s.httpServer.Addr)
}

// Start starts the JSON-RPC and Websocket server.
func (s *Server) Start() error {
	if s.ctx.Err() != nil {
		return s.ctx.Err()
	}

	log.Infof("Starting RPC server on %s", s.HttpURL())
	log.Infof("Starting websockets server on %s", s.WsURL())

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
		if err := s.httpServer.Shutdown(s.ctx); err != nil {
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
	Env() common.Environment
	SetSwapTimeout(timeout time.Duration)
	SwapTimeout() time.Duration
	SwapManager() swap.Manager
	SwapCreatorAddr() ethcommon.Address
	SetXMRDepositAddress(*mcrypto.Address, types.Hash)
	ClearXMRDepositAddress(types.Hash)
	ETHClient() extethclient.EthClient
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

// SwapManager ...
type SwapManager = swap.Manager
