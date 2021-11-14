package rpc

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("rpc")

type Server struct {
	s    *rpc.Server
	port uint32
}

type Config struct {
	Port uint32
	Net  Net
}

func NewServer(cfg *Config) *Server {
	s := rpc.NewServer()
	s.RegisterCodec(NewCodec(), "application/json")
	s.RegisterService(NewNetService(cfg.Net), "net")
	return &Server{
		s:    s,
		port: cfg.Port,
	}
}

func (s *Server) Start() {
	r := mux.NewRouter()
	r.Handle("/", s.s)

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), r)
		if err != nil {
			log.Errorf("failed to start RPC server: %s", err)
		}
	}()
}
