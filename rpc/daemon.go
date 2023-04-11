package rpc

import (
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/net"
)

// DaemonService handles RPC requests for swapd version, administration and (in the future) status requests.
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

// Shutdown swapd
func (s *DaemonService) Shutdown(_ *http.Request, _ *interface{}, _ *interface{}) error {
	return s.server.Stop()
}

// VersionResponse ...
type VersionResponse struct {
	SwapdVersion    string             `json:"swapdVersion" validate:"required"`
	P2PVersion      string             `json:"p2pVersion" validate:"required"`
	Env             common.Environment `json:"env" validate:"required"`
	SwapCreatorAddr ethcommon.Address  `json:"swapCreatorAddress" validate:"required"`
}

// Version returns version & misc info about swapd and its dependencies
func (s *DaemonService) Version(_ *http.Request, _ *interface{}, resp *VersionResponse) error {
	resp.SwapdVersion = cliutil.GetVersion()
	resp.P2PVersion = net.ProtocolID
	resp.Env = s.pb.Env()
	resp.SwapCreatorAddr = s.pb.SwapCreatorAddr()
	return nil
}
