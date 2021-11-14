package rpc

import (
	"net/http"
	"time"

	"github.com/noot/atomic-swap/net"

	"github.com/libp2p/go-libp2p-core/peer"
)

type Net interface {
	Discover(provides net.ProvidesCoin, searchTime time.Duration) ([]peer.AddrInfo, error)
}

type NetService struct {
	backend Net
}

func NewNetService(net Net) *NetService {
	return &NetService{
		backend: net,
	}
}

type DiscoverRequest struct {
	Provides   []string `json:"provides"`
	SearchTime uint64   `json:"searchTime",omitempty"`
}

type DiscoverResponse struct {
	Peers []string `json:"peers"`
}

// Discover discovers peers over the network that provide a certain coin up for `SearchTime` duration of time.
func (s *NetService) Discover(r *http.Request, req *DiscoverRequest, resp *DiscoverResponse) error {
	resp.Peers = []string{"noot"}
	return nil
}
