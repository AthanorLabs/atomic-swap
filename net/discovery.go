package net

import (
	"context"
	"fmt"
	"time"

	"github.com/athanorlabs/atomic-swap/common/types"

	libp2phost "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	libp2pdiscovery "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

const (
	tryAdvertiseTimeout = time.Second * 30
	defaultAdvertiseTTL = time.Minute * 5
	defaultMinPeers     = 3  // TODO: make this configurable
	defaultMaxPeers     = 50 // TODO: make this configurable
)

type discovery struct {
	ctx         context.Context
	dht         *dual.DHT
	h           libp2phost.Host
	rd          *libp2pdiscovery.RoutingDiscovery
	provides    []types.ProvidesCoin
	advertiseCh chan struct{}
	offerAPI    Handler
}

func newDiscovery(
	ctx context.Context,
	h libp2phost.Host,
	bnsFunc func() []peer.AddrInfo,
) (*discovery, error) {
	dhtOpts := []dual.Option{
		dual.DHTOption(kaddht.BootstrapPeersFunc(bnsFunc)),
		dual.DHTOption(kaddht.Mode(kaddht.ModeAutoServer)),
	}

	// There is libp2p bug when calling `dual.New` with a cancelled context creating a panic,
	// so we added the extra guard below:
	// Panic:  https://github.com/jbenet/goprocess/blob/v0.1.4/impl-mutex.go#L99
	// Caller: https://github.com/libp2p/go-libp2p-kad-dht/blob/v0.17.0/dht.go#L222
	//
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// not cancelled, continue on
	}

	dht, err := dual.New(ctx, h, dhtOpts...)
	if err != nil {
		return nil, err
	}

	rd := libp2pdiscovery.NewRoutingDiscovery(dht)

	return &discovery{
		ctx:         ctx,
		dht:         dht,
		h:           h,
		rd:          rd,
		advertiseCh: make(chan struct{}),
	}, nil
}

func (d *discovery) setOfferAPI(offerAPI Handler) {
	d.offerAPI = offerAPI
}

func (d *discovery) start() error {
	if d.offerAPI == nil {
		return errNilOfferAPI
	}

	err := d.dht.Bootstrap(d.ctx)
	if err != nil {
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	// wait to connect to bootstrap peers
	time.Sleep(time.Second)
	go d.advertiseLoop()
	go d.discoverLoop()

	log.Debug("discovery started!")
	return nil
}

func (d *discovery) stop() error {
	return d.dht.Close()
}

func (d *discovery) advertiseLoop() {
	ttl := d.advertise()

	for {
		select {
		case <-d.advertiseCh:
			d.provides = []types.ProvidesCoin{types.ProvidesXMR}
			ttl = d.advertise()
		case <-time.After(ttl):
			// the DHT clears provider records (ie. who is advertising what content)
			// every 24 hours.
			// so, if we don't have any offers available for 24 hours, then we are
			// no longer present in the DHT as a provider.
			// otherwise, we'll be present, but no offers will be sent when peers
			// query us.
			offers := d.offerAPI.GetOffers()
			if len(offers) == 0 {
				continue
			}

			ttl = d.advertise()
		case <-d.ctx.Done():
			return
		}
	}
}

// advertise advertises that we provide XMR in the DHT.
// note: we only advertise that we are an XMR provider, but we don't
// advertise our specific offers.
// to find what our offers are, peers need to send us a QueryRequest
// over the query subprotocol.
// the return value is the amount of time the caller should wait before
// trying to advertise again.
func (d *discovery) advertise() time.Duration {
	log.Debug("advertising in the DHT...")
	err := d.dht.Bootstrap(d.ctx)
	if err != nil {
		log.Warnf("failed to bootstrap DHT: err=%s", err)
		return tryAdvertiseTimeout
	}

	for _, provides := range d.provides {
		_, err = d.rd.Advertise(d.ctx, string(provides))
		if err != nil {
			log.Debugf("failed to advertise in the DHT: err=%s", err)
			return tryAdvertiseTimeout
		}
	}

	_, err = d.rd.Advertise(d.ctx, "")
	if err != nil {
		log.Debugf("failed to advertise in the DHT: err=%s", err)
		return tryAdvertiseTimeout
	}

	return defaultAdvertiseTTL
}

func (d *discovery) discoverLoop() {
	timer := time.NewTicker(time.Minute)

	for {
		select {
		case <-d.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			if len(d.h.Network().Peers()) >= defaultMinPeers {
				continue
			}

			// if our peer count is low, try to find some peers
			timer := time.NewTimer(time.Minute)

			_, err := d.findPeers("", timer.C)
			if err != nil {
				log.Errorf("failed to find peers: %s", err)
			}

			timer.Stop()
		}
	}
}

func (d *discovery) findPeers(provides string, done <-chan time.Time) ([]peer.AddrInfo, error) {
	peerCh, err := d.rd.FindPeers(d.ctx, provides)
	if err != nil {
		return nil, err
	}

	peers := []peer.AddrInfo{}

	for {
		select {
		case <-d.ctx.Done():
			return peers, d.ctx.Err()
		case <-done:
			return peers, nil
		case peer := <-peerCh:
			if peer.ID == d.h.ID() || peer.ID == "" {
				continue
			}

			log.Debugf("found new peer via DHT: peer=%s", peer.ID)
			peers = append(peers, peer)

			// found a peer, try to connect if we need more peers
			if len(d.h.Network().Peers()) < defaultMaxPeers {
				err = d.h.Connect(d.ctx, peer)
				if err != nil {
					log.Debugf("failed to connect to discovered peer %s: %s", peer.ID, err)
				}
			} else {
				d.h.Peerstore().AddAddrs(peer.ID, peer.Addrs, peerstore.PermanentAddrTTL)
			}
		}
	}
}

func (d *discovery) discover(
	provides types.ProvidesCoin,
	searchTime time.Duration,
) ([]peer.AddrInfo, error) {
	log.Debugf("attempting to find DHT peers that provide [%s] for %vs...",
		provides,
		searchTime.Seconds(),
	)

	timer := time.NewTimer(searchTime)
	defer timer.Stop()
	return d.findPeers(string(provides), timer.C)
}
