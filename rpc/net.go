package rpc

import (
	"fmt"
	"net/http"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/net"

	"github.com/libp2p/go-libp2p-core/peer"
)

const defaultSearchTime = time.Second * 12

type Net interface {
	Discover(provides net.ProvidesCoin, searchTime time.Duration) ([]peer.AddrInfo, error)
	Query(who peer.AddrInfo) (*net.QueryResponse, error)
	Initiate(who peer.ID, msg *net.InitiateMessage) error
}

type Protocol interface {
	Provides() net.ProvidesCoin
	InitiateProtocol(providesAmount, desiredAmount uint64) error
	SendKeysMessage() (*net.SendKeysMessage, error)
}

type NetService struct {
	backend  Net
	protocol Protocol
}

func NewNetService(net Net, protocol Protocol) *NetService {
	return &NetService{
		backend:  net,
		protocol: protocol,
	}
}

type DiscoverRequest struct {
	Provides   net.ProvidesCoin `json:"provides"`
	SearchTime uint64           `json:"searchTime"` // in seconds
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
	who, err := net.StringToAddrInfo(req.Multiaddr)
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

type InitiateRequest struct {
	PeerID         string
	ProvidesCoin   net.ProvidesCoin
	ProvidesAmount uint64
	DesiredAmount  uint64
}

type InitiateResponse struct {
	Success bool
}

func (s *NetService) Initiate(_ *http.Request, req *InitiateRequest, resp *InitiateResponse) error {
	skm, err := s.protocol.SendKeysMessage()
	if err != nil {
		return err
	}

	msg := &net.InitiateMessage{
		Provides:        req.ProvidesCoin,
		ProvidesAmount:  req.ProvidesAmount,
		DesiredAmount:   req.DesiredAmount,
		SendKeysMessage: skm,
	}

	return s.backend.Initiate(peer.ID(req.PeerID), msg)
}
