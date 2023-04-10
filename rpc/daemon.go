package rpc

import (
	"net/http"
	"runtime/debug"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/common"
)

// DamonService handles general daemon operations
type DaemonService struct {
	server *Server
	cfg *Config
}

func NewDaemonService (server *Server, cfg *Config) *DaemonService {
	return &DaemonService {
		server,
		cfg,
	}
}

type ShutdownRequest struct {}
type ShutdownResponse struct {}

func (s *DaemonService) Shutdown(_ *http.Request, req *ShutdownRequest, resp *ShutdownResponse) error {
	return s.server.Stop()
}

type VersionRequest struct {}
type VersionResponse struct {
	SwapdVersion string `json:"swapd_version" validate:"required"`
	P2PVersion string `json:"p2p_version" validate:"required"`
	Env common.Environment `json:"env" validate:"required"`
	SwapCreatorAddr ethcommon.Address `json:"swap_creator_address" validate:"required"`
}

func (s *DaemonService) Version(_ *http.Request, req *VersionRequest, resp *VersionResponse) error {
	resp.SwapdVersion = cliutil.GetVersion()

	resp.P2PVersion = func () string {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return "N/A (Can't read build info)"
		}

		for _, e := range info.Deps {
			if(e.Path == "github.com/athanorlabs/go-p2p-net") {
				return e.Version
			}
		}
		return "N/A"
	} ()

	resp.Env = s.cfg.ProtocolBackend.Env()
	resp.SwapCreatorAddr = s.cfg.ProtocolBackend.SwapCreatorAddr()
	return nil
}
