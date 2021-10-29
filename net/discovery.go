package net

import (
	"context"

	libp2phost "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
)

type discovery struct {
	ctx context.Context
	dht *dual.DHT
}

func newDiscovery(ctx context.Context, h libp2phost.Host, bnsFunc func() []peer.AddrInfo) (*discovery, error) {
	dhtOpts := []dual.Option{
		dual.DHTOption(kaddht.BootstrapPeersFunc(bnsFunc)),
		dual.DHTOption(kaddht.Mode(kaddht.ModeAutoServer)),
	}

	dht, err := dual.New(ctx, h, dhtOpts...)
	if err != nil {
		return nil, err
	}

	return &discovery{
		ctx: ctx,
		dht: dht,
	}, nil
}
