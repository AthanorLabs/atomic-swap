package rpc

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/net"

	"github.com/libp2p/go-libp2p-core/peer"
)

const defaultSearchTime = time.Second * 12

// Net contains the functions required by the rpc service into the network.
type Net interface {
	Addresses() []string
	Advertise()
	Discover(provides types.ProvidesCoin, searchTime time.Duration) ([]peer.AddrInfo, error)
	Query(who peer.AddrInfo) (*net.QueryResponse, error)
	Initiate(who peer.AddrInfo, msg *net.SendKeysMessage, s common.SwapState) error
}

// NetService is the RPC service prefixed by net_.
type NetService struct {
	net   Net
	alice Alice
	bob   Bob
	sm    SwapManager
}

// NewNetService ...
func NewNetService(net Net, alice Alice, bob Bob, sm SwapManager) *NetService {
	return &NetService{
		net:   net,
		alice: alice,
		bob:   bob,
		sm:    sm,
	}
}

// AddressesResponse ...
type AddressesResponse struct {
	Addrs []string `json:"addresses"`
}

// Addresses returns the multiaddresses this node is listening on.
func (s *NetService) Addresses(_ *http.Request, _ *interface{}, resp *AddressesResponse) error {
	resp.Addrs = s.net.Addresses()
	return nil
}

// DiscoverRequest ...
type DiscoverRequest struct {
	Provides   types.ProvidesCoin `json:"provides"`
	SearchTime uint64             `json:"searchTime"` // in seconds
}

// DiscoverResponse ...
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

	peers, err := s.net.Discover(req.Provides, searchTime)
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

// QueryPeerRequest ...
type QueryPeerRequest struct {
	// Multiaddr of peer to query
	Multiaddr string `json:"multiaddr"`
}

// QueryPeerResponse ...
type QueryPeerResponse struct {
	Offers []*types.Offer `json:"offers"`
}

// QueryPeer queries a peer for the coins they provide, their maximum amounts, and desired exchange rate.
func (s *NetService) QueryPeer(_ *http.Request, req *QueryPeerRequest, resp *QueryPeerResponse) error {
	who, err := net.StringToAddrInfo(req.Multiaddr)
	if err != nil {
		return err
	}

	msg, err := s.net.Query(who)
	if err != nil {
		return err
	}

	resp.Offers = msg.Offers
	return nil
}

// TakeOfferRequest ...
type TakeOfferRequest struct {
	Multiaddr      string  `json:"multiaddr"`
	OfferID        string  `json:"offerID"`
	ProvidesAmount float64 `json:"providesAmount"`
}

// TakeOfferResponse ...
type TakeOfferResponse struct {
	ID uint64 `json:"id"`
}

// TakeOffer initiates a swap with the given peer by taking an offer they've made.
func (s *NetService) TakeOffer(_ *http.Request, req *TakeOfferRequest, resp *TakeOfferResponse) error {
	id, _, err := s.takeOffer(req.Multiaddr, req.OfferID, req.ProvidesAmount)
	if err != nil {
		return err
	}

	resp.ID = id
	return nil
}

func (s *NetService) takeOffer(multiaddr, offerID string,
	providesAmount float64) (uint64, <-chan types.Status, error) {
	swapState, err := s.alice.InitiateProtocol(providesAmount)
	if err != nil {
		return 0, nil, err
	}

	skm, err := swapState.SendKeysMessage()
	if err != nil {
		return 0, nil, err
	}

	skm.OfferID = offerID
	skm.ProvidedAmount = providesAmount

	who, err := net.StringToAddrInfo(multiaddr)
	if err != nil {
		return 0, nil, err
	}

	if err = s.net.Initiate(who, skm, swapState); err != nil {
		_ = swapState.ProtocolExited()
		return 0, nil, err
	}

	info := s.sm.GetOngoingSwap()
	if info == nil {
		return 0, nil, errors.New("failed to get swap info after initiating")
	}

	return swapState.ID(), info.StatusCh(), nil
}

// TakeOfferSyncResponse ...
type TakeOfferSyncResponse struct {
	ID     uint64 `json:"id"`
	Status string `json:"status"`
}

// TakeOfferSync initiates a swap with the given peer by taking an offer they've made.
// It synchronously waits until the swap is completed before returning its status.
func (s *NetService) TakeOfferSync(_ *http.Request, req *TakeOfferRequest,
	resp *TakeOfferSyncResponse) error {
	swapState, err := s.alice.InitiateProtocol(req.ProvidesAmount)
	if err != nil {
		return err
	}

	skm, err := swapState.SendKeysMessage()
	if err != nil {
		return err
	}

	skm.OfferID = req.OfferID
	skm.ProvidedAmount = req.ProvidesAmount

	who, err := net.StringToAddrInfo(req.Multiaddr)
	if err != nil {
		return err
	}

	if err = s.net.Initiate(who, skm, swapState); err != nil {
		_ = swapState.ProtocolExited()
		return err
	}

	resp.ID = swapState.ID()

	const checkSwapSleepDuration = time.Millisecond * 100

	for {
		time.Sleep(checkSwapSleepDuration)

		info := s.sm.GetPastSwap(resp.ID)
		if info == nil {
			continue
		}

		resp.Status = info.Status().String()
		break
	}

	return nil
}

// MakeOfferRequest ...
type MakeOfferRequest struct {
	MinimumAmount float64            `json:"minimumAmount"`
	MaximumAmount float64            `json:"maximumAmount"`
	ExchangeRate  types.ExchangeRate `json:"exchangeRate"`
}

// MakeOfferResponse ...
type MakeOfferResponse struct {
	ID string `json:"offerID"`
}

// MakeOffer creates and advertises a new swap offer.
func (s *NetService) MakeOffer(_ *http.Request, req *MakeOfferRequest, resp *MakeOfferResponse) error {
	id, _, err := s.makeOffer(req)
	if err != nil {
		return err
	}

	resp.ID = id
	s.net.Advertise()
	return nil
}

func (s *NetService) makeOffer(req *MakeOfferRequest) (string, *types.OfferExtra, error) {
	o := &types.Offer{
		Provides:      types.ProvidesXMR,
		MinimumAmount: req.MinimumAmount,
		MaximumAmount: req.MaximumAmount,
		ExchangeRate:  req.ExchangeRate,
	}

	offerExtra, err := s.bob.MakeOffer(o)
	if err != nil {
		return "", nil, err
	}

	return o.GetID().String(), offerExtra, nil
}

// SetGasPriceRequest ...
type SetGasPriceRequest struct {
	GasPrice uint64
}

// SetGasPrice sets the gas price (in wei) to be used for ethereum transactions.
func (s *NetService) SetGasPrice(_ *http.Request, req *SetGasPriceRequest, _ *interface{}) error {
	s.alice.SetGasPrice(req.GasPrice)
	s.bob.SetGasPrice(req.GasPrice)
	return nil
}
