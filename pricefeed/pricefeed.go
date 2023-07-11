// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package pricefeed implements routines to retrieve on-chain price feeds from chainlink's
// decentralized oracle network.
package pricefeed

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log/v2"

	"github.com/athanorlabs/atomic-swap/common"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
)

const (
	// fallbackOptimismEndpoint is an RPC endpoint for optimism mainnet. Note that we
	// tried https://mainnet.optimism.io first, but it is severely rate limited
	// to around 2 requests/second.
	fallbackOptimismEndpoint = "https://optimism.blockpi.network/v1/rpc/public"

	// https://data.chain.link/optimism/mainnet/crypto-usd/eth-usd
	chainlinkETHToUSDProxy = "0x13e3ee699d1909e989722e753853ae30b17e08c5"

	// https://data.chain.link/optimism/mainnet/crypto-usd/xmr-usd
	chainlinkXMRToUSDProxy = "0x2a8d91686a048e98e6ccf1a89e82f40d14312672"
)

var (
	errUnsupportedNetwork = errors.New("unsupported network")
	log                   = logging.Logger("pricefeed")
)

// PriceFeed contains the interesting data from a chainlink price feed query.
type PriceFeed struct {
	Description string // "COIN / USD"
	Price       *apd.Decimal
	UpdatedAt   time.Time
}

func getOptimismEndpoint() string {
	endpoint := os.Getenv("ETH_OPTIMISM_ENDPOINT")
	if endpoint == "" {
		endpoint = fallbackOptimismEndpoint
	}
	return endpoint
}

// GetETHUSDPrice returns the current ETH/USD price from the Chainlink oracle.
func GetETHUSDPrice(ctx context.Context, ec *ethclient.Client) (*PriceFeed, error) {
	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	switch chainID.Uint64() {
	case common.OpMainnetChainID:
		// No extra work to do
	case common.MainnetChainID, common.SepoliaChainID:
		ec, err = ethclient.Dial(getOptimismEndpoint())
		if err != nil {
			return nil, err
		}
		defer ec.Close()
	case common.GanacheChainID, common.HardhatChainID:
		return &PriceFeed{
			Description: "ETH / USD (fake)",
			Price:       apd.New(123412345678, -8), // 1234.12345678
			UpdatedAt:   time.Now(),
		}, nil
	default:
		return nil, errUnsupportedNetwork
	}

	return getChainlinkPriceFeed(ctx, chainlinkETHToUSDProxy, ec)
}

// GetXMRUSDPrice returns the current XMR/USD price from the Chainlink oracle.
func GetXMRUSDPrice(ctx context.Context, ec *ethclient.Client) (*PriceFeed, error) {
	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	switch chainID.Uint64() {
	case common.OpMainnetChainID:
		// No extra work to do
	case common.MainnetChainID, common.SepoliaChainID:
		// Push stagenet/sepolia users to a mainnet endpoint
		ec, err = ethclient.Dial(getOptimismEndpoint())
		if err != nil {
			return nil, err
		}
		defer ec.Close()
	case common.GanacheChainID, common.HardhatChainID:
		return &PriceFeed{
			Description: "XMR / USD (fake)",
			Price:       apd.New(12312345678, -8), // 123.12345678
			UpdatedAt:   time.Now(),
		}, nil
	default:
		return nil, errUnsupportedNetwork
	}

	return getChainlinkPriceFeed(ctx, chainlinkXMRToUSDProxy, ec)
}

// getChainlinkPriceFeed retries the latest price feed data from the given contract address.
func getChainlinkPriceFeed(ctx context.Context, feedAddress string, ec *ethclient.Client) (*PriceFeed, error) {
	chainlinkPriceFeedProxy, err := contracts.NewAggregatorV3Interface(ethcommon.HexToAddress(feedAddress), ec)
	if err != nil {
		return nil, err
	}

	opts := &bind.CallOpts{
		Context: ctx,
	}

	roundData, err := chainlinkPriceFeedProxy.LatestRoundData(opts)
	if err != nil {
		return nil, err
	}

	decimals, err := chainlinkPriceFeedProxy.Decimals(opts)
	if err != nil {
		return nil, err
	}

	price := apd.NewWithBigInt(new(apd.BigInt).SetMathBigInt(roundData.Answer), -int32(decimals))
	_, _ = price.Reduce(price) // push even multiples of 10 to the exponent
	updatedAt := time.Unix(roundData.UpdatedAt.Int64(), 0)

	description, err := chainlinkPriceFeedProxy.Description(opts)
	if err != nil {
		return nil, err
	}

	log.Debugf("%s: $%s (%s)", description, price, updatedAt)
	return &PriceFeed{
		Description: description,
		Price:       price,
		UpdatedAt:   updatedAt,
	}, nil
}
