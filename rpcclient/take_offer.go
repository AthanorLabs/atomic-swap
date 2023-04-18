// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package rpcclient

import (
	"github.com/cockroachdb/apd/v3"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/common/types"
)

// TakeOffer calls net_takeOffer.
func (c *Client) TakeOffer(peerID peer.ID, offerID types.Hash, providesAmount *apd.Decimal) error {
	const (
		method = "net_takeOffer"
	)

	req := &rpctypes.TakeOfferRequest{
		PeerID:         peerID,
		OfferID:        offerID,
		ProvidesAmount: providesAmount,
	}

	if err := c.Post(method, req, nil); err != nil {
		return err
	}

	return nil
}
