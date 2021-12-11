package net

import (
	"context"
	"fmt"
	"time"

	"github.com/noot/atomic-swap/common"

	libp2phost "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	libp2pdiscovery "github.com/libp2p/go-libp2p-discovery"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
)

const (
	initialAdvertisementTimeout = time.Millisecond
	tryAdvertiseTimeout         = time.Second * 30
)

type discovery struct {
	ctx      context.Context
	dht      *dual.DHT
	h        libp2phost.Host
	rd       *libp2pdiscovery.RoutingDiscovery
	provides []common.ProvidesCoin
}

func newDiscovery(ctx context.Context, h libp2phost.Host, bnsFunc func() []peer.AddrInfo,
	provides ...common.ProvidesCoin) (*discovery, error) {
	dhtOpts := []dual.Option{
		dual.DHTOption(kaddht.BootstrapPeersFunc(bnsFunc)),
		dual.DHTOption(kaddht.Mode(kaddht.ModeAutoServer)),
	}

	dht, err := dual.New(ctx, h, dhtOpts...)
	if err != nil {
		return nil, err
	}

	rd := libp2pdiscovery.NewRoutingDiscovery(dht)

	return &discovery{
		ctx:      ctx,
		dht:      dht,
		h:        h,
		rd:       rd,
		provides: provides,
	}, nil
}

func (d *discovery) start() error {
	err := d.dht.Bootstrap(d.ctx)
	if err != nil {
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	// wait to connect to bootstrap peers
	time.Sleep(time.Second)
	go d.advertise()

	log.Debug("discovery started!")
	return nil
}

func (d *discovery) stop() error {
	return d.dht.Close()
}

func (d *discovery) advertise() {
	ttl := initialAdvertisementTimeout

	for {
		select {
		case <-time.After(ttl):
			log.Debug("advertising in the DHT...")
			err := d.dht.Bootstrap(d.ctx)
			if err != nil {
				log.Warnf("failed to bootstrap DHT: err=%s", err)
				continue
			}

			for _, provides := range d.provides {
				ttl, err = d.rd.Advertise(d.ctx, string(provides))
				if err != nil {
					log.Debugf("failed to advertise in the DHT: err=%s", err)
					ttl = tryAdvertiseTimeout
				}
			}

			if len(d.provides) == 0 {
				ttl, err = d.rd.Advertise(d.ctx, "")
				if err != nil {
					log.Debugf("failed to advertise in the DHT: err=%s", err)
					ttl = tryAdvertiseTimeout
				}
			}
		case <-d.ctx.Done():
			return
		}
	}
}

func (d *discovery) discover(provides common.ProvidesCoin,
	searchTime time.Duration) ([]peer.AddrInfo, error) {
	log.Debugf("attempting to find DHT peers that provide %s for %s...", provides, searchTime)

	peerCh, err := d.rd.FindPeers(d.ctx, string(provides))
	if err != nil {
		return nil, err
	}

	timer := time.NewTicker(searchTime)
	peers := []peer.AddrInfo{}

	for {
		select {
		case <-d.ctx.Done():
			timer.Stop()
			return peers, d.ctx.Err()
		case <-timer.C:
			return peers, nil
		case peer := <-peerCh:
			if peer.ID == d.h.ID() || peer.ID == "" {
				continue
			}

			log.Debugf("found new peer via DHT: peer=%s", peer.ID)
			peers = append(peers, peer)

			// // found a peer, try to connect if we need more peers
			// if len(d.h.Network().Peers()) < d.maxPeers {
			// 	err = d.h.Connect(d.ctx, peer)
			// 	if err != nil {
			// 		logger.Trace("failed to connect to discovered peer", "peer", peer.ID, "err", err)
			// 	}
			// } else {
			d.h.Peerstore().AddAddrs(peer.ID, peer.Addrs, peerstore.PermanentAddrTTL)
			//return
			// }
		}
	}
}
