// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package pricefeed

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/tests"
)

func init() {
	logging.SetLogLevel("pricefeed", "debug")
}

func newOptimismClient(t *testing.T) *ethclient.Client {
	ec, err := ethclient.Dial(optimismEndpoint)
	require.NoError(t, err)
	t.Cleanup(func() {
		ec.Close()
	})

	return ec
}

func TestGetETHUSDPrice_nonDev(t *testing.T) {
	for i, ec := range []*ethclient.Client{tests.NewEthSepoliaClient(t), newOptimismClient(t)} {
		if i > 0 {
			// Make sure we don't exceed the per second rate limit of the free endpoint
			time.Sleep(time.Second)
		}
		feed, err := GetETHUSDPrice(context.Background(), ec)
		require.NoError(t, err)
		t.Logf("%s is $%s (updated: %s)", feed.Description, feed.Price, feed.UpdatedAt)
		assert.Equal(t, "ETH / USD", feed.Description)
		assert.False(t, feed.Price.Negative)
		assert.False(t, feed.Price.IsZero())
	}
}

func TestGetETHUSDPrice_dev(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	feed, err := GetETHUSDPrice(context.Background(), ec)
	require.NoError(t, err)
	assert.Equal(t, "ETH / USD (fake)", feed.Description)
	assert.Equal(t, "1234.12345678", feed.Price.String())
}

func TestGetXMRUSDPrice_nonDev(t *testing.T) {
	for i, ec := range []*ethclient.Client{tests.NewEthSepoliaClient(t), newOptimismClient(t)} {
		if i > 0 {
			// Make sure we don't exceed the per second rate limit of the free endpoint
			time.Sleep(time.Second)
		}
		feed, err := GetXMRUSDPrice(context.Background(), ec)
		require.NoError(t, err)
		t.Logf("%s is $%s (updated: %s)", feed.Description, feed.Price, feed.UpdatedAt)
		assert.Equal(t, "XMR / USD", feed.Description)
		assert.False(t, feed.Price.Negative)
		assert.False(t, feed.Price.IsZero())
	}
}

func TestGetXMRUSDPrice_dev(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	feed, err := GetXMRUSDPrice(context.Background(), ec)
	require.NoError(t, err)
	assert.Equal(t, "XMR / USD (fake)", feed.Description)
	assert.Equal(t, "123.12345678", feed.Price.String())
}
