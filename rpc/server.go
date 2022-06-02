package rpc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/protocol/swap"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("rpc")

// Server represents the JSON-RPC server
type Server struct {
	s        *rpc.Server
	wsServer *wsServer
	port     uint16
	wsPort   uint16
}

// Config ...
type Config struct {
	Ctx             context.Context
	Port            uint16
	WsPort          uint16
	Net             Net
	XMRTaker        XMRTaker
	XMRMaker        XMRMaker
	ProtocolBackend ProtocolBackend
}

// NewServer ...
func NewServer(cfg *Config) (*Server, error) {
	s := rpc.NewServer()
	s.RegisterCodec(NewCodec(), "application/json")

	ns := NewNetService(cfg.Net, cfg.XMRTaker, cfg.XMRMaker, cfg.ProtocolBackend.SwapManager())
	if err := s.RegisterService(ns, "net"); err != nil {
		return nil, err
	}

	if err := s.RegisterService(NewPersonalService(cfg.XMRMaker, cfg.ProtocolBackend), "personal"); err != nil {
		return nil, err
	}

	if err := s.RegisterService(NewSwapService(cfg.ProtocolBackend.SwapManager(), cfg.XMRTaker, cfg.XMRMaker, cfg.Net), "swap"); err != nil {
		return nil, err
	}

	return &Server{
		s:        s,
		wsServer: newWsServer(cfg.Ctx, cfg.ProtocolBackend.SwapManager(), ns),
		port:     cfg.Port,
		wsPort:   cfg.WsPort,
	}, nil
}

// Start starts the JSON-RPC server.
func (s *Server) Start() <-chan error {
	errCh := make(chan error)

	go func() {
		r := mux.NewRouter()
		r.Handle("/", s.s)

		headersOk := handlers.AllowedHeaders([]string{"content-type", "username", "password"})
		methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
		originsOk := handlers.AllowedOrigins([]string{"*"})

		log.Infof("starting RPC server on http://localhost:%d", s.port)

		if err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), handlers.CORS(headersOk, methodsOk, originsOk)(r)); err != nil { //nolint:lll
			log.Errorf("failed to start http RPC server: %s", err)
			errCh <- err
		}
	}()

	go func() {
		r := mux.NewRouter()
		r.Handle("/", s.wsServer)

		headersOk := handlers.AllowedHeaders([]string{"content-type", "username", "password"})
		methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
		originsOk := handlers.AllowedOrigins([]string{"*"})

		log.Infof("starting websockets server on ws://localhost:%d", s.wsPort)

		if err := http.ListenAndServe(fmt.Sprintf(":%d", s.wsPort), handlers.CORS(headersOk, methodsOk, originsOk)(r)); err != nil { //nolint:lll
			log.Errorf("failed to start websockets RPC server: %s", err)
			errCh <- err
		}
	}()

	return errCh
}

// Protocol represents the functions required by the rpc service into the protocol handler.
type Protocol interface {
	Provides() types.ProvidesCoin
	GetOngoingSwapState() common.SwapState
}

type ProtocolBackend interface {
	SetGasPrice(uint64)
	SetSwapTimeout(timeout time.Duration)
	SwapManager() *swap.Manager
}

// XMRTaker ...
type XMRTaker interface {
	Protocol
	InitiateProtocol(providesAmount float64, offer *types.Offer) (common.SwapState, error)
	Refund() (ethcommon.Hash, error)
}

// XMRMaker ...
type XMRMaker interface {
	Protocol
	MakeOffer(offer *types.Offer) (*types.OfferExtra, error)
	SetMoneroWalletFile(file, password string) error
	GetOffers() []*types.Offer
	ClearOffers()
}

// SwapManager ...
type SwapManager interface {
	GetPastIDs() []uint64
	GetPastSwap(id uint64) *swap.Info
	GetOngoingSwap() *swap.Info
}
