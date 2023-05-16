// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package rpcclient

import (
	"testing"

	"github.com/cockroachdb/apd/v3"

	"github.com/athanorlabs/atomic-swap/common/rpctypes"
	"github.com/athanorlabs/atomic-swap/rpc"

	"github.com/stretchr/testify/require"
)

func TestNet_Discover(t *testing.T) {
	ns := rpc.NewNetService(new(mockNet), new(mockXMRTaker), nil, mockSwapManager(t), false)

	req := &rpctypes.DiscoverRequest{
		Provides: "",
	}

	resp := new(rpctypes.DiscoverResponse)

	err := ns.Discover(nil, req, resp)
	require.NoError(t, err)
	require.Equal(t, 0, len(resp.PeerIDs))
}

func TestNet_Query(t *testing.T) {
	ns := rpc.NewNetService(new(mockNet), new(mockXMRTaker), nil, mockSwapManager(t), false)

	req := &rpctypes.QueryPeerRequest{
		PeerID: "12D3KooWDqCzbjexHEa8Rut7bzxHFpRMZyDRW1L6TGkL1KY24JH5",
	}

	resp := new(rpctypes.QueryPeerResponse)

	err := ns.QueryPeer(nil, req, resp)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Offers))
}

func TestNet_TakeOffer(t *testing.T) {
	ns := rpc.NewNetService(new(mockNet), new(mockXMRTaker), nil, mockSwapManager(t), false)

	req := &rpctypes.TakeOfferRequest{
		PeerID:         "12D3KooWDqCzbjexHEa8Rut7bzxHFpRMZyDRW1L6TGkL1KY24JH5",
		OfferID:        testSwapID,
		ProvidesAmount: apd.New(1, 0),
	}

	err := ns.TakeOffer(nil, req, nil)
	require.NoError(t, err)
}

func TestNet_TakeOfferSync(t *testing.T) {
	ns := rpc.NewNetService(new(mockNet), new(mockXMRTaker), nil, mockSwapManager(t), false)

	req := &rpctypes.TakeOfferRequest{
		PeerID:         "12D3KooWDqCzbjexHEa8Rut7bzxHFpRMZyDRW1L6TGkL1KY24JH5",
		OfferID:        testSwapID,
		ProvidesAmount: apd.New(1, 0),
	}

	resp := new(rpc.TakeOfferSyncResponse)

	err := ns.TakeOfferSync(nil, req, resp)
	require.NoError(t, err)
}
