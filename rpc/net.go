package rpc

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/rpctypes"
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
	Initiate(who peer.AddrInfo, msg *net.SendKeysMessage, s common.SwapStateNet) error
	CloseProtocolStream(types.Hash)
}

// NetService is the RPC service prefixed by net_.
type NetService struct {
	net      Net
	xmrtaker XMRTaker
	xmrmaker XMRMaker
	sm       SwapManager
}

// NewNetService ...
func NewNetService(net Net, xmrtaker XMRTaker, xmrmaker XMRMaker, sm SwapManager) *NetService {
	return &NetService{
		net:      net,
		xmrtaker: xmrtaker,
		xmrmaker: xmrmaker,
		sm:       sm,
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

// Discover discovers peers over the network that provide a certain coin up for `SearchTime` duration of time.
func (s *NetService) Discover(_ *http.Request, req *rpctypes.DiscoverRequest, resp *rpctypes.DiscoverResponse) error {
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

// QueryPeer queries a peer for the coins they provide, their maximum amounts, and desired exchange rate.
func (s *NetService) QueryPeer(_ *http.Request, req *rpctypes.QueryPeerRequest,
	resp *rpctypes.QueryPeerResponse) error {
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

// TakeOffer initiates a swap with the given peer by taking an offer they've made.
func (s *NetService) TakeOffer(_ *http.Request, req *rpctypes.TakeOfferRequest,
	resp *rpctypes.TakeOfferResponse) error {
	_, infofile, err := s.takeOffer(req.Multiaddr, req.OfferID, req.ProvidesAmount)
	if err != nil {
		return err
	}

	resp.InfoFile = infofile
	return nil
}

func (s *NetService) takeOffer(multiaddr, offerID string,
	providesAmount float64) (<-chan types.Status, string, error) {
	who, err := net.StringToAddrInfo(multiaddr)
	if err != nil {
		return nil, "", err
	}

	queryResp, err := s.net.Query(who)
	if err != nil {
		return nil, "", err
	}

	var (
		found bool
		offer *types.Offer
	)
	for _, maybeOffer := range queryResp.Offers {
		if strings.Compare(maybeOffer.GetID().String(), offerID) == 0 {
			found = true
			offer = maybeOffer
			break
		}
	}

	if !found {
		return nil, "", errNoOfferWithID
	}

	swapState, err := s.xmrtaker.InitiateProtocol(providesAmount, offer)
	if err != nil {
		return nil, "", fmt.Errorf("failed to initiate protocol: %w", err)
	}

	skm, err := swapState.SendKeysMessage()
	if err != nil {
		return nil, "", err
	}

	skm.OfferID = offerID
	skm.ProvidedAmount = providesAmount

	if err = s.net.Initiate(who, skm, swapState); err != nil {
		_ = swapState.Exit()
		return nil, "", err
	}

	id, err := offerIDStringToHash(offerID)
	if err != nil {
		return nil, "", err
	}

	info := s.sm.GetOngoingSwap(id)
	if info == nil {
		return nil, "", errFailedToGetSwapInfo
	}

	return info.StatusCh(), swapState.InfoFile(), nil
}

// TakeOfferSyncResponse ...
type TakeOfferSyncResponse struct {
	InfoFile string `json:"infoFile"`
	Status   string `json:"status"`
}

// TakeOfferSync initiates a swap with the given peer by taking an offer they've made.
// It synchronously waits until the swap is completed before returning its status.
func (s *NetService) TakeOfferSync(_ *http.Request, req *rpctypes.TakeOfferRequest,
	resp *TakeOfferSyncResponse) error {
	offerID, err := offerIDStringToHash(req.OfferID)
	if err != nil {
		return err
	}

	_, infofile, err := s.takeOffer(req.Multiaddr, req.OfferID, req.ProvidesAmount)
	if err != nil {
		return err
	}

	resp.InfoFile = infofile

	const checkSwapSleepDuration = time.Millisecond * 100

	for {
		time.Sleep(checkSwapSleepDuration)

		info := s.sm.GetPastSwap(offerID)
		if info == nil {
			continue
		}

		resp.Status = info.Status().String()
		break
	}

	return nil
}

// MakeOffer creates and advertises a new swap offer.
func (s *NetService) MakeOffer(_ *http.Request, req *rpctypes.MakeOfferRequest,
	resp *rpctypes.MakeOfferResponse) error {
	id, extra, err := s.makeOffer(req)
	if err != nil {
		return err
	}

	resp.ID = id
	resp.InfoFile = extra.InfoFile
	s.net.Advertise()
	return nil
}

func (s *NetService) makeOffer(req *rpctypes.MakeOfferRequest) (string, *types.OfferExtra, error) {
	o := types.NewOffer(
		types.ProvidesXMR,
		req.MinimumAmount,
		req.MaximumAmount,
		req.ExchangeRate,
	)

	offerExtra, err := s.xmrmaker.MakeOffer(o)
	if err != nil {
		return "", nil, err
	}

	return o.GetID().String(), offerExtra, nil
}
