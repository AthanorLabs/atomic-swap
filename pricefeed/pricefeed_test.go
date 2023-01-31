package pricefeed

import (
	"context"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/tests"
)

func init() {
	logging.SetLogLevel("pricefeed", "debug")
}

func getMainnetEndpoint() string {
	endpoint := os.Getenv("ETH_MAINNET_ENDPOINT")
	if endpoint == "" {
		endpoint = mainnetEndpoint
	}
	return endpoint
}

func TestGetETHUSDPrice_mainnet(t *testing.T) {
	ec, err := ethclient.Dial(getMainnetEndpoint())
	require.NoError(t, err)
	defer ec.Close()

	feed, err := GetETHUSDPrice(context.Background(), ec)
	require.NoError(t, err)
	t.Logf("%s is $%s (updated: %s)", feed.Description, feed.Price, feed.UpdatedAt)
	assert.Equal(t, "ETH / USD", feed.Description)
	assert.False(t, feed.Price.Negative)
	assert.False(t, feed.Price.IsZero())
}

func TestGetETHUSDPrice_dev(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	feed, err := GetETHUSDPrice(context.Background(), ec)
	require.NoError(t, err)
	assert.Equal(t, "ETH / USD (fake)", feed.Description)
	assert.Equal(t, "1234.12345678", feed.Price.String())
}

func TestGetXMRUSDPrice_mainnet(t *testing.T) {
	ec, err := ethclient.Dial(getMainnetEndpoint())
	require.NoError(t, err)
	defer ec.Close()

	feed, err := GetXMRUSDPrice(context.Background(), ec)
	require.NoError(t, err)
	t.Logf("%s is $%s (updated: %s)", feed.Description, feed.Price, feed.UpdatedAt)
	assert.Equal(t, "XMR / USD", feed.Description)
	assert.False(t, feed.Price.Negative)
	assert.False(t, feed.Price.IsZero())
}

func TestGetXMRUSDPrice_dev(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	feed, err := GetXMRUSDPrice(context.Background(), ec)
	require.NoError(t, err)
	assert.Equal(t, "XMR / USD (fake)", feed.Description)
	assert.Equal(t, "123.12345678", feed.Price.String())
}
