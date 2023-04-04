// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package rpcclient

import (
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/rpctypes"
)

// Discover calls net_discover.
func (c *Client) Discover(provides string, searchTime uint64) ([]peer.ID, error) {
	const (
		method = "net_discover"
	)

	req := &rpctypes.DiscoverRequest{
		Provides:   provides,
		SearchTime: searchTime,
	}
	res := &rpctypes.DiscoverResponse{}

	if err := c.Post(method, req, res); err != nil {
		return nil, err
	}

	return res.PeerIDs, nil
}

// QueryAll calls net_queryAll.
func (c *Client) QueryAll(provides coins.ProvidesCoin, searchTime uint64) ([]*rpctypes.PeerWithOffers, error) {
	const (
		method = "net_queryAll"
	)

	req := &rpctypes.QueryAllRequest{
		Provides:   string(provides),
		SearchTime: searchTime,
	}
	res := &rpctypes.QueryAllResponse{}

	if err := c.Post(method, req, res); err != nil {
		return nil, err
	}

	return res.PeersWithOffers, nil
}
