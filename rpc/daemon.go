package rpc

import (
	"net/http"
	"runtime/debug"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/common"
)

// DaemonService handles general daemon operations
type DaemonService struct {
	server *Server
	cfg    *Config
}

// NewDaemonService ...
func NewDaemonService(server *Server, cfg *Config) *DaemonService {
	return &DaemonService{
		server,
		cfg,
	}
}

// ShutdownRequest ...
type ShutdownRequest struct{}

// ShutdownResponse ...
type ShutdownResponse struct{}

// Shutdown swapd
func (s *DaemonService) Shutdown(_ *http.Request, req *ShutdownRequest, resp *ShutdownResponse) error {
	return s.server.Stop()
}

// VersionRequest ...
type VersionRequest struct{}

// VersionResponse ...
type VersionResponse struct {
	SwapdVersion    string             `json:"swapdVersion" validate:"required"`
	P2PVersion      string             `json:"p2pVersion" validate:"required"`
	Env             common.Environment `json:"env" validate:"required"`
	SwapCreatorAddr ethcommon.Address  `json:"swapCreatorAddress" validate:"required"`
}

// Version returns version & misc info about swapd and its dependencies
func (s *DaemonService) Version(_ *http.Request, req *VersionRequest, resp *VersionResponse) error {
	resp.SwapdVersion = cliutil.GetVersion()

	resp.P2PVersion = func() string {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return "N/A (Can't read build info)"
		}

		for _, e := range info.Deps {
			if e.Path == "github.com/athanorlabs/go-p2p-net" {
				return e.Version
			}
		}
		return "N/A"
	}()

	resp.Env = s.cfg.ProtocolBackend.Env()
	resp.SwapCreatorAddr = s.cfg.ProtocolBackend.SwapCreatorAddr()
	return nil
}
