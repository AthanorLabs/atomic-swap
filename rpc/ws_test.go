package rpc

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/rpcclient/wsclient"

	"github.com/stretchr/testify/require"
)

const (
	testMultiaddr = "/ip4/192.168.0.102/tcp/9933/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2"
)

var (
	testSwapID  = types.Hash{99}
	testTimeout = time.Second * 5
)

func newServer(t *testing.T) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := &Config{
		Ctx:             ctx,
		Address:         "127.0.0.1:0", // OS assigned port
		Net:             new(mockNet),
		ProtocolBackend: newMockProtocolBackend(),
		XMRTaker:        new(mockXMRTaker),
		XMRMaker:        new(mockXMRMaker),
	}

	s, err := NewServer(cfg)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		err := s.Start()
		require.ErrorIs(t, err, http.ErrServerClosed)
		wg.Done()
	}()
	time.Sleep(time.Millisecond * 300) // let server start up

	t.Cleanup(func() {
		cancel()
		// Using non-cancelled context, so shutdown waits for clients to disconnect before unblocking
		err := s.httpServer.Shutdown(context.Background())
		require.NoError(t, err)
		wg.Wait() // unblocks when server exits
	})

	return s
}

func TestSubscribeSwapStatus(t *testing.T) {
	s := newServer(t)

	c, err := wsclient.NewWsClient(s.ctx, s.WsURL())
	require.NoError(t, err)

	ch, err := c.SubscribeSwapStatus(testSwapID)
	require.NoError(t, err)

	select {
	case status := <-ch:
		require.Equal(t, types.CompletedSuccess, status)
	case <-time.After(testTimeout):
		t.Fatal("test timed out")
	}
}

func TestSubscribeMakeOffer(t *testing.T) {
	s := newServer(t)

	c, err := wsclient.NewWsClient(s.ctx, s.WsURL())
	require.NoError(t, err)

	id, ch, err := c.MakeOfferAndSubscribe(0.1, 1, 0.05, types.EthAssetETH)
	require.NoError(t, err)
	require.NotEqual(t, id, testSwapID)
	select {
	case status := <-ch:
		require.Equal(t, types.CompletedSuccess, status)
	case <-time.After(testTimeout):
		t.Fatal("test timed out")
	}
}

func TestSubscribeTakeOffer(t *testing.T) {
	s := newServer(t)

	cliCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
	})
	c, err := wsclient.NewWsClient(cliCtx, s.WsURL())
	require.NoError(t, err)

	ch, err := c.TakeOfferAndSubscribe(testMultiaddr, testSwapID.String(), 1)
	require.NoError(t, err)

	select {
	case status := <-ch:
		require.Equal(t, types.CompletedSuccess, status)
	case <-time.After(testTimeout):
		t.Fatal("test timed out")
	}
}
