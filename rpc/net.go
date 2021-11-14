package rpc

import (
	"fmt"
	"net/http"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/net"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

const defaultSearchTime = time.Second * 12

type Net interface {
	Discover(provides net.ProvidesCoin, searchTime time.Duration) ([]peer.AddrInfo, error)
	Query(who peer.AddrInfo) (*net.QueryResponse, error)
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
	Provides   net.ProvidesCoin `json:"provides"`
	SearchTime uint64           `json:"searchTime",omitempty"` // in seconds
}

type DiscoverResponse struct {
	Peers [][]string `json:"peers"`
}

// Discover discovers peers over the network that provide a certain coin up for `SearchTime` duration of time.
func (s *NetService) Discover(_ *http.Request, req *DiscoverRequest, resp *DiscoverResponse) error {
	searchTime, err := time.ParseDuration(fmt.Sprintf("%ds", req.SearchTime))
	if err != nil {
		return err
	}

	if searchTime == 0 {
		searchTime = defaultSearchTime
	}

	peers, err := s.backend.Discover(net.ProvidesCoin(req.Provides), searchTime)
	if err != nil {
		return err
	}

	resp.Peers = make([][]string, len(peers))
	for i, p := range peers {
		resp.Peers[i] = addrInfoToStrings(p)
	}

	return nil
}

func addrInfoToStrings(addrInfo peer.AddrInfo) []string {
	strs := make([]string, len(addrInfo.Addrs))
	for i, addr := range addrInfo.Addrs {
		strs[i] = fmt.Sprintf("%s/p2p/%s", addr, addrInfo.ID)
	}
	return strs
}

func stringToAddrInfo(s string) (peer.AddrInfo, error) {
	maddr, err := multiaddr.NewMultiaddr(s)
	if err != nil {
		return peer.AddrInfo{}, err
	}
	p, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return peer.AddrInfo{}, err
	}
	return *p, err
}

type QueryPeerRequest struct {
	// Multiaddr of peer to query
	Multiaddr string `json:"multiaddr"`
}

type QueryPeerResponse struct {
	Provides      []net.ProvidesCoin  `json:"provides"`
	MaximumAmount []uint64            `json:"maximumAmount"`
	ExchangeRate  common.ExchangeRate `json:"exchangeRate"`
}

func (s *NetService) QueryPeer(_ *http.Request, req *QueryPeerRequest, resp *QueryPeerResponse) error {
	who, err := stringToAddrInfo(req.Multiaddr)
	if err != nil {
		return err
	}

	msg, err := s.backend.Query(who)
	if err != nil {
		return err
	}

	resp.Provides = msg.Provides
	resp.MaximumAmount = msg.MaximumAmount
	resp.ExchangeRate = msg.ExchangeRate
	return nil
}
