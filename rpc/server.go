package rpc

import (
	"fmt"
	"net/http"

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
	Port     uint16
	Net      Net
	Protocol Protocol
}

// NewServer ...
func NewServer(cfg *Config) (*Server, error) {
	s := rpc.NewServer()
	s.RegisterCodec(NewCodec(), "application/json")
	if err := s.RegisterService(NewNetService(cfg.Net, cfg.Protocol), "net"); err != nil {
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

		log.Infof("starting RPC server on http://localhost:%d", s.port)

		if err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), r); err != nil {
			log.Errorf("failed to start RPC server: %s", err)
			errCh <- err
		}
	}()

	return errCh
}
