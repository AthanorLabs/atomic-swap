package rpc

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"time"

	"github.com/MarinX/monerorpc/wallet"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/txsender"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"

	logging "github.com/ipfs/go-log"
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

	if err := rpcServer.RegisterService(NewPersonalService(cfg.XMRMaker, cfg.ProtocolBackend), "personal"); err != nil {
		return nil, err
	}

	if err := rpcServer.RegisterService(NewSwapService(cfg.ProtocolBackend.SwapManager(), cfg.XMRTaker, cfg.XMRMaker, cfg.Net), "swap"); err != nil { //nolint:lll
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

	return &Server{
		ctx:        cfg.Ctx,
		listener:   ln,
		httpServer: server,
	}, nil
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
	log.Infof("Starting RPC server on %s", s.HttpURL())
	log.Infof("Starting websockets server on %s", s.WsURL())

	serverErr := make(chan error, 1)
	go func() {
		// Serve never returns nil. It returns http.ErrServerClosed if it was terminated
		// by the Shutdown.
		err := s.httpServer.Serve(s.listener)
		serverErr <- fmt.Errorf("RPC server failed: %w", err)
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
		log.Errorf("RPC server failed: %s", err)
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
	Provides() types.ProvidesCoin
	GetOngoingSwapState(types.Hash) common.SwapState
}

// ProtocolBackend represents protocol/backend.Backend
type ProtocolBackend interface {
	Env() common.Environment
	SetGasPrice(uint64)
	SetSwapTimeout(timeout time.Duration)
	SwapManager() swap.Manager
	SetEthAddress(ethcommon.Address)
	EthBalance() (ethcommon.Address, *big.Int, error)
	SetXMRDepositAddress(mcrypto.Address, types.Hash)
	ClearXMRDepositAddress(types.Hash)
}

// XMRTaker ...
type XMRTaker interface {
	Protocol
	InitiateProtocol(providesAmount float64, offer *types.Offer) (common.SwapState, error)
	Refund(types.Hash) (ethcommon.Hash, error)
	ExternalSender(offerID types.Hash) (*txsender.ExternalSender, error)
}

// XMRMaker ...
type XMRMaker interface {
	Protocol
	MakeOffer(offer *types.Offer, relayerEndpoint string, relayerCommission float64) (*types.OfferExtra, error)
	GetOffers() []*types.Offer
	ClearOffers([]string) error
	GetMoneroBalance() (string, *wallet.GetBalanceResponse, error)
}

// SwapManager ...
type SwapManager = swap.Manager
