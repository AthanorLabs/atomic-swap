package rpc

import (
	"testing"

	"github.com/noot/atomic-swap/common/rpctypes"
	"github.com/noot/atomic-swap/common/types"

	"github.com/stretchr/testify/require"
)

func TestNet_Discover(t *testing.T) {
	ns := NewNetService(new(mockNet), new(mockAlice), nil, new(mockSwapManager))

	req := &rpctypes.DiscoverRequest{
		Provides: "",
	}

	resp := new(rpctypes.DiscoverResponse)

	err := ns.Discover(nil, req, resp)
	require.NoError(t, err)
	require.Equal(t, 0, len(resp.Peers))
}

func TestNet_Query(t *testing.T) {
	ns := NewNetService(new(mockNet), new(mockAlice), nil, new(mockSwapManager))

	req := &rpctypes.QueryPeerRequest{
		Multiaddr: "/ip4/127.0.0.1/tcp/9900/p2p/12D3KooWDqCzbjexHEa8Rut7bzxHFpRMZyDRW1L6TGkL1KY24JH5",
	}

	resp := new(rpctypes.QueryPeerResponse)

	err := ns.QueryPeer(nil, req, resp)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Offers))
}

func TestNet_TakeOffer(t *testing.T) {
	ns := NewNetService(new(mockNet), new(mockAlice), nil, new(mockSwapManager))

	offer := &types.Offer{}

	req := &rpctypes.TakeOfferRequest{
		Multiaddr:      "/ip4/127.0.0.1/tcp/9900/p2p/12D3KooWDqCzbjexHEa8Rut7bzxHFpRMZyDRW1L6TGkL1KY24JH5",
		OfferID:        offer.GetID().String(),
		ProvidesAmount: 1,
	}

	resp := new(rpctypes.TakeOfferResponse)

	err := ns.TakeOffer(nil, req, resp)
	require.NoError(t, err)
	require.Equal(t, testSwapID, resp.ID)
}

func TestNet_TakeOfferSync(t *testing.T) {
	ns := NewNetService(new(mockNet), new(mockAlice), nil, new(mockSwapManager))

	offer := &types.Offer{}

	req := &rpctypes.TakeOfferRequest{
		Multiaddr:      "/ip4/127.0.0.1/tcp/9900/p2p/12D3KooWDqCzbjexHEa8Rut7bzxHFpRMZyDRW1L6TGkL1KY24JH5",
		OfferID:        offer.GetID().String(),
		ProvidesAmount: 1,
	}

	resp := new(TakeOfferSyncResponse)

	err := ns.TakeOfferSync(nil, req, resp)
	require.NoError(t, err)
	require.Equal(t, testSwapID, resp.ID)
}
