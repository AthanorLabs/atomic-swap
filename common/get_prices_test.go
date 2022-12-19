package common

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func TestGetETHUSDPrice_mainnet(t *testing.T) {
	ec, err := ethclient.Dial(MainnetEndpoint)
	require.NoError(t, err)

	price, err := GetETHUSDPrice(context.Background(), ec)
	require.NoError(t, err)
	require.NotEqual(t, big.NewInt(0), price)
}

func TestGetETHUSDPrice_dev(t *testing.T) {
	ec, err := ethclient.Dial(DefaultEthEndpoint)
	require.NoError(t, err)

	price, err := GetETHUSDPrice(context.Background(), ec)
	require.NoError(t, err)
	require.NotEqual(t, big.NewInt(0), price)
}

func TestGetXMRUSDPrice_mainnet(t *testing.T) {
	ec, err := ethclient.Dial(MainnetEndpoint)
	require.NoError(t, err)

	price, err := GetXMRUSDPrice(context.Background(), ec)
	require.NoError(t, err)
	require.NotEqual(t, big.NewInt(0), price)
}

func TestGetXMRUSDPrice_dev(t *testing.T) {
	ec, err := ethclient.Dial(DefaultEthEndpoint)
	require.NoError(t, err)

	price, err := GetXMRUSDPrice(context.Background(), ec)
	require.NoError(t, err)
	require.NotEqual(t, big.NewInt(0), price)
}
