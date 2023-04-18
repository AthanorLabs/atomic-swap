// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package rpc

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/rpcclient/wsclient"
)

var (
	testPeerID, _ = peer.Decode("12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2")
	testSwapID    = types.Hash{99}
	testTimeout   = time.Second * 5
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

	min := coins.StrToDecimal("0.1")
	max := coins.StrToDecimal("1")
	exRate := coins.ToExchangeRate(coins.StrToDecimal("0.05"))
	offerResp, ch, err := c.MakeOfferAndSubscribe(min, max, exRate, types.EthAssetETH, false)
	require.NoError(t, err)
	require.NotEqual(t, offerResp.OfferID, testSwapID)
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

	ch, err := c.TakeOfferAndSubscribe(testPeerID, testSwapID, apd.New(1, 0))
	require.NoError(t, err)

	select {
	case status := <-ch:
		require.Equal(t, types.CompletedSuccess, status)
	case <-time.After(testTimeout):
		t.Fatal("test timed out")
	}
}
