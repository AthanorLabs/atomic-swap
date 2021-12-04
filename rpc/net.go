package rpc

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/net"

	"github.com/libp2p/go-libp2p-core/peer"
)

const defaultSearchTime = time.Second * 12

type Net interface {
	Addresses() []string
	Discover(provides common.ProvidesCoin, searchTime time.Duration) ([]peer.AddrInfo, error)
	Query(who peer.AddrInfo) (*net.QueryResponse, error)
	Initiate(who peer.AddrInfo, msg *net.InitiateMessage, s net.SwapState) error
}

type Protocol interface {
	Provides() common.ProvidesCoin
	InitiateProtocol(providesAmount, desiredAmount float64) (net.SwapState, error)
}

type NetService struct {
	net      Net
	protocol Protocol
}

func NewNetService(net Net, protocol Protocol) *NetService {
	return &NetService{
		net:      net,
		protocol: protocol,
	}
}

type AddressesResponse struct {
	Addrs []string `json:"addresses"`
}

func (s *NetService) Addresses(_ *http.Request, _ *interface{}, resp *AddressesResponse) error {
	resp.Addrs = s.net.Addresses()
	return nil
}

type DiscoverRequest struct {
	Provides   common.ProvidesCoin `json:"provides"`
	SearchTime uint64              `json:"searchTime"` // in seconds
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

	peers, err := s.net.Discover(common.ProvidesCoin(req.Provides), searchTime)
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
	Provides      []common.ProvidesCoin `json:"provides"`
	MaximumAmount []float64             `json:"maximumAmount"`
	ExchangeRate  common.ExchangeRate   `json:"exchangeRate"`
}

func (s *NetService) QueryPeer(_ *http.Request, req *QueryPeerRequest, resp *QueryPeerResponse) error {
	who, err := net.StringToAddrInfo(req.Multiaddr)
	if err != nil {
		return err
	}

	msg, err := s.net.Query(who)
	if err != nil {
		return err
	}

	resp.Provides = msg.Provides
	resp.MaximumAmount = msg.MaximumAmount
	resp.ExchangeRate = msg.ExchangeRate
	return nil
}

type InitiateRequest struct {
	Multiaddr      string              `json:"multiaddr"`
	ProvidesCoin   common.ProvidesCoin `json:"provides"`
	ProvidesAmount float64             `json:"providesAmount"`
	DesiredAmount  float64             `json:"desiredAmount"`
}

type InitiateResponse struct {
	Success bool `json:"success"`
}

func (s *NetService) Initiate(_ *http.Request, req *InitiateRequest, resp *InitiateResponse) error {
	if req.ProvidesCoin == "" {
		return errors.New("must specify 'provides' coin")
	}

	swapState, err := s.protocol.InitiateProtocol(req.ProvidesAmount, req.DesiredAmount)
	if err != nil {
		return err
	}

	skm, err := swapState.SendKeysMessage()
	if err != nil {
		return err
	}

	msg := &net.InitiateMessage{
		Provides:        req.ProvidesCoin,
		ProvidesAmount:  req.ProvidesAmount,
		DesiredAmount:   req.DesiredAmount,
		SendKeysMessage: skm,
	}

	who, err := net.StringToAddrInfo(req.Multiaddr)
	if err != nil {
		return err
	}

	if err = s.net.Initiate(who, msg, swapState); err != nil {
		resp.Success = false
		return err
	}

	resp.Success = true
	return nil
}

type SetGasPriceRequest struct {
	GasPrice uint64
}

func (s *NetService) SetGasPrice(_ *http.Request, req *SetGasPriceRequest, _ *interface{}) error {
	return nil
}
