package rpc

import (
	"fmt"
	"net/http"

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
	s    *rpc.Server
	port uint16
}

// Config ...
type Config struct {
	Port        uint16
	Net         Net
	Alice       Alice
	Bob         Bob
	SwapManager SwapManager
}

// NewServer ...
func NewServer(cfg *Config) (*Server, error) {
	s := rpc.NewServer()
	s.RegisterCodec(NewCodec(), "application/json")
	if err := s.RegisterService(NewNetService(cfg.Net, cfg.Alice, cfg.Bob, cfg.SwapManager), "net"); err != nil { //nolint:lll
		return nil, err
	}

	if err := s.RegisterService(NewPersonalService(cfg.Bob), "personal"); err != nil {
		return nil, err
	}

	if err := s.RegisterService(NewSwapService(cfg.SwapManager, cfg.Alice), "swap"); err != nil {
		return nil, err
	}

	return &Server{
		s:    s,
		port: cfg.Port,
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
			log.Errorf("failed to start RPC server: %s", err)
			errCh <- err
		}
	}()

	return errCh
}

// Protocol represents the functions required by the rpc service into the protocol handler.
type Protocol interface {
	Provides() types.ProvidesCoin
	SetGasPrice(gasPrice uint64)
}

// Alice ...
type Alice interface {
	Protocol
	InitiateProtocol(providesAmount float64) (common.SwapState, error)
	Refund() (ethcommon.Hash, error)
}

// Bob ...
type Bob interface {
	Protocol
	MakeOffer(offer *types.Offer) error
	SetMoneroWalletFile(file, password string) error
}

// SwapManager ...
type SwapManager interface {
	GetPastIDs() []uint64
	GetPastSwap(id uint64) *swap.Info
	GetOngoingSwap() *swap.Info
}
