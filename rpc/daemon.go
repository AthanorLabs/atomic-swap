package rpc

import (
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/net"
)

// DaemonService handles RPC requests for swapd version, administration and (in the future) status requests.
type DaemonService struct {
	stopServer      func()
	env             common.Environment
	swapCreatorAddr *ethcommon.Address
}

// NewDaemonService creates a new daemon service. `swapCreatorAddr` is optional
// and not set by bootnodes.
func NewDaemonService(stopServer func(), env common.Environment, swapCreatorAddr *ethcommon.Address) *DaemonService {
	return &DaemonService{
		stopServer:      stopServer,
		env:             env,
		swapCreatorAddr: swapCreatorAddr,
	}
}

// Shutdown swapd
func (s *DaemonService) Shutdown(_ *http.Request, _ *any, _ *any) error {
	s.stopServer()
	return nil
}

// VersionResponse contains the version response provided by both swapd and
// bootnodes. In the case of bootnodes, the swapCreatorAddress is nil.
type VersionResponse struct {
	SwapdVersion    string             `json:"swapdVersion" validate:"required"`
	P2PVersion      string             `json:"p2pVersion" validate:"required"`
	Env             common.Environment `json:"env" validate:"required"`
	SwapCreatorAddr *ethcommon.Address `json:"swapCreatorAddress,omitempty"`
}

// Version returns version & misc info about swapd and its dependencies
func (s *DaemonService) Version(_ *http.Request, _ *any, resp *VersionResponse) error {
	resp.SwapdVersion = cliutil.GetVersion()
	resp.P2PVersion = fmt.Sprintf("%s/%d", net.ProtocolID, common.ChainIDFromEnv(s.env))
	resp.Env = s.env
	resp.SwapCreatorAddr = s.swapCreatorAddr
	return nil
}
