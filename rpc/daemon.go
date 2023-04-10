package rpc

import (
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/net"
)

// DaemonService handles general daemon operations
type DaemonService struct {
	server *Server
	pb     ProtocolBackend
}

// NewDaemonService ...
func NewDaemonService(server *Server, pb ProtocolBackend) *DaemonService {
	return &DaemonService{
		server,
		pb,
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
	resp.P2PVersion = net.ProtocolID
	resp.Env = s.pb.Env()
	resp.SwapCreatorAddr = s.pb.SwapCreatorAddr()
	return nil
}
