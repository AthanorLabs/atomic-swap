// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package rpcclient

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/rpc"
)

var (
	testPeerID, _ = peer.Decode("12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2")
	testSwapID    = types.Hash{99}
	testTimeout   = time.Second * 5
)

func newServer(t *testing.T) (*rpc.Server, *rpc.Config) {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := &rpc.Config{
		Ctx:             ctx,
		Env:             common.Development,
		Address:         "127.0.0.1:0", // OS assigned port
		Net:             new(mockNet),
		ProtocolBackend: newMockProtocolBackend(t),
		XMRTaker:        new(mockXMRTaker),
		XMRMaker:        new(mockXMRMaker),
		Namespaces:      rpc.AllNamespaces(),
	}

	s, err := rpc.NewServer(cfg)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		err := s.Start()
		require.ErrorIs(t, err, context.Canceled)
		wg.Done()
	}()
	time.Sleep(time.Millisecond * 300) // let server start up

	t.Cleanup(func() {
		// ctx is local to this function, but we don't want to shut down the server
		// by canceling it until the end of the test.
		cancel()
		wg.Wait() // wait for the server to exit
	})

	return s, cfg
}

func TestSubscribeSwapStatus(t *testing.T) {
	ctx := context.Background()
	s, _ := newServer(t)

	c, err := NewWsClient(ctx, s.Port())
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
	ctx := context.Background()
	s, cfg := newServer(t)

	c, err := NewWsClient(ctx, s.Port())
	require.NoError(t, err)

	min := coins.StrToDecimal("0.1")
	max := coins.StrToDecimal("1")
	exRate := coins.ToExchangeRate(coins.StrToDecimal("0.05"))
	offerResp, ch, err := c.MakeOfferAndSubscribe(min, max, exRate, types.EthAssetETH, false)
	require.NoError(t, err)
	require.NotEqual(t, offerResp.OfferID, testSwapID)

	cfg.ProtocolBackend.SwapManager().PushNewStatus(offerResp.OfferID, types.CompletedSuccess)

	select {
	case status := <-ch:
		require.Equal(t, types.CompletedSuccess, status)
	case <-time.After(testTimeout):
		t.Fatal("test timed out")
	}
}

func TestSubscribeTakeOffer(t *testing.T) {
	s, _ := newServer(t)

	cliCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
	})
	c, err := NewWsClient(cliCtx, s.Port())
	require.NoError(t, err)

	ch, err := c.TakeOfferAndSubscribe(testPeerID, testSwapID, apd.New(1, 0))
	require.NoError(t, err)

	select {
	case status := <-ch:
		require.Equal(t, types.CompletedSuccess, status)
	case <-time.After(testTimeout):
		t.Fatal("test timed out")
	}
}
